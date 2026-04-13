package monitor

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type AiModelMetricsCollector struct {
	requestCounters *prometheus.CounterVec
	errorCounters   *prometheus.CounterVec
	tokenCounters   *prometheus.CounterVec
	responseTimers  *prometheus.HistogramVec
	registry        *prometheus.Registry
	countersCache   sync.Map
	errorCache      sync.Map
	tokenCache      sync.Map
	timerCache      sync.Map
}

func NewAiModelMetricsCollector(registry *prometheus.Registry) *AiModelMetricsCollector {
	collector := &AiModelMetricsCollector{
		registry: registry,
		requestCounters: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ai_model_requests_total",
				Help: "AI模型总请求次数",
			},
			[]string{"user_id", "app_id", "model_name", "status"},
		),
		errorCounters: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ai_model_errors_total",
				Help: "AI模型错误次数",
			},
			[]string{"user_id", "app_id", "model_name", "error_message"},
		),
		tokenCounters: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ai_model_tokens_total",
				Help: "AI模型Token消耗总数",
			},
			[]string{"user_id", "app_id", "model_name", "token_type"},
		),
		responseTimers: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ai_model_response_duration_seconds",
				Help:    "AI模型响应时间",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"user_id", "app_id", "model_name"},
		),
	}

	registry.MustRegister(collector.requestCounters)
	registry.MustRegister(collector.errorCounters)
	registry.MustRegister(collector.tokenCounters)
	registry.MustRegister(collector.responseTimers)

	return collector
}

func (c *AiModelMetricsCollector) RecordRequest(userId, appId, modelName, status string) {
	c.requestCounters.WithLabelValues(userId, appId, modelName, status).Inc()
}

func (c *AiModelMetricsCollector) RecordError(userId, appId, modelName, errorMessage string) {
	c.errorCounters.WithLabelValues(userId, appId, modelName, errorMessage).Inc()
}

func (c *AiModelMetricsCollector) RecordTokenUsage(userId, appId, modelName, tokenType string, tokenCount float64) {
	c.tokenCounters.WithLabelValues(userId, appId, modelName, tokenType).Add(tokenCount)
}

func (c *AiModelMetricsCollector) RecordResponseTime(userId, appId, modelName string, duration time.Duration) {
	c.responseTimers.WithLabelValues(userId, appId, modelName).Observe(duration.Seconds())
}

func (c *AiModelMetricsCollector) RecordResponseTimeStart(userId, appId, modelName string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start)
		c.RecordResponseTime(userId, appId, modelName, duration)
	}
}
