package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ProcessedMetrics = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pethealth_processed_metrics_total",
		Help: "The total number of processed pet metrics",
	}, []string{"shard", "status"})

	DatabaseRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "pethealth_db_request_duration_seconds",
		Help:    "Histogram of response latency for database requests",
		Buckets: prometheus.DefBuckets,
	}, []string{"shard", "operation"})
)

func ObserveDBQuery(shardName string, operation string, f func() error) error {
	start := time.Now()
	err := f()
	duration := time.Since(start).Seconds()
	DatabaseRequestDuration.WithLabelValues(shardName, operation).Observe(duration)

	if err != nil {
		ProcessedMetrics.WithLabelValues(shardName, "error").Inc()
	} else {
		ProcessedMetrics.WithLabelValues(shardName, "success").Inc()
	}
	return err
}
