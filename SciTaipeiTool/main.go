package main

// HTTP 後端服務入口

import (
	"SciTaipeiTool/internal/auth"
	"SciTaipeiTool/internal/ftygrpc"
	"SciTaipeiTool/internal/handler"
	"SciTaipeiTool/middleware"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"SciTaipeiTool/internal/dataprovider/db"
	"SciTaipeiTool/internal/dataprovider/models"

	"flag"

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
func loadConfig(env string) (*Config, error) {
	file, err := os.Open("config/config." + env + ".json")
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
	// 組態設定載入
	// 變數名稱env ，啟動程式時沒有指定 -env，則程式中的 env 變數將會是 "dev"
	env := flag.String("env", "dev", "選擇環境 (dev, prod)") // flag.String：這個函數返回的是一個指標
	flag.Parse()

	config, err := loadConfig(*env)

	if err != nil {
		log.Fatalf("無法載入組態: %v", err)
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
			log.Fatalf("Failed to initialize gRPC client: %v", err)
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

	}

	// 啟動伺服器
	go func() {
		// 監聽 :8080
		if err := router.Run(config.Host + ":" + config.Port); err != nil {
			log.Fatalf("伺服器啟動失敗: %v", err)
		}
	}()

	// 捕捉關閉信號
	waitForShutdown()
}
