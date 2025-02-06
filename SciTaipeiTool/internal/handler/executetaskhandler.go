package handler

// 定義 HTTP API 的處理方法（路由和參數解析）。定義每個 HTTP API 的行為，將請求轉發給業務邏輯層。
import (
	"SciTaipeiTool/internal/ftygrpc"
	pb "SciTaipeiTool/proto/taskexecutor"
	"sync"
	"time"

	"net/http"

	"context"

	"github.com/gin-gonic/gin"
)

// Handler 是 HTTP 請求處理器
type ExecuteTaskHandler struct {
	GRpcClients []*ftygrpc.Client
	// GRpcClient *ftygrpc.Client
}

type Task struct {
	FactoryID string
	TaskNames []string
}

// ExecuteTaskHandler 處理 /ExecuteTask 的 HTTP POST 請求
func (h *ExecuteTaskHandler) ExecuteTask(c *gin.Context) {
	// var req pb.TaskRequest

	var tasks []Task

	// 解析並驗證 POST 請求的 JSON，然後塞進 req
	if err := c.ShouldBindJSON(&tasks); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	// 檢查參數
	if len(tasks) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameters cant't be empty."})
		return
	}

	var reqList []*pb.TaskRequest = make([]*pb.TaskRequest, 0)
	// 整理需求清單
	for _, task := range tasks {
		for _, s := range task.TaskNames {
			reqList = append(reqList, &pb.TaskRequest{
				FactoryId: task.FactoryID,
				TaskName:  s,
			})
		}
	}

	// 從 gin.Context 獲取標準的 context.Context，並設置超時
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if len(h.GRpcClients) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No available gRPC clients"})
		return
	}

	// 使用WaitGroup處理非同步
	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []string
	var errors []string

	// 遍歷 grpc client
	for _, client := range h.GRpcClients {
		wg.Add(1)
		go func(c *ftygrpc.Client) {
			defer wg.Done()
			processGRPCRequests(ctx, c, reqList, &mu, &results, &errors)
		}(client)
	}

	wg.Wait()

	// 如果所有請求都失敗，則回傳錯誤
	if len(results) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "All gRPC requests failed",
			"details": errors,
		})
		return
	}

	// 返回成功的 JSON 響應
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": results,
	})
}

func processGRPCRequests(ctx context.Context, client *ftygrpc.Client, reqList []*pb.TaskRequest, mu *sync.Mutex, results *[]string, errors *[]string) {
	for _, req := range reqList {
		if req.FactoryId == client.FactoryId {
			result, err := client.ExecuteTask(ctx, req.FactoryId, req.TaskName)

			// 加鎖後僅在需要時更新
			mu.Lock()
			if err != nil {
				*errors = append(*errors, err.Error())
			} else {
				*results = append(*results, result)
			}
			mu.Unlock()
		}
	}
}

func (h *ExecuteTaskHandler) GetScripts(c *gin.Context) {

	// 從 gin.Context 獲取標準的 context.Context，並設置超時
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if len(h.GRpcClients) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No available gRPC clients"})
		return
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []gin.H
	var errors []string

	for _, client := range h.GRpcClients {
		wg.Add(1)
		go func(client *ftygrpc.Client) {
			defer wg.Done()
			result, err := client.GetScripts(ctx)
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errors = append(errors, err.Error())
			} else if result != nil {
				// 將需要的欄位提取出來
				results = append(results, gin.H{
					"FactoryId":   result.FactoryId,
					"ScriptFiles": result.ScriptFiles,
				})
			}
		}(client)
	}

	wg.Wait()
	// var aa [2]string = [2]string{"script1.ps1", "script2.ps1"}
	// results = append(results, gin.H{
	// 	"FactoryId":   "PH2",
	// 	"ScriptFiles": aa,
	// })

	if len(results) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "All gRPC requests failed",
			"details": errors,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"results": results,
	})
}
