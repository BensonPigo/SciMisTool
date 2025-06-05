package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	ProcessRuns = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "process_runs_total",
			Help: "批次執行次數",
		},
		[]string{"type"}, // type: ddl or dml
	)
	ProcessErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "process_errors_total",
			Help: "批次錯誤次數",
		},
		[]string{"type"},
	)
	ProcessDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "process_duration_seconds",
			Help:    "批次處理耗時（秒）",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"type"},
	)
)

func init() {
	prometheus.MustRegister(ProcessRuns, ProcessErrors, ProcessDuration)
}
