package main

// HTTP 後端服務入口

import (
	"SciTaipeiTool/internal/auth"
	"SciTaipeiTool/internal/ftygrpc"
	"SciTaipeiTool/internal/handler"
	"SciTaipeiTool/middleware"
	"SciTaipeiTool/scilog"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"SciTaipeiTool/internal/dataprovider/db"
	"SciTaipeiTool/internal/dataprovider/models"

	"github.com/gin-gonic/gin"
)

type Config struct {
	Host string `json:"host"`
	Port string `json:"port"`
	DB   struct {
		Host     string `json:"host"`
		User     string `json:"user"`
		Password string `json:"password"`
		Name     string `json:"name"`
	} `json:"db"`
	Grpcs []struct {
		Factory string `json:"factory"`
		Server  string `json:"server"`
		Timeout int    `json:"timeout"`
	} `json:"grpcservers"`
	Jwt_secret_key string `json:"jwt_secret_key"`
}

// 載入Config配置
func loadConfig() (*Config, error) {
	// 自動尋找 config 目錄下第一個 config.*.json
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

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

// 捕捉關閉信號並釋放資源
func waitForShutdown() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	<-shutdown // 等待信號
	log.Println("正在關閉伺服器...")

	// 關閉資料庫連接
	sqlDB, err := db.GetDB().DB()
	if err == nil {
		sqlDB.Close()
		log.Println("資料庫連接已關閉")
	}

	log.Println("伺服器已關閉")
}

func main() {

	// 建立 Logger
	logger, err := scilog.NewFileLogger()
	if err != nil {
		// 如果 Logger 建立失敗，立即中止程式並印出錯誤
		panic(fmt.Sprintf("無法建立日誌：%v", err))
	}
	// 確保程式結束前會把緩衝區 flush
	defer logger.Sync()

	// 建立 Sugared Logger，方便後續呼叫
	sugar := logger.Sugar()
	sugar.Info("Logger 初始化成功")

	// 組態設定載入
	config, err := loadConfig()

	if err != nil {
		sugar.Fatalf("無法載入組態: %v", err)
	}

	/*----------初始化----------*/
	// 資料庫
	db.InitDB(config.DB.Name)
	dbConn := db.GetDB()

	// 僅在啟動時進行資料表遷移
	models.Migrate(dbConn)

	// 初始化 gRPC Client
	var gRpcClients []*ftygrpc.Client = make([]*ftygrpc.Client, 0)

	for _, grpc := range config.Grpcs {

		gRpcClient, err := ftygrpc.NewClient(grpc.Server, time.Duration(grpc.Timeout)*time.Second)
		if err != nil {
			sugar.Fatalf("Failed to initialize gRPC client: %v", err)
		}
		gRpcClient.FactoryId = grpc.Factory
		gRpcClients = append(gRpcClients, gRpcClient)
	}
	defer func() {
		// gRpcClient.Close()

		for _, client := range gRpcClients {
			client.Close()
		}
	}()
	if len(gRpcClients) == 0 {
		sugar.Fatalf("無法載入任何gRPC Client")

		return
	}
	// 密鑰
	auth.Init(config.Jwt_secret_key)

	/*----------初始化----------*/

	// 以下為Web，創建一個新的 Gin 實例
	router := gin.Default()

	// CORS(跨來源資源共享)設定，允許來自前端的請求：
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // 修改為特定域名提高安全性
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST,PATCH, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 註冊路由

	// 公開路由
	lh := &handler.LoginHandler{DatabaseName: config.DB.Name}
	router.POST("/api/v1/users/Login", lh.Login)
	router.POST("/api/v1/users/Register", lh.Register)
	router.PATCH("/api/v1/users/ResetPassword", lh.ResetPassword)
	router.POST("/api/v1/users/RefreshToken", lh.RefreshToken)

	// 受保護的路由
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware())
	{
		h := &handler.ExecuteTaskHandler{GRpcClients: gRpcClients}
		protected.POST("/ExecuteTask", h.ExecuteTask)
		protected.GET("/GetScripts", h.GetScripts)
		protected.POST("/users/Logout", lh.Logout)

		slh := &handler.ServiceLogHandler{GRpcClients: gRpcClients}
		router.GET("/api/v1/service/log", gin.WrapF(slh.GetServiceLog))
	}

	// 啟動伺服器
	go func() {
		// 監聽 :8080
		if err := router.Run(config.Host + ":" + config.Port); err != nil {
			sugar.Fatalf("伺服器啟動失敗: %v", err)
		}
	}()

	// 捕捉關閉信號
	waitForShutdown()
}
