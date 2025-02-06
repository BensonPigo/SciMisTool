package models

import (
	"time"

	"gorm.io/gorm"
)

// User 定義資料庫模型
type User struct {
	ID       int    `gorm:"primaryKey"`
	Email    string `gorm:"uniqueIndex"`
	Username string
	Password string
}

type RefreshToken struct {
	ID        int       `gorm:"primaryKey"`
	Token     string    `gorm:"unique;not null"` // Token 本身
	UserID    int       `gorm:"not null"`        // 關聯的用戶 ID
	ExpiresAt time.Time `gorm:"not null"`        // 過期時間
}

// Migrate 自動遷移資料表
func Migrate(db *gorm.DB) {
	db.AutoMigrate(&User{})
	db.AutoMigrate(&RefreshToken{})
}
