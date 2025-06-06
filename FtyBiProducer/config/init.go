package config

import (
	"fmt"
	"net/url"
	"strings"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

// 由外層傳入 dbCfg，再從裡面取 ConnectionString
func InitGormDB(dbCfg DBConfig) (*gorm.DB, error) {
	// 1. 如果帶 '\'，先拆 host + instance
	host := dbCfg.Host
	instance := dbCfg.Instance
	if parts := strings.SplitN(dbCfg.Host, `\`, 2); len(parts) == 2 {
		host = parts[0]
		instance = parts[1]
	}

	// 2. 密碼做 URL Escape
	pwd := url.QueryEscape(dbCfg.Password)

	// 3. path 模式組 DSN，把 query timeout 加進去
	dsn := fmt.Sprintf(
		"sqlserver://%s:%s@%s/%s"+
			"?database=%s&encrypt=%s&connection+timeout=%d&query+timeout=%d",
		dbCfg.User, pwd,
		host, instance,
		dbCfg.Name, dbCfg.Encrypt,
		int(dbCfg.Timeout.Seconds()),
		int(dbCfg.QueryTimeout.Seconds()),
	)

	// 4. 開始連線
	db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		// 印出 DSN 方便 debug，但請小心不要在正式環境曝光密碼
		return nil, fmt.Errorf("連接 SQL Server 失敗: %w, DSN: %s", err, dsn)
	}
	return db, nil
}
