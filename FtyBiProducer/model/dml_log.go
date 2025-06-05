// DdlLog 結構定義
package model

import "time"

type DmlLog struct {
	SerialNo      int64     `gorm:"column:SerialNo;primaryKey"`
	JSON          string    `gorm:"column:JSON"`
	ReceivedByTPE bool      `gorm:"column:ReceivedByTPE"`
	GenerateDate  time.Time `gorm:"column:GenerateDate"`
}

// TableName 明確指定資料表名稱
func (DmlLog) TableName() string {
	return "DmlLog"
}
