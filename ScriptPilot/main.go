package main

import (
	pb "ScriptPilot/proto/taskexecutor" // 替換為實際的 Protobuf 生成包路徑
	"ScriptPilot/util"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"

	"errors"
	"flag"

	"path/filepath"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	FactoryID      string `json:"FactoryId"`
	ScriptRootPath string `json:"ScriptRootPath"`
	TcpPort        int    `json:"TcpPort"`
}

// 緩存配置以避免重複讀取
var (
	config Config
	// configOnce sync.Once
	// configErr  error
)

func Init(env *string) {
	vp := util.CreateConfig("systemparameter", *env)
	fmt.Println(vp.AllSettings())
	config.FactoryID = vp.GetString("FactoryID")
	config.ScriptRootPath = vp.GetString("ScriptRootPath")
	config.TcpPort = vp.GetInt("TcpPort")
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
	// 組態設定載入
	// 變數名稱env ，啟動程式時沒有指定 -env，則程式中的 env 變數將會是 "dev"
	env := flag.String("env", "dev", "選擇環境 (dev, prod)") // flag.String：這個函數返回的是一個指標
	flag.Parse()

	// 載入組態設定
	Init(env)

	// 啟動 TCP 監聽
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(config.TcpPort))
	if err != nil {
		panic(fmt.Sprintf("無法啟動伺服器: %v", err))
	}

	// 建立 gRPC 伺服器
	ggrpcServer := grpc.NewServer()
	pb.RegisterTaskExecutorServer(ggrpcServer, &server{})
	reflection.Register(ggrpcServer) // 註冊反射服務（方便測試和調試）

	fmt.Printf("gRPC 伺服器啟動，監聽 :%d", config.TcpPort)
	if err := ggrpcServer.Serve(listener); err != nil {
		panic(fmt.Sprintf("伺服器啟動失敗: %v", err))
	}
}
