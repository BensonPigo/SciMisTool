package scilog

import (
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewFileLogger 建立一個依日期切檔的 Zap Logger
func NewFileLogger() (*zap.Logger, error) {
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

	logDir := "./logs"
	baseName := "ScriptPilot"
	pattern := filepath.Join(logDir, baseName+"-%Y-%m-%d.log")

	rotateWriter, err := rotatelogs.New(
		pattern,
		rotatelogs.WithLinkName(filepath.Join(logDir, baseName+".log")),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		return nil, err
	}

	ws := zapcore.AddSync(rotateWriter)
	level := zapcore.InfoLevel
	core := zapcore.NewCore(encoder, ws, level)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return logger, nil
}
