package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"SciTaipeiTool/internal/ftygrpc"
	pb "SciTaipeiTool/proto/taskexecutor"
)

type ServiceLogHandler struct {
	Client *ftygrpc.Client
}

func (h *ServiceLogHandler) GetServiceLog(w http.ResponseWriter, r *http.Request) {
	factoryID := r.URL.Query().Get("factoryID")
	serviceName := r.URL.Query().Get("serviceName")
	rawDate := r.URL.Query().Get("logDate")

	logDate, err := time.Parse("2006-01-02", rawDate)
	if err != nil {
		http.Error(w, "logDate \u683c\u5f0f\u932f\u8aa4", http.StatusBadRequest)
		return
	}

	req := &pb.GetServiceLogRequest{
		ServiceName: serviceName,
		Date:        logDate.String(),
	}
	resp, err := h.Client.GetServiceLog(context.Background(), req)
	if err != nil {
		http.Error(w, "gRPC \u547c\u53eb\u5931\u6557: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if resp.GetLogContent() == "" {
		http.Error(w, "\u627e\u4e0d\u5230\u5c0d\u61c9\u7684 log", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"serviceName": serviceName,
		"logDate":     logDate.Format("2006-01-02"),
		"content":     resp.GetLogContent(),
	})
}
