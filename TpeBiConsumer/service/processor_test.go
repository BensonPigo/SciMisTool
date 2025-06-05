package service

import (
	model "TpeBiConsumer/model"
	"context"
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func setupDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	// 讀取真實測試 DB 的連線字串
	dsn := `Server=testing\PH1;User ID=SCIMIS;Password=27128299;Database=POWERBIReportData;Encrypt=disable`
	if dsn == "" {
		t.Fatal("環境變數 TEST_SQLSERVER_DSN 尚未設定")
	}

	// 連到真實 MS SQL
	db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("無法連線到測試 MS SQL: %v", err)
	}

	// 回傳一個 cleanup，用來在測試結束時關閉 DB 連線
	cleanup := func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}

	// 第二個回傳值留作舊版 mock 相容，永遠回傳 nil
	return db, nil, cleanup
}

// TestDdlLogProcess 使用 MS SQL 測試 DDL 執行
func TestDdlLogProcess_With_Testing_SQLServer(t *testing.T) {

	// 1. 設定測試DB
	gormDB, _, cleanup := setupDB(t)
	defer cleanup()

	// 2. 開始模擬，測試 ALTER TABLE [dbo].[P_CuttingBCS] ALTER COLUMN MatchFabric varchar(101) NOT NULL 能不能被正常執行
	xml := `<DDLData><EventData><EVENT_INSTANCE><EventType>ALTER_TABLE</EventType><PostTime>2025-05-28T09:21:48.627</PostTime><SPID>59</SPID><ServerName>TESTING\PH1</ServerName><LoginName>DOMAIN\benson.chung</LoginName><UserName>dbo</UserName><DatabaseName>POWERBIReportData</DatabaseName><SchemaName>dbo</SchemaName><ObjectName>P_CuttingBCS</ObjectName><ObjectType>TABLE</ObjectType><AlterTableActionList><Alter><Columns><Name>MatchFabric</Name></Columns></Alter></AlterTableActionList><TSQLCommand><SetOptions ANSI_NULLS="ON" ANSI_NULL_DEFAULT="ON" ANSI_PADDING="ON" QUOTED_IDENTIFIER="ON" ENCRYPTED="FALSE" />` +
		`<CommandText>ALTER TABLE [dbo].[P_CuttingBCS] ALTER COLUMN MatchFabric varchar(101) NOT NULL</CommandText>` + // 這裡是關鍵
		`</TSQLCommand></EVENT_INSTANCE></EventData><Timestamp>2025-05-28T09:21:48.627</Timestamp></DDLData>`

	// 包裝成訊息
	var xmlList []string
	xmlList = append(xmlList, xml)
	message := model.DdlMessage{
		BatchID: 111,
		XMLList: xmlList,
	}

	jsonBytes, _ := json.Marshal(message)

	// 3. 目的是要測 DdlLogProcess
	proc := NewProcessor(gormDB)
	err := proc.DdlLogProcess(context.Background(), jsonBytes)
	if err != nil {
		t.Errorf("預期不會錯誤，實際: %v", err)
	}

	// 再去查 INFORMATION_SCHEMA.COLUMNS 確認欄位已經正確修改
	var cnt int
	gormDB.Raw(`
  SELECT COUNT(*) 
  FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA='dbo' 
    AND TABLE_NAME='P_CuttingBCS' 
    AND COLUMN_NAME='MatchFabric'
	AND CHARACTER_MAXIMUM_LENGTH=101
`).Scan(&cnt)
	if cnt != 1 {
		t.Errorf("欄位沒有異動，cnt = %d", cnt)
	}
}

// TestDmlLogProcess 使用 MS SQL 測試 DML Delete 與 Insert
func TestDmlLogProcess_Success_MockSQLServer(t *testing.T) {
	gormDB, _, cleanup := setupDB(t)
	defer cleanup()

	// 模擬Producer發送的消息
	// 測試案例：P_CuttingBCS資料表，刪掉 BIInsertDate = 2025-04-27 16:00:31.450、新增 BIInsertDate = 2025-04-28 16:00:31.450
	// 關注 P_CuttingBCS.BIInsertDate 即可
	jsonDel := `{"Action":"Delete","Data":{"TableName":"P_CuttingBCS","HistoryUkey":"1","OrderID":"25080109II023","SewingLineID":"02","RequestDate":"2025-07-24","BIFactoryID":"PH1","BIInsertDate":"2025-04-28 16:00:31.450"}}`
	jsonIns := `{"Action":"Insert","Data":{"TableName":"P_CuttingBCS","MDivisionID":"PM2","FactoryID":"MWI","BrandID":"LLL","StyleID":"LM7B92S","SeasonID":"25WI","CDCodeNew":"SKNNM","FabricType":"KNIT","POID":"25080109II","Category":"Bulk","WorkType":"","MatchFabric":"  ","OrderID":"25080109II023","SciDelivery":"2025-08-15","BuyerDelivery":"2025-09-01","OrderQty":"2403","SewInLineDate":"2025-07-21","SewOffLineDate":"2025-07-26","SewingLineID":"02","RequestDate":"2025-07-24","StdQty":"499","StdQtyByLine":"499","AccuStdQty":"1845","AccuStdQtyByLine":"1845","AccuEstCutQty":"0","AccuEstCutQtyByLine":"0","SupplyCutQty":"0","SupplyCutQtyByLine":"0","BalanceCutQty":"0","BalanceCutQtyByLine":"0","SupplyCutQtyVSStdQty":"499","SupplyCutQtyVSStdQtyByLine":"499","BIFactoryID":"PH1","BIInsertDate":"2025-04-27 16:00:31.450"}}`

	var jsonList []string
	jsonList = append(jsonList, jsonDel)
	jsonList = append(jsonList, jsonIns)
	message := model.DmlMessage{
		BatchID:  111,
		JSONList: jsonList,
	}
	jsonBytes, _ := json.Marshal(message)

	// consumer 接收並處理
	proc := NewProcessor(gormDB)
	err := proc.DmlLogProcess(context.Background(), jsonBytes)
	if err != nil {
		t.Errorf("預期不會錯誤，實際: %v", err)
	}

	// 再去確認資料已經正確修改
	var cnt int
	gormDB.Raw(`
select COUNT(1) from P_CuttingBCS
where OrderID='25080109II023'
and SewingLineID='02'
and RequestDate='2025-07-24'
and BIFactoryID='PH1'
and BIInsertDate='2025-04-27 16:00:31.450'
`).Scan(&cnt)
	if cnt != 1 {
		t.Errorf("資料沒有異動，cnt = %d", cnt)
	}
}
