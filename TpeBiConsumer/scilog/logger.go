package scilog

import (
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewFileLogger 建立一個依「日期切檔」的 Zap Logger
func NewFileLogger() (*zap.Logger, error) {
	// 1. 定義 EncoderConfig（JSON 格式）
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:      "time",
		LevelKey:     "level",
		MessageKey:   "msg",
		CallerKey:    "caller",
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}
	encoder := zapcore.NewJSONEncoder(encoderCfg)

	// 2. 設定 rotatelogs，按照「每天」產生新檔
	//    ./logs/app-YYYY-MM-DD.log，並將最近 7 天的檔案保留，
	//    並且在 ./logs/app.log 指向當前的檔名
	logDir := "./logs"
	baseName := "TpeBiConsumer"
	// 使用時間格式字串：%Y-%m-%d 代表年月日
	pattern := filepath.Join(logDir, baseName+"-%Y-%m-%d.log")

	// 旋轉日誌的關鍵引數
	rotateWriter, err := rotatelogs.New(
		pattern,
		rotatelogs.WithLinkName(filepath.Join(logDir, baseName+".log")), // 建立一條軟連結 -> 當前檔案
		rotatelogs.WithMaxAge(7*24*time.Hour),                           // 最多保留 7 天
		rotatelogs.WithRotationTime(24*time.Hour),                       // 每 24 小時切檔一次
	)
	if err != nil {
		return nil, err
	}

	// 3. 建立 zapcore，將輸出導向 rotateWriter
	ws := zapcore.AddSync(rotateWriter)
	// 假設最低只記錄 Info 級別以上
	level := zapcore.InfoLevel
	core := zapcore.NewCore(encoder, ws, level)

	// 4. 最後建立 Logger，並加上呼叫者資訊與錯誤堆疊
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return logger, nil
}
