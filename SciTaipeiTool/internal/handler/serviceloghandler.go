package handler

import (
	"context"
	"net/http"
	"time"

	"SciTaipeiTool/internal/ftygrpc"
	pb "SciTaipeiTool/proto/taskexecutor"
)

type ServiceLogHandler struct {
	GRpcClients []*ftygrpc.Client
}

func (h *ServiceLogHandler) GetServiceLog(w http.ResponseWriter, r *http.Request) {
	serviceName := r.URL.Query().Get("serviceName")
	rawDate := r.URL.Query().Get("logDate")
	factoryID := r.URL.Query().Get("factoryId")

	if factoryID == "" {
		http.Error(w, "missing factoryId", http.StatusBadRequest)
		return
	}

	logDate, err := time.Parse("2006-01-02", rawDate)
	if err != nil {
		http.Error(w, "logDate \u683c\u5f0f\u932f\u8aa4", http.StatusBadRequest)
		return
	}

	var client *ftygrpc.Client
	for _, c := range h.GRpcClients {
		if c.FactoryId == factoryID {
			client = c
			break
		}
	}
	if client == nil {
		http.Error(w, "invalid factoryId", http.StatusBadRequest)
		return
	}

	// Call gRPC
	req := &pb.GetServiceLogRequest{
		ServiceName: serviceName,
		Date:        logDate.Format("2006-01-02"),
	}
	resp, err := client.GetServiceLog(context.Background(), req)
	if err != nil {
		http.Error(w, "gRPC \u547c\u53eb\u5931\u6557: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	// 直接寫出 JSON 內容
	w.Write([]byte(resp.LogContent))
}
