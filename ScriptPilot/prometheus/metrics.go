package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// 定義 Prometheus 的指標
var (
	LogCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "log_events_total",
			Help: "Total number of log events",
		},
		[]string{"level", "module"}, // 記錄 log 等級和模組名稱
	)
)

func InitMetrics() {
	// 將指標註冊到 Prometheus 默認的 Registry
	prometheus.MustRegister(LogCounter)
}
