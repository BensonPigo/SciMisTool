// DdlLog 查詢與更新
package db

import (
	"FtyBiProducer/model"
	"context"

	"gorm.io/gorm"
)

// 找出未處理的DdlLog (ReceivedByTPE = False)
func GetUnprocessedDdlLogs(ctx context.Context, db *gorm.DB) ([]model.DdlLog, error) {
	var unProcessDdlLog []model.DdlLog
	const batchSize = 10000
	if err := db.WithContext(ctx).Where("ReceivedByTPE = ?", false).Order("SerialNo").Limit(batchSize).Find(&unProcessDdlLog).Error; err != nil {
		return nil, err
	}

	return unProcessDdlLog, nil
}

// MarkDdlProcessedByBatch 將指定 SerialNo 的 ReceivedByTPE 設為 true
func MarkDdlProcessedByBatch(ctx context.Context, db *gorm.DB, batchID int64) error {
	// 1. 撈出那筆 LogBatchDdlRecord
	var rec model.LogBatchDdlRecord
	if err := db.WithContext(ctx).
		First(&rec, "LogBatchDdlRecordID = ?", batchID).
		Error; err != nil {
		return err
	}

	// 2. 在 ProcessFrom ~ ProcessTo 範圍內，一次更新所有 DdlLog
	if err := db.WithContext(ctx).
		Model(&model.DdlLog{}).
		Where("SerialNo BETWEEN ? AND ?", rec.SerialNoFrom, rec.SerialNoTo).
		Update("ReceivedByTPE", true).
		Error; err != nil {
		return err
	}

	return nil
}
