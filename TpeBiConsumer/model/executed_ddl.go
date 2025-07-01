package model

import "time"

// ExecutedDDL 記錄已經成功執行過的 DDL 指令
// SQLHash 為 primary key，用於避免重複
// SQLText 則保存原始 SQL 內容，便於追蹤
// CreatedAt 紀錄執行時間，由資料庫自動填入

type ExecutedDDL struct {
	SQLHash   string    `gorm:"column:SQLHash;primaryKey"`
	SQLText   string    `gorm:"column:SQLText"`
	CreatedAt time.Time `gorm:"column:CreatedAt;autoCreateTime"`
}

func (ExecutedDDL) TableName() string {
	return "ExecutedDDL"
}
