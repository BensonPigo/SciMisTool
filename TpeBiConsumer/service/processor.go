// 核心處理邏輯
package service

import (
	model "TpeBiConsumer/model"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	mssql "github.com/microsoft/go-mssqldb"

	"gorm.io/gorm"
)

type Processor struct {
	db *gorm.DB
	// 緩存： key = tableName, value = []gorm.ColumnType
	colTypeCache map[string][]gorm.ColumnType
	// 已執行過的 DDL 指令，用於避免重複執行
	executedDDL map[string]struct{}
	ddlMu       sync.RWMutex
}

// 建立只需要 db 的 Processor
func NewProcessor(db *gorm.DB) *Processor {
	p := &Processor{
		db:           db,
		colTypeCache: make(map[string][]gorm.ColumnType),
		executedDDL:  make(map[string]struct{}),
	}
	// 確保 ExecutedDDL 資料表存在
	_ = db.AutoMigrate(&model.ExecutedDDL{})
	return p
}

func (p *Processor) getColumnTypesOnce(ctx context.Context, tableName string) ([]gorm.ColumnType, error) {
	if types, ok := p.colTypeCache[tableName]; ok {
		return types, nil
	}
	types, err := p.db.WithContext(ctx).Migrator().ColumnTypes(tableName)
	if err != nil {
		return nil, err
	}
	p.colTypeCache[tableName] = types
	return types, nil
}

// convertValue 將 raw 任意型別轉成符合 ct（gorm.ColumnType）所對應的 Go 原生型別

func convertValue(raw interface{}, ct gorm.ColumnType) (interface{}, error) {
	if raw == nil {
		return nil, nil
	}

	dbType := strings.ToUpper(ct.DatabaseTypeName())

	switch {
	// ========== 整數 ==========
	case dbType == "INT" || dbType == "BIGINT" || dbType == "SMALLINT" || dbType == "TINYINT":
		switch v := raw.(type) {
		case int64:
			return v, nil
		case float64:
			// 由於 JSON 數字預設會成為 float64，需要轉回 int64
			return int64(v), nil
		case int:
			return int64(v), nil
		case string:
			if v == "" {
				return nil, nil
			}
			parsed, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("欄位 %s 轉 int 失敗: %w", ct.Name(), err)
			}
			return parsed, nil
		default:
			return nil, fmt.Errorf("欄位 %s (int) 不合法型別: %T", ct.Name(), raw)
		}

	// ========== 浮點/數值 ==========
	case dbType == "FLOAT" ||
		strings.HasPrefix(dbType, "DECIMAL") ||
		strings.HasPrefix(dbType, "NUMERIC"):
		switch v := raw.(type) {
		case float64:
			return v, nil
		case string:
			if v == "" {
				return nil, nil
			}
			parsed, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, fmt.Errorf("欄位 %s 轉 float 失敗: %w", ct.Name(), err)
			}
			return parsed, nil
		default:
			return nil, fmt.Errorf("欄位 %s (float) 不合法型別: %T", ct.Name(), raw)
		}

	// ========== 布林 ==========
	case dbType == "BIT":
		switch v := raw.(type) {
		case bool:
			return v, nil
		case string:
			lowered := strings.ToLower(v)
			if lowered == "true" || v == "1" {
				return true, nil
			}
			if lowered == "false" || v == "0" {
				return false, nil
			}
			return nil, fmt.Errorf("欄位 %s 轉 bool 失敗: 無法解析 \"%s\"", ct.Name(), v)
		default:
			return nil, fmt.Errorf("欄位 %s (bool) 不合法型別: %T", ct.Name(), raw)
		}

	// ========== 日期/時間 ==========
	case dbType == "DATE" ||
		dbType == "DATETIME" ||
		dbType == "DATETIME2" ||
		dbType == "SMALLDATETIME":
		switch v := raw.(type) {
		case time.Time:
			return v, nil
		case string:
			if v == "" {
				return nil, nil
			}
			// 嘗試幾個常見格式
			formats := []string{
				"2006-01-02 15:04:05.9999999",
				"2006-01-02 15:04:05",
				"2006-01-02",
				time.RFC3339,
			}
			var parsed time.Time
			var err error
			for _, fmtStr := range formats {
				parsed, err = time.Parse(fmtStr, v)
				if err == nil {
					return parsed, nil
				}
			}
			return nil, fmt.Errorf("欄位 %s 轉 time 失敗: %v", ct.Name(), err)
		default:
			return nil, fmt.Errorf("欄位 %s (time) 不合法型別: %T", ct.Name(), raw)
		}

	// ========== 其餘當字串處理 (VARCHAR, NVARCHAR, CHAR, TEXT, UNIQUEIDENTIFIER…) ==========
	default:
		switch v := raw.(type) {
		case string:
			return v, nil
		case []byte:
			return v, nil
		default:
			// 其他都用 fmt.Sprint() 轉成字串
			return fmt.Sprint(v), nil
		}
	}
}

// batchUpsertWithMerge：在同一條 *mssql.Conn 底下完成
// 1) 建 temp table (#…)
// 2) BulkCopy 寫入 temp table
// 3) MERGE 回正式 table
// 4) DROP temp table
func (p *Processor) batchUpsertWithMerge(ctx context.Context, tableName string, rawDatas []map[string]interface{}) error {
	// 1. 先從 GORM 拿到 *sql.DB
	sqlDB, err := p.db.DB()
	if err != nil {
		return fmt.Errorf("無法從 GORM 拿到 *sql.DB: %w", err)
	}

	// 2. 要求一條專屬 *sql.Conn，保證 temp table 存活在這個 session 裡
	sqlConn, err := sqlDB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("建立 *sql.Conn 失敗: %w", err)
	}
	defer sqlConn.Close()

	// 3. 把 *sql.Conn 轉成 *mssql.Conn，以利後續 BulkCopy、ExecContext 等操作
	var mssqlConn *mssql.Conn
	if err := sqlConn.Raw(func(driverConn interface{}) error {
		c, ok := driverConn.(*mssql.Conn)
		if !ok {
			return fmt.Errorf("預期 driverConn 是 *mssql.Conn，但實際是 %T", driverConn)
		}
		mssqlConn = c
		return nil
	}); err != nil {
		return fmt.Errorf("從 *sql.Conn 取得 *mssql.Conn 失敗: %w", err)
	}

	// 4. 在同一個 mssqlConn session 上，用 ExecContext 建立 temp table
	rawUUID := uuid.New().String()
	uid := strings.ReplaceAll(rawUUID, "-", "")
	tempTable := fmt.Sprintf("#%s_Stagin_%s", tableName, uid)
	createTempSQL := fmt.Sprintf("SELECT TOP 0 * INTO %s FROM %s;", tempTable, tableName)
	if _, err := sqlConn.ExecContext(ctx, createTempSQL); err != nil {
		return fmt.Errorf("建立 temp table %s 失敗: %w", tempTable, err)
	}
	// 5. 如果 rawDatas 為空，直接 drop temp table 並結束
	if len(rawDatas) == 0 {
		if _, err := sqlConn.ExecContext(ctx, fmt.Sprintf("DROP TABLE %s;", tempTable)); err != nil {
			return fmt.Errorf("刪除 temp table %s 失敗: %w", tempTable, err)
		}
		return nil
	}

	// 6. 準備 BulkCopy：先取得正式表 (tableName) 的欄位清單
	columnTypes, err := p.getColumnTypesOnce(ctx, tableName)
	if err != nil {
		return fmt.Errorf("取得 %s 欄位資訊失敗: %w", tableName, err)
	}
	sort.Slice(columnTypes, func(i, j int) bool {
		return columnTypes[i].Name() < columnTypes[j].Name()
	})

	// 排除無法 INSERT 的自動遞增欄位（例如 IDENTITY）
	cols := make([]string, 0, len(columnTypes))
	colTypeMap := make(map[string]gorm.ColumnType, len(columnTypes))
	skipCols := make(map[string]struct{})
	for _, ct := range columnTypes {
		if isAI, ok := ct.AutoIncrement(); ok && isAI {
			skipCols[ct.Name()] = struct{}{}
			continue
		}
		cols = append(cols, ct.Name())
		colTypeMap[ct.Name()] = ct
	}

	// 7. 把 rawDatas 轉成 [][]interface{} 以供 BulkCopy 使用
	convertedRows := make([][]interface{}, 0, len(rawDatas))
	for rowIdx, row := range rawDatas {
		vals := make([]interface{}, len(cols))
		anyValue := false
		for i, col := range cols {
			rawVal, exists := row[col]
			if !exists {
				vals[i] = nil
				continue
			}
			conv, convErr := convertValue(rawVal, colTypeMap[col])
			if convErr != nil {
				return fmt.Errorf("第 %d 筆，欄位 %s 轉換失敗: %w", rowIdx, col, convErr)
			}
			if conv != nil {
				anyValue = true
			}
			vals[i] = conv
		}
		if anyValue {
			convertedRows = append(convertedRows, vals)
		}
	}

	// 8. 建立 BulkCopy 物件，把資料寫進 tempTable
	bulk := mssqlConn.CreateBulk(tempTable, cols)
	for idx, rowVals := range convertedRows {
		if err := bulk.AddRow(rowVals); err != nil {
			return fmt.Errorf(" Bulk AddRow 第 %d 筆失敗: %w", idx, err)
		}
	}
	if _, err := bulk.Done(); err != nil {
		return fmt.Errorf(" Bulk Done 失敗: %w", err)
	}

	// 9. 組合 MERGE 語法並執行
	var allCols, keyCols []string
	for _, ct := range columnTypes {
		if _, skip := skipCols[ct.Name()]; skip {
			continue
		}
		allCols = append(allCols, ct.Name())
		if isPK, _ := ct.PrimaryKey(); isPK {
			keyCols = append(keyCols, ct.Name())
		}
	}
	sort.Strings(allCols)
	sort.Strings(keyCols)

	// 9.1 ON 條件：T.PK = S.PK
	var joinConds []string
	for _, k := range keyCols {
		joinConds = append(joinConds, fmt.Sprintf("T.%s = S.%s", k, k))
	}
	onClause := strings.Join(joinConds, " AND ")

	// 9.2 UPDATE 子句：非 PK 欄位全部更新
	var updateCols []string
	for _, col := range allCols {
		isPrimary := false
		for _, k := range keyCols {
			if col == k {
				isPrimary = true
				break
			}
		}
		if isPrimary {
			continue
		}
		updateCols = append(updateCols, fmt.Sprintf("T.%s = S.%s", col, col))
	}
	updateClause := strings.Join(updateCols, ", ")

	// 9.3 INSERT 欄位列表與 VALUES 列表
	insertCols := strings.Join(allCols, ", ")
	var insertVals []string
	for _, col := range allCols {
		insertVals = append(insertVals, "S."+col)
	}
	insertValsClause := strings.Join(insertVals, ", ")

	mergeSQL := fmt.Sprintf(`
MERGE INTO %s AS T
USING %s AS S
ON %s
WHEN MATCHED THEN
    UPDATE SET %s
WHEN NOT MATCHED THEN
    INSERT (%s) VALUES (%s);`,
		tableName, tempTable,
		onClause,
		updateClause,
		insertCols, insertValsClause,
	)
	if _, err := sqlConn.ExecContext(ctx, mergeSQL); err != nil {
		return fmt.Errorf("執行 MERGE 失敗: %w", err)
	}

	// 10. Merge 完後，DROP 掉這張 temp table
	dropSQL := fmt.Sprintf("DROP TABLE %s;", tempTable)
	if _, err := sqlConn.ExecContext(ctx, dropSQL); err != nil {
		return fmt.Errorf("DROP temp table %s 失敗: %w", tempTable, err)
	}

	return nil
}

func sanitizeEscape(raw string) string {
	// 只要不是 \", \\, \/, \b, \f, \n, \r, \t, \uXXXX，就把 "\" 自動前置 "\\"。
	// 以避免出現 \5 這種不合法跳脫序列。
	re := regexp.MustCompile(`\\([^"\\/bfnrtu])`)
	return re.ReplaceAllString(raw, `\\\\$1`)
}

func (p *Processor) DmlLogProcess(ctx context.Context, body []byte) error {
	// 解析 JSON
	var msg model.DmlMessage
	if err := sonic.Unmarshal(body, &msg); err != nil {
		return fmt.Errorf("解析 DdlMessage 失敗: %w", err)
	}

	// 定義一個簡單的結構來存放 Action + Data，解析各筆 JSON 物件
	type dmlEntry struct {
		Action string                 `json:"Action"`
		Data   map[string]interface{} `json:"Data"`
	}

	// 把 JSONList 裡的每筆字串都解成 dmlEntry
	var entries []dmlEntry
	for idx, raw := range msg.JSONList {
		var e dmlEntry
		err := sonic.Unmarshal([]byte(raw), &e)

		if err == nil {
			// 直接解析成功，存下結果
			entries = append(entries, e)
			continue
		}

		// 2) 如果錯誤訊息裡面包含「invalid escape」或類似，就做 sanitize，再重試
		if _, ok := err.(*json.UnmarshalTypeError); ok {
			// 如果是型別不符，可能不是跳脫字元問題，可直接回傳
			return fmt.Errorf("第 %d 筆 JSON 型別錯誤: %w", idx, err)
		}
		// 判別錯誤訊息裡面是否有「invalid escape」這段字串（可能依 JSON library 而定）
		msg := err.Error()
		if !regexp.MustCompile(`invalid escape`).MatchString(msg) {
			// 如果不是跳脫字元錯誤，就直接回傳
			return fmt.Errorf("第 %d 筆 JSON 解析失敗: %w", idx, err)
		}

		// 3) 要 sanitize 了
		sanitized := sanitizeEscape(raw)

		// 再用 sonic.Unmarshal 嘗試一次
		var e2 dmlEntry
		if err2 := sonic.Unmarshal([]byte(sanitized), &e2); err2 != nil {
			// sanitize 後還是失敗，紀錄並跳過
			fmt.Printf("第 %d 筆 sanitize 後仍解析失敗，略過: %s\n原始: %s\n", idx, err2, raw)
			continue
		}
		// sanitize 成功，再把結果存進 entries
		entries = append(entries, e2)
	}

	// 分類： Delete 、 Insert
	var deletes, inserts []dmlEntry

	for _, e := range entries {
		switch e.Action {
		case "Delete":
			deletes = append(deletes, e)
		case "Insert":
			inserts = append(inserts, e)
		default:
			// 忽略其他 Action
		}
	}

	/*------------------ Delete Start-----------------------*/
	// 開啟 transaction
	tx := p.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return fmt.Errorf("開啟 transaction 失敗：%w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if tx.Error != nil {
			// 如果中途有錯誤（tx.Error 被設過），就 rollback
			tx.Rollback()
		}
	}()

	// 執行 Delete (利用 p.getColumnTypesOnce 取 primary key )
	for idx, e := range deletes {
		tableName, ok := e.Data["TableName"].(string)
		if !ok || tableName == "" {
			return fmt.Errorf("第 %d 筆 Delete 未指定 TableName", idx)
		}

		// ① 從緩存或第一次查詢取得該表的 ColumnTypes
		columnTypes, err := p.getColumnTypesOnce(ctx, tableName)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("取得 %s 欄位資訊失敗: %w", tableName, err)
		}

		// ② 從 columnTypes 裡面篩出所有 PrimaryKey()
		cond := make(map[string]interface{})
		for _, ct := range columnTypes {
			isPK, _ := ct.PrimaryKey()
			if !isPK {
				continue
			}
			pkName := ct.Name()
			if val, exists := e.Data[pkName]; exists {
				cond[pkName] = val
			}
		}

		if len(cond) == 0 {
			tx.Rollback()
			return fmt.Errorf("第 %d 筆 Delete 找不到主鍵欄位或對應資料 (table=%s)", idx, tableName)
		}

		// ③ 執行刪除
		if err := tx.
			Table(tableName).
			Where(cond).
			Delete(nil).
			Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("刪除失敗 idx=%d, table=%s: %w", idx, tableName, err)
		}
	}

	// Commit
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit 失敗：%w", err)
	}
	/*------------------ Delete End-----------------------*/

	/*------------------ Insert Start-----------------------*/
	// 注意，Upsert 是在完全不同的 mssqlConn 連線，所以跟 Delete 不會綁一起
	// 把所有要 Insert 的依照 tableName 分組
	insertGroups := make(map[string][]map[string]interface{})

	for _, e := range inserts {
		tableName, ok := e.Data["TableName"].(string)
		if !ok || tableName == "" {
			tx.Rollback()
			return fmt.Errorf("插入前找不到 TableName")
		}

		// 把 "TableName" 欄位移除，其他欄位都留下來
		data := make(map[string]interface{}, len(e.Data))
		for k, v := range e.Data {
			if k == "TableName" {
				continue
			}
			if str, ok := v.(string); ok {
				data[k] = strings.ReplaceAll(str, `\n`, "\r\n")
			} else {
				data[k] = v
			}
		}

		insertGroups[tableName] = append(insertGroups[tableName], data)
	}

	// 針對每個 tableName，一次批次先寫 staging，再 MERGE，再清空 staging
	for tableName, datas := range insertGroups {
		if err := p.batchUpsertWithMerge(ctx, tableName, datas); err != nil {
			return fmt.Errorf(" Upsert 失敗 (table=%s): %w", tableName, err)
		}
	}
	/*------------------ Insert End-----------------------*/

	return nil

}

func (p *Processor) DdlLogProcess(ctx context.Context, body []byte) error {
	// 1. 解析 JSON 取得 XMLList
	var message model.DdlMessage
	if err := sonic.Unmarshal([]byte(body), &message); err != nil {
		return fmt.Errorf(" XML 解析失敗: %w", err)
	}

	// 2. 用迴圈處理每個 XML 串
	for _, xmlStr := range message.XMLList {
		// 2.1 解析單筆 XML
		var data model.DdlData
		if err := xml.Unmarshal([]byte(xmlStr), &data); err != nil {
			return fmt.Errorf(" XML 解析失敗: %w", err)
		}
		// 2.2 取出 DDL 語法，並去除多餘空白
		sqlText := strings.TrimSpace(data.EventData.Instance.TSQLCommand.CommandText)
		if sqlText == "" {
			return fmt.Errorf(" XML 未包含 CommandText")
		}
		normalized := strings.ToUpper(strings.Join(strings.Fields(sqlText), " "))
		hash := sha1.Sum([]byte(normalized))
		hashHex := hex.EncodeToString(hash[:])

		// 3. 檢查是否已執行過相同 DDL（DB + 記憶體）
		p.ddlMu.RLock()
		_, done := p.executedDDL[hashHex]
		p.ddlMu.RUnlock()

		if !done {
			var cnt int64
			if err := p.db.WithContext(ctx).
				Model(&model.ExecutedDDL{}).
				Where("SQLHash = ?", hashHex).
				Count(&cnt).Error; err != nil {
				return fmt.Errorf("查詢 DDL 執行紀錄失敗: %w", err)
			}
			if cnt > 0 {
				p.ddlMu.Lock()
				p.executedDDL[hashHex] = struct{}{}
				p.ddlMu.Unlock()
				done = true
			}
		}

		if done {
			continue
		}

		// 4. 在指定資料庫執行 DDL
		res := p.db.WithContext(ctx).Exec(sqlText)
		if err := res.Error; err != nil {
			// 把原始錯誤與 SQL 都印出來
			return fmt.Errorf("執行 DDL 失敗: %v; SQL: %s", err, sqlText)
		}

		// 5. 記錄此 DDL 已成功執行（寫入 DB + 快取）
		rec := model.ExecutedDDL{
			SQLHash: hashHex,
			SQLText: normalized,
		}
		if err := p.db.WithContext(ctx).Create(&rec).Error; err != nil {
			return fmt.Errorf("記錄 DDL 執行失敗: %w", err)
		}
		p.ddlMu.Lock()
		p.executedDDL[hashHex] = struct{}{}
		p.ddlMu.Unlock()
	}

	return nil
}
