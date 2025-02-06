package db

import (
	"log"

	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	once sync.Once // 確保資料庫連接只初始化一次
)

// InitDB 初始化資料庫連接
func InitDB(databasePath string) {
	once.Do(func() { // 單例模式
		var err error
		db, err = gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
		if err != nil {
			log.Fatalf("無法連接資料庫: %v", err)
		}
	})
}

// GetDB 獲取資料庫連接
func GetDB() *gorm.DB {
	if db == nil {
		log.Fatal("資料庫尚未初始化。請先調用 InitDB。")
	}
	return db
}
