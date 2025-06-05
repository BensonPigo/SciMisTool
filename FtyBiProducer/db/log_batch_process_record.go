// 批次處理記錄新增與查詢
package db

import (
	"FtyBiProducer/model"
	"context"

	"gorm.io/gorm"
)

// 寫入批次處理紀錄
func InsertLogBatchDdlRecord(ctx context.Context, db *gorm.DB, record *model.LogBatchDdlRecord) (int64, error) {

	// InsertLogBatchProcessRecord 建立一筆處理紀錄並回傳 ID
	// 明確跳過 ProcessTime 欄位
	if err := db.WithContext(ctx).
		Omit("ProcessTime").
		Create(record).
		Error; err != nil {
		return 0, err
	}

	return record.LogBatchDdlRecordID, nil
}
func InsertLogBatchDmlRecord(ctx context.Context, db *gorm.DB, record *model.LogBatchDmlRecord) (int64, error) {

	// InsertLogBatchProcessRecord 建立一筆處理紀錄並回傳 ID
	// 明確跳過 ProcessTime 欄位
	if err := db.WithContext(ctx).
		Omit("ProcessTime").
		Create(record).
		Error; err != nil {
		return 0, err
	}

	return record.LogBatchDmlRecordID, nil
}
