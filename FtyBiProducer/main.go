// 啟動服務與排程
package main

import (
	"FtyBiProducer/config"
	mq "FtyBiProducer/mq"
	scilog "FtyBiProducer/scilog"
	"FtyBiProducer/service"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"FtyBiProducer/metrics"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {

	// 1. 建立 Logger
	logger, err := scilog.NewFileLogger()
	if err != nil {
		// 如果 Logger 建立失敗，立即中止程式並印出錯誤
		panic(fmt.Sprintf("無法建立日誌：%v", err))
	}
	// 確保程式結束前會把緩衝區 flush
	defer logger.Sync()

	// 2. 建立 Sugared Logger，方便後續呼叫
	sugar := logger.Sugar()
	sugar.Info("Logger 初始化成功")

	// 3. 建立可取消的 Context，訂閱 SIGINT 和 SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 4. 各種設定管理
       cfg, err := config.LoadConfig(config.ConfigFilePath)
	if err != nil {
		sugar.Fatalf("載入設定失敗：：", zap.Error(err))
	}

	// 5. 啟動 Metrics Server. 先啟動一個專門用來暴露 /metrics 的 HTTP 伺服器
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		sugar.Info("Metrics 伺服器啟動，監聽 :" + strconv.Itoa(cfg.Prometheus.MetricsPort) + "/metrics")
		if err := http.ListenAndServe(":"+strconv.Itoa(cfg.Prometheus.MetricsPort), nil); err != nil {
			sugar.Fatalf("Metrics server 錯誤", zap.Error(err))
		}
	}()

	// 6. 建立 MQ Client
	mqClient, err := mq.NewMQClient(cfg.MQ)
	if err != nil {
		sugar.Fatalf("MQ 初始化失敗：", zap.Error(err))
	}
	defer mqClient.Close()

	// 7. 初始化資料庫
	db, err := config.InitGormDB(cfg.DB)
	if err != nil {
		sugar.Fatalf("初始化資料庫失敗：", zap.Error(err))
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	// 8. 建立 Processor
	proc := service.New(db, mqClient)

	// 9. 用 WaitGroup 等待兩條線程結束
	var wg sync.WaitGroup
	wg.Add(2)

	// 10-1. Producer 1 號 處理DDL: DdlLogProcess
	go func() {
		defer wg.Done()

		// 10-1-1. 建立 Ticker，依 ProcessDdlInterval 週期執行
		// ticker := time.NewTicker(cfg.ProcessDdlInterval)
		// defer ticker.Stop()

		sugar.Info("DDL 服務啟動，每" + cfg.ProcessDdlInterval.String() + "秒執行一次批次處理")

		// 10-1-2. 主迴圈: 呼叫 Processor 處理（序列化 + 發送），並收集 metrics，並優雅關閉
		for {
			select {
			case <-ctx.Done():
				sugar.Info("DDL服務收到關機訊號，開始清理...")
				// TODO: 若有正在執行的批次，可考慮等待 proc.Wait(ctx) 或自訂 timeout
				sugar.Info("DDL服務清理完成，服務終止。")
				return

			// case <-ticker.C:

			default:
				func() {
					// 每次批次開始時，建立一個帶期限的 Context（例如 30 秒）
					batchCtx, cancel := context.WithTimeout(ctx, cfg.ProcessTimeout)
					// 確保在此批次結束後取消，避免 context 泄漏
					defer cancel() // 這個 defer 屬於這個匿名函式，而不是main

					// 執行一次批次
					start := time.Now()
					var logCtn int
					if err := proc.DdlLogProcess(batchCtx, &logCtn); err != nil {
						metrics.ProcessErrors.WithLabelValues("ddl").Inc()
						sugar.Fatalf("DDL批次處理失敗：", zap.Error(err))
					} else {
						metrics.ProcessRuns.WithLabelValues("ddl").Inc()
						if logCtn > 0 {
							sugar.Info("DML批次處理完成，共%d筆", logCtn)
						}
					}
					metrics.ProcessDuration.WithLabelValues("ddl").Observe(time.Since(start).Seconds())
				}()
			}
		}
	}()

	// 10-2. Producer 2 號 處理DML: DmlLogProcess
	go func() {
		defer wg.Done()

		// 10-2-1. 建立 Ticker，依 ProcessDmlInterval 週期執行
		// ticker := time.NewTicker(cfg.ProcessDmlInterval)
		// defer ticker.Stop()

		sugar.Info("DML 服務啟動，每" + cfg.ProcessDdlInterval.String() + "秒執行一次批次處理")

		// 10-2-2. 主迴圈: 呼叫 Processor 處理（序列化 + 發送），並收集 metrics，並優雅關閉
		for {
			select {
			case <-ctx.Done():
				sugar.Info("DML服務收到關機訊號，開始清理...")
				// TODO: 若有正在執行的批次，可考慮等待 proc.Wait(ctx) 或自訂 timeout
				sugar.Info("DML服務清理完成，服務終止。")
				return

			// case <-ticker.C:
			default:
				func() {
					// 每次批次開始時，建立一個帶期限的 Context（例如 30 秒）
					batchCtx, cancel := context.WithTimeout(ctx, cfg.ProcessTimeout)
					// 確保在此批次結束後取消，避免 context 泄漏
					defer cancel() // 這個 defer 屬於這個匿名函式，而不是main

					// 執行一次批次
					start := time.Now()
					var logCtn int
					if err := proc.DmlLogProcess(batchCtx, &logCtn); err != nil {
						metrics.ProcessErrors.WithLabelValues("dml").Inc()
						sugar.Fatalf("DML批次處理失敗：", zap.Error(err))
					} else {
						metrics.ProcessRuns.WithLabelValues("dml").Inc()
						if logCtn > 0 {
							sugar.Info("DML批次處理完成，共 " + strconv.Itoa(logCtn) + " 筆")
						}
					}
					metrics.ProcessDuration.WithLabelValues("dml").Observe(time.Since(start).Seconds())
				}()
			}
		}
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	<-done
}
