package main

import (
	pb "ScriptPilot/proto/taskexecutor" // 替換為實際的 Protobuf 生成包路徑
	"ScriptPilot/scilog"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"

	"errors"

	"path/filepath"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	FactoryID       string            `json:"FactoryId"`
	ScriptRootPath  string            `json:"ScriptRootPath"`
	TcpPort         int               `json:"TcpPort"`
	ServiceLogPaths map[string]string `json:"ServiceLogPaths"`
}

// 緩存配置以避免重複讀取
var (
	config Config
	// configOnce sync.Once
	// configErr  error
)

func loadConfig() (*Config, error) {
	matches, err := filepath.Glob("config/config.*.json")
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, errors.New("在 config/ 目錄下找不到任何 config.*.json 檔案")
	}
	cfgPath := matches[0]
	file, err := os.Open(cfgPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// server 結構實現 TaskExecutorServer 介面
type server struct {
	pb.UnimplementedTaskExecutorServer
}

// ExecuteTask 方法實現
func (s *server) ExecuteTask(ctx context.Context, req *pb.TaskRequest) (*pb.TaskResponse, error) {
	factoryId := req.FactoryId
	taskName := req.TaskName

	// 取得腳本
	scriptPath, err := getScriptPath(taskName, factoryId)
	if err != nil {
		return &pb.TaskResponse{
			Message: "找不到對應的腳本，請檢查TaskName參數以及腳本檔案",
			Error:   err.Error(),
		}, nil
	}

	// 腳本執行
	err = executePowerShellScript(scriptPath)
	if err != nil {
		return &pb.TaskResponse{
			Message: "腳本執行失敗",
			Error:   err.Error(),
		}, nil
	}

	return &pb.TaskResponse{
		Message: "腳本執行成功",
	}, nil
}

// 取得所有腳本
func (s *server) GetScripts(ctx context.Context, empty *pb.Empty) (*pb.GetScriptsResponse, error) {

	var ps1Files []string
	// 使用 filepath.Walk 遞迴遍歷目錄
	filepath.Walk(config.ScriptRootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// 如果訪問某個檔案或目錄失敗，返回錯誤
			return err
		}

		// 檢查是否是檔案並且副檔名是 .ps1
		if !info.IsDir() && filepath.Ext(path) == ".ps1" {
			ps1Files = append(ps1Files, path)
		}
		return err
	})

	return &pb.GetScriptsResponse{
		FactoryId:   config.FactoryID,
		ScriptFiles: ps1Files,
	}, nil
}

// GetServiceLog fetches log content for a specific service and date
func (s *server) GetServiceLog(ctx context.Context, req *pb.GetServiceLogRequest) (*pb.GetServiceLogResponse, error) {
	dir, ok := config.ServiceLogPaths[strings.ToLower(req.ServiceName)]
	if !ok {
		return nil, fmt.Errorf("service not found")
	}

	cleaned := strings.ReplaceAll(req.Date, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "/", "")
	t, err := time.Parse("20060102", cleaned)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %v", err)
	}
	normalized := t.Format("2006-01-02")

	filePath := dir + normalized + ".log"
	data, err := LogsToJSONArray(filePath)
	if err != nil {
		// return nil, err
		return &pb.GetServiceLogResponse{LogContent: string("")}, nil
	}
	return &pb.GetServiceLogResponse{LogContent: string(data)}, nil
}

func LogsToJSONArray(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 用 Decoder 一行一物件解析
	dec := sonic.ConfigFastest.NewDecoder(bytes.NewReader(data))
	var arr []interface{}
	for {
		// 跳過可能的空白
		if !dec.More() {
			break
		}
		var obj interface{}
		if err := dec.Decode(&obj); err != nil {
			return nil, err
		}
		arr = append(arr, obj)
	}

	// 最後一次性序列化成 JSON Array
	return json.Marshal(arr)
}

// 根據 taskName 獲取腳本路徑
func getScriptPath(taskName string, factoryId string) (string, error) {

	// 載入配置
	if config.FactoryID != factoryId {
		return "", errors.New("FactoryID 不符: " + config.FactoryID)
	}

	scriptFileName := config.ScriptRootPath + "\\" + taskName //+ ".ps1"

	return scriptFileName, nil
}

// 執行 PowerShell 腳本
func executePowerShellScript(scriptPath string) error {
	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	return cmd.Start()
}

func main() {
	logger, err := scilog.NewFileLogger()
	if err != nil {
		panic(fmt.Sprintf("無法建立日誌：%v", err))
	}
	defer logger.Sync()
	sugar := logger.Sugar()
	sugar.Info("Logger 初始化成功")

	cfg, err := loadConfig()
	if err != nil {
		sugar.Fatalf("無法載入組態: %v", err)
	}
	config = *cfg

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(config.TcpPort))
	if err != nil {
		sugar.Fatalf("無法啟動伺服器: %v", err)
	}

	// 建立 gRPC 伺服器
	ggrpcServer := grpc.NewServer()
	pb.RegisterTaskExecutorServer(ggrpcServer, &server{})
	reflection.Register(ggrpcServer) // 註冊反射服務（方便測試和調試）

	sugar.Infof("gRPC 伺服器啟動，監聽 :%d", config.TcpPort)
	if err := ggrpcServer.Serve(listener); err != nil {
		sugar.Fatalf("伺服器啟動失敗: %v", err)
	}
}
