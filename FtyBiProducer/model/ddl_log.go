// DdlLog 結構定義
package model

import "time"

type DdlLog struct {
	SerialNo      int64     `gorm:"column:SerialNo;primaryKey"`
	XML           string    `gorm:"column:XML"`
	ReceivedByTPE bool      `gorm:"column:ReceivedByTPE"`
	GenerateDate  time.Time `gorm:"column:GenerateDate"`
}

// TableName 明確指定資料表名稱
func (DdlLog) TableName() string {
	return "DdlLog"
}
