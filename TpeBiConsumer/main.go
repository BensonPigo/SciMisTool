// 啟動服務與排程
package main

import (
	"TpeBiConsumer/config"
	mq "TpeBiConsumer/mq"
	scilog "TpeBiConsumer/scilog"
	"TpeBiConsumer/service"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	// "github.com/prometheus/client_golang/prometheus/promhttp"
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

	// 3. 載入設定檔
	cfg, err := config.LoadConfig(config.ConfigFilePath)
	if err != nil {
		sugar.Fatalf("載入設定失敗：%v", err)
	}

	// 4. 啟動 Metrics Server. 先啟動一個專門用來暴露 /metrics 的 HTTP 伺服器
	// go func() {
	//         http.Handle("/metrics", promhttp.Handler())
	//         sugar.Info(fmt.Sprintf("Metrics 伺服器啟動，監聽 :%d/metrics", cfg.Prometheus.MetricsPort))
	//         if err := http.ListenAndServe(":"+strconv.Itoa(cfg.Prometheus.MetricsPort), nil); err != nil {
	//                 sugar.Fatalf("Metrics server 錯誤", zap.Error(err))
	//         }
	// }()

	// 5. 建立可取消的 Context，訂閱 SIGINT 和 SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 6. 建立 MQ Client
	mqClient, err := mq.NewMQClient(cfg.MQ)
	if err != nil {
		sugar.Fatalf("MQ 初始化失敗：%v", err)
	}
	defer mqClient.Close()

	// 7. 初始化資料庫
	db, err := config.InitGormDB(cfg.DB)
	if err != nil {
		sugar.Fatalf("初始化資料庫失敗：%v", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	// 8. 建立 Processor
	proc := service.NewProcessor(db)

	// 9. 建立 consumer 列表
	consumerCount := cfg.ConsumerCount
	consumers := make([]*mq.Consumer, 0, consumerCount)

	for i := 0; i < consumerCount; i++ {
		idx := i // 建立一個新的區域複本
		c := mq.NewConsumer(mqClient, sugar)

		// // （可選）設定 prefetch count，避免一次拉太多未 Ack 的訊息
		// prefetchCount := 1
		// if err := c.Ch().Qos(prefetchCount, 0, false); err != nil {
		// 	log.Fatalf("設定 Qos 失敗：%v", err)
		// }

		// 啟動並行處理
		err := c.Start(ctx, func(ctx context.Context, routingKey string, body []byte) error {
			sugar.Infof("[Consumer %d] 收到訊息，RoutingKey=%s", idx, routingKey)
			switch routingKey {
			case string(mq.RoutingKeyDDL):
				return proc.DdlLogProcess(ctx, body)
			case string(mq.RoutingKeyDML):
				return proc.DmlLogProcess(ctx, body)
			default:
				return fmt.Errorf("無效 routing key: %s", routingKey)
			}
		})
		if err != nil {
			sugar.Fatalf("啟動 Consumer[%d] 失敗：%v", i, err)
		}
		sugar.Infof("Consumer[%d] 已啟動", i)
		consumers = append(consumers, c)
	}

	sugar.Infof("所有 %d 個 Consumer 已啟動，等待訊息...", consumerCount)

	// 8. 等待關機訊號
	<-ctx.Done()
	sugar.Info("收到關機信號，開始等待 consumer 處理結束…")

	// 9. 依序等待所有 consumer 結束
	for i, c := range consumers {
		sugar.Infof("等待 Consumer[%d] 完成…", i)
		c.Wait()
	}
	sugar.Info("所有 Consumer 處理完畢，優雅關閉")
}
