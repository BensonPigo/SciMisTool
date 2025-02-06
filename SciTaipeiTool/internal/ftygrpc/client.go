package ftygrpc

// 負責 gRPC 的client端連線與調用邏輯。包含連接 gRPC Server 的邏輯，提供封裝的調用方法

import (
	"SciTaipeiTool/proto/taskexecutor"
	pb "SciTaipeiTool/proto/taskexecutor"
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client 是 gRPC 客戶端的封裝結構
type Client struct {
	conn      *grpc.ClientConn
	client    pb.TaskExecutorClient
	FactoryId string
}

// NewClient 創建並返回一個 gRPC Client
func NewClient(target string, timeout time.Duration) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 使用 grpc.DialContext (雖然是 Deprecated，但仍可用於目前的需求)
	conn, err := grpc.DialContext(ctx, target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewTaskExecutorClient(conn)

	return &Client{conn: conn, client: client}, nil
}

// Close 關閉 gRPC 連線
func (c *Client) Close() error {
	return c.conn.Close()
}

// 調用 TaskExecutor 的 ExecuteTask 方法
func (c *Client) ExecuteTask(ctx context.Context, factoryId string, taskName string) (string, error) {
	req := &pb.TaskRequest{
		FactoryId: factoryId,
		TaskName:  taskName,
	}

	res, err := c.client.ExecuteTask(ctx, req)
	if err != nil {
		return "", err
	}
	return res.Message, nil
}

// 調用 TaskExecutor 的 GetScripts 方法
func (c *Client) GetScripts(ctx context.Context) (*taskexecutor.GetScriptsResponse, error) {
	req := &pb.Empty{}

	res, err := c.client.GetScripts(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
