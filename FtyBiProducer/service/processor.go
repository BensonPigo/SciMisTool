// 核心處理邏輯
package service

import (
	dbLayer "FtyBiProducer/db"
	"FtyBiProducer/model"
	"FtyBiProducer/mq"
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"gorm.io/gorm"
)

type Processor struct {
	db       *gorm.DB
	mqClient *mq.MQClient
}

// New 建構 Processor 時，把 MQ Client 傳進來
func New(db *gorm.DB, mqClient *mq.MQClient) *Processor {
	return &Processor{db: db, mqClient: mqClient}
}

// Process 負責撈取、打包 message 組成 JSON、呼叫 Publish
func (p *Processor) DdlLogProcess(ctx context.Context, logCtn *int) error {
	// 取得待處理 DDL Log
	ddlLogs, err := dbLayer.GetUnprocessedDdlLogs(ctx, p.db)

	if err != nil {
		return fmt.Errorf(" 查詢失敗%v", err)
	}

	*logCtn = len(ddlLogs)
	if len(ddlLogs) > 0 {
		// 找出這個區間最早~最晚的時間
		minSN, maxSN := ddlLogs[0].SerialNo, ddlLogs[0].SerialNo
		for _, log := range ddlLogs[1:] {
			if log.SerialNo < minSN {
				minSN = log.SerialNo
			}
			if log.SerialNo > maxSN {
				maxSN = log.SerialNo
			}
		}

		// 建立批次處理的紀錄
		record := model.LogBatchDdlRecord{
			SerialNoFrom: minSN,
			SerialNoTo:   maxSN,
		}

		// 批次處理紀錄 寫入DB
		batchID, err := dbLayer.InsertLogBatchDdlRecord(ctx, p.db, &record)
		if err != nil {
			return fmt.Errorf("新增 LogBatchDdlRecord 失敗：%v", err)
		}

		// 取出所有 XML
		var xmlList []string
		for _, log := range ddlLogs {
			xmlList = append(xmlList, log.XML)
		}

		// 包裝成訊息
		message := model.DdlMessage{
			BatchID: batchID,
			XMLList: xmlList,
		}

		// JSON 編碼
		jsonBytes, err := sonic.Marshal(message)
		if err != nil {
			return fmt.Errorf("轉換 JSON 失敗：%v", err)
		}

		// 無資料則結束
		if jsonBytes == nil {
			return nil
		}

		// 發送消息
		if err := p.mqClient.Publish(ctx, mq.RoutingKeyProducerDDL, jsonBytes); err != nil {
			return fmt.Errorf("發送 MQ 訊息失敗：%w", err)
		}

		if err := dbLayer.MarkDdlProcessedByBatch(ctx, p.db, batchID); err != nil {
			return fmt.Errorf("標記DDL_Log失敗 : %w", err)
		}

	}
	return nil

}

func (p *Processor) DmlLogProcess(ctx context.Context, logCtn *int) error {
	// 取得待處理 DML Log
	dmlLogs, err := dbLayer.GetUnprocessedDmlLogs(ctx, p.db)

	if err != nil {
		return fmt.Errorf(" 查詢失敗%v", err)
	}
	*logCtn = len(dmlLogs)
	if len(dmlLogs) > 0 {
		// 找出這個區間最小~最大的SerialNo
		minSN, maxSN := dmlLogs[0].SerialNo, dmlLogs[0].SerialNo
		for _, log := range dmlLogs[1:] {
			if log.SerialNo < minSN {
				minSN = log.SerialNo
			}
			if log.SerialNo > maxSN {
				maxSN = log.SerialNo
			}
		}

		// 建立批次處理的紀錄
		record := model.LogBatchDmlRecord{
			SerialNoFrom: minSN,
			SerialNoTo:   maxSN,
		}

		// 批次處理紀錄 寫入DB
		batchID, err := dbLayer.InsertLogBatchDmlRecord(ctx, p.db, &record)
		if err != nil {
			return fmt.Errorf("新增 ProcessRecord 失敗：%v", err)
		}

		// 取出所有 JSON
		var jsonList []string
		for _, log := range dmlLogs {
			jsonList = append(jsonList, log.JSON)
		}

		// 包裝成訊息
		message := model.DmlMessage{
			BatchID:  batchID,
			JSONList: jsonList,
		}

		// JSON 編碼
		jsonBytes, err := sonic.Marshal(message)
		if err != nil {
			return fmt.Errorf("轉換 JSON 失敗：%v", err)
		}

		// 無資料則結束
		if jsonBytes == nil {
			return nil
		}

		// 發送消息
		if err := p.mqClient.Publish(ctx, mq.RoutingKeyProducerDML, jsonBytes); err != nil {
			return fmt.Errorf("發送 MQ 訊息失敗：%w", err)
		}

		// 開啟transaction
		tx := p.db.WithContext(ctx).Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				panic(r)
			}
		}()
		if err := dbLayer.MarkDmlProcessedByBatch(ctx, p.db, batchID); err != nil {
			return fmt.Errorf("標記DML_Log失敗 : %w", err)
		}

		// commit
		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("commit 失敗：%w", err)
		}
	}
	return nil

}
