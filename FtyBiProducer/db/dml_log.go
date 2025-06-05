// DdlLog 查詢與更新
package db

import (
	"FtyBiProducer/model"
	"context"
	"strings"

	"gorm.io/gorm"
)

// 找出未處理的DdlLog (ReceivedByTPE = False)
func GetUnprocessedDmlLogs(ctx context.Context, db *gorm.DB) ([]model.DmlLog, error) {
	var unProcessDdlLog []model.DmlLog
	const batchSize = 1000

	if err := db.WithContext(ctx).Where("ReceivedByTPE = ?", false).Order("SerialNo").Limit(batchSize).Find(&unProcessDdlLog).Error; err != nil {
		return nil, err
	}

	// 把所有換行都換成字面上的 \n
	for i := range unProcessDdlLog {
		unProcessDdlLog[i].JSON = strings.ReplaceAll(unProcessDdlLog[i].JSON, "\n", `\\n`)
		// 如果還有 \r，也可以：
		unProcessDdlLog[i].JSON = strings.ReplaceAll(unProcessDdlLog[i].JSON, "\r", `\\r`)
	}
	return unProcessDdlLog, nil
}

// MarkDmlProcessedByBatch 將指定 SerialNo 的 ReceivedByTPE 設為 true
func MarkDmlProcessedByBatch(ctx context.Context, db *gorm.DB, batchID int64) error {
	// 1. 撈出那筆 LogBatchDmlRecord
	var rec model.LogBatchDmlRecord

	if err := db.WithContext(ctx).
		First(&rec, "LogBatchDmlRecordID = ?", batchID).
		Error; err != nil {
		return err
	}

	//  2. 分批更新：在 ProcessFrom ~ ProcessTo 範圍內，一次更新所有 DmlLog
	res := db.WithContext(ctx).
		Model(&model.DmlLog{}).
		Where("SerialNo BETWEEN ? AND ? AND ReceivedByTPE = ?", rec.SerialNoFrom, rec.SerialNoTo, false).
		Update("ReceivedByTPE", true)
	if err := res.Error; err != nil {
		return err
	}

	return nil
}
