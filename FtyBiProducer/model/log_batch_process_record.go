// 批次處理記錄結構
package model

import "time"

type LogBatchDmlRecord struct {
	LogBatchDmlRecordID int64     `gorm:"column:LogBatchDmlRecordID;primaryKey"`
	SerialNoFrom        int64     `gorm:"column:SerialNoFrom"`
	SerialNoTo          int64     `gorm:"column:SerialNoTo"`
	ProcessTime         time.Time `gorm:"column:ProcessTime->"` // 加上 -> tag，表示「只讀欄位」，GORM 不會在 INSERT 或 UPDATE 時帶這個欄位
}

// TableName 明確指定資料表名稱
func (LogBatchDmlRecord) TableName() string {
	return "LogBatchDmlRecord"
}

type LogBatchDdlRecord struct {
	LogBatchDdlRecordID int64     `gorm:"column:LogBatchDdlRecordID;primaryKey"`
	SerialNoFrom        int64     `gorm:"column:SerialNoFrom"`
	SerialNoTo          int64     `gorm:"column:SerialNoTo"`
	ProcessTime         time.Time `gorm:"column:ProcessTime->"` // 加上 -> tag，表示「只讀欄位」，GORM 不會在 INSERT 或 UPDATE 時帶這個欄位
}

// TableName 明確指定資料表名稱
func (LogBatchDdlRecord) TableName() string {
	return "LogBatchDdlRecord"
}
