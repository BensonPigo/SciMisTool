// 核心處理邏輯
package service

import (
	dbLayer "FtyBiProducer/db"
	"FtyBiProducer/model"
	"FtyBiProducer/mq"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/bytedance/sonic"
	"gorm.io/gorm"
)

type Processor struct {
	db       *gorm.DB
	mqClient *mq.MQClient
	// 緩存： key = tableName, value = []gorm.ColumnType
	colTypeCache map[string][]gorm.ColumnType
}

// New 建構 Processor 時，把 MQ Client 傳進來
func New(db *gorm.DB, mqClient *mq.MQClient) *Processor {
	return &Processor{
		db:           db,
		mqClient:     mqClient,
		colTypeCache: make(map[string][]gorm.ColumnType),
	}
}

func (p *Processor) getColumnTypesOnce(ctx context.Context, tableName string) ([]gorm.ColumnType, error) {
	if types, ok := p.colTypeCache[tableName]; ok {
		return types, nil
	}
	types, err := p.db.WithContext(ctx).Migrator().ColumnTypes(tableName)
	if err != nil {
		return nil, err
	}
	p.colTypeCache[tableName] = types
	return types, nil
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
		if err := p.mqClient.Publish(ctx, mq.RoutingKeyDDL, jsonBytes); err != nil {
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
		if err := p.mqClient.Publish(ctx, mq.RoutingKeyDML, jsonBytes); err != nil {
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

func (p *Processor) DmlLogGenerate(ctx context.Context) error {
	// 1. 從 BITaskInfo.Name 取得所有目標資料表
	var tableNames []string
	if err := p.db.WithContext(ctx).Table("BITaskInfo").Pluck("Name", &tableNames).Error; err != nil {
		return fmt.Errorf("查詢 BITaskInfo 失敗: %w", err)
	}
	for _, name := range tableNames {
		if _, err := p.getColumnTypesOnce(ctx, name); err != nil {
			return fmt.Errorf("取得 %s 欄位資訊失敗: %w", name, err)
		}
		if err := p.generateLogsForTable(ctx, name, "Insert"); err != nil {
			return err
		}
		hist := fmt.Sprintf("%s_History", name)
		if _, err := p.getColumnTypesOnce(ctx, hist); err != nil {
			return fmt.Errorf("取得 %s 欄位資訊失敗: %w", hist, err)
		}
		if err := p.generateLogsForTable(ctx, hist, "Delete"); err != nil {
			return err
		}
	}
	return nil
}

func (p *Processor) generateLogsForTable(ctx context.Context, tableName, action string) error {
	var rows []map[string]interface{}
	if err := p.db.WithContext(ctx).
		Table(tableName).
		Where("BIStatus = ?", "New").
		Find(&rows).Error; err != nil {
		return fmt.Errorf("查詢 %s 失敗: %w", tableName, err)
	}
	for _, row := range rows {
		tx := p.db.WithContext(ctx).Begin()
		if err := tx.Error; err != nil {
			return fmt.Errorf("開啟 transaction 失敗: %w", err)
		}

		data := make(map[string]interface{}, len(row)+1)
		data["TableName"] = tableName
		for k, v := range row {
			data[k] = v
		}
		entry := map[string]interface{}{
			"Action": action,
			"Data":   data,
		}
		jsonBytes, err := sonic.Marshal(entry)
		if err != nil {
			_ = p.updateBIStatus(ctx, tx, tableName, row, "Pending")
			tx.Rollback()
			return fmt.Errorf("JSON 編碼失敗: %w", err)
		}
		if err := tx.Exec("INSERT INTO [DmlLog]([JSON])VALUES(?)", string(jsonBytes)).Error; err != nil {
			_ = p.updateBIStatus(ctx, tx, tableName, row, "Pnding")
			tx.Rollback()
			return fmt.Errorf("寫入 DmlLog 失敗: %w", err)
		}

		// 完成 Status = Complete
		if err := p.updateBIStatus(ctx, tx, tableName, row, "Complete"); err != nil {
			tx.Rollback()
			return fmt.Errorf("更新 BIStatus 失敗: %w", err)
		}

		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("commit 失敗：%w", err)
		}
	}

	return nil
}

// updateBIStatus 根據指定表格的主鍵欄位更新資料列的 BIStatus
func (p *Processor) updateBIStatus(ctx context.Context, tx *gorm.DB, table string, row map[string]interface{}, status string) error {
	// 取得欄位描述
	columnTypes, err := p.getColumnTypesOnce(ctx, table)
	if err != nil {
		return err
	}

	// ★ 先正規化 row，避免 []byte 被誤判 ★
	normalizeRow(columnTypes, row)

	// 取得table pkey
	pkCond := make(map[string]interface{})
	// 有些table 無 pkey，則全欄位比對
	fullCond := make(map[string]interface{})
	for _, ct := range columnTypes {
		name := ct.Name()
		val, ok := row[name]
		if !ok {
			continue
		}
		if name == "FOB" {
			fullCond[name] = val
		}
		fullCond[name] = val
		if isPK, _ := ct.PrimaryKey(); isPK {
			pkCond[name] = val
		}
	}
	// 決定使用 fullCond 還是 pkCond
	cond := fullCond
	if len(pkCond) != 0 {
		cond = pkCond
	}

	// 執行更新
	res := tx.Table(table).Where(cond).Update("BIStatus", status)
	if res.Error != nil {
		return res.Error
	}

	// 如果沒更新到任何列，就回傳 ErrRecordNotFound
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// 將 row 內的 []byte / []uint8 轉成正確型別
func normalizeRow(cols []gorm.ColumnType, row map[string]interface{}) {
	for _, c := range cols {
		name := c.Name()

		v, ok := row[name]
		if !ok {
			continue
		}
		b, ok := v.([]byte) // gorm 回傳的是 []uint8，與 []byte 同底層
		if !ok {
			continue
		}

		dbType := strings.ToUpper(c.DatabaseTypeName())
		switch dbType {
		case "NUMERIC", "DECIMAL", "MONEY", "SMALLMONEY":
			// 需比較數值 → 轉成 float64（或自行改用 decimal 套件）
			if f, err := strconv.ParseFloat(string(b), 64); err == nil {
				row[name] = f
			} else {
				row[name] = string(b) // 解析失敗就先用字串
			}
		default:
			// 其他型別一律轉成字串
			row[name] = string(b)
		}
	}
}
