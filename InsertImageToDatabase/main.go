package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	_ "github.com/microsoft/go-mssqldb"
)

func main() {
	// 定義 CLI 參數
	dbServer := flag.String("server", "", "資料庫伺服器名稱 (如 TEST\\SPS)")
	dbName := flag.String("db", "", "資料庫名稱")
	userID := flag.String("user", "", "SQL 使用者帳號")
	password := flag.String("password", "", "SQL 密碼")
	tableName := flag.String("table", "", "資料表名稱")
	pkey := flag.String("pkey", "", "主鍵 (用逗號分隔，如 'ID,Seq')")
	pkeyValues := flag.String("pval", "", "主鍵對應值 (用逗號分隔，如 '1001,1')")
	imageColumn := flag.String("column", "", "圖片欄位名稱")
	imagePath := flag.String("image", "", "圖片路徑")

	// 解析 CLI 參數
	flag.Parse()

	// 檢查必要參數
	if *imagePath == "" || *dbServer == "" || *dbName == "" || *userID == "" || *password == "" || *tableName == "" || *pkey == "" || *pkeyValues == "" || *imageColumn == "" {
		log.Fatal("缺少必要參數，請輸入完整指令，例如：\n" +
			"upload.exe -image=\"C:\\Users\\benson.chung\\Downloads\\SciMIsTool (2).jpg\" -server=\"TEST\\MIS\" -db=\"MES\" -user=\"SCIMIS\" -password=\"27128299\" -table=\"PMSFile\" -pkey=\"ID,Seq\" -column=\"Image\" -pval=\"1001,1\"")
	}

	// 讀取圖片
	imageData, err := ioutil.ReadFile(*imagePath)
	if err != nil {
		log.Fatalf("讀取圖片失敗: %v", err)
	}

	// 建立 SQL Server 連線字串
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s;encrypt=disable",
		*dbServer, *userID, *password, *dbName)

	// 連接 SQL Server
	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatalf("無法連接 SQL Server: %v", err)
	}
	defer db.Close()

	// 解析主鍵欄位與值
	pkeyFields := strings.Split(*pkey, ",")
	pkeyVals := strings.Split(*pkeyValues, ",")

	if len(pkeyFields) != len(pkeyVals) {
		log.Fatal("PKey 欄位數與 PKey 值數不匹配，請檢查輸入")
	}

	// 建立 WHERE 條件
	var whereClauses []string
	var args []interface{}

	for i, field := range pkeyFields {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = @p%d", field, i+1))
		args = append(args, pkeyVals[i])
	}
	whereSQL := strings.Join(whereClauses, " AND ")

	// SQL 更新語法
	query := fmt.Sprintf("UPDATE %s.dbo.%s SET %s = @image WHERE %s", *dbName, *tableName, *imageColumn, whereSQL)

	// 執行 SQL 更新
	args = append(args, sql.Named("image", imageData))
	_, err = db.Exec(query, args...)
	if err != nil {
		log.Fatalf("圖片儲存失敗: %v", err)
	}

	fmt.Println("✅ 圖片已成功儲存至資料庫！")
}
