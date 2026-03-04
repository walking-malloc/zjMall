package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ============================================
// HTTP 请求指标（高优先级）
// ============================================

// HTTPRequestsTotal HTTP 请求总数（Counter）
var HTTPRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	},
	[]string{"method", "path", "status"}, // 标签：请求方法、路径、状态码
)

// HTTPRequestDuration HTTP 请求耗时（Histogram）
var HTTPRequestDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	},
	[]string{"method", "path"},
)

// ============================================
// 数据库指标（高优先级）
// ============================================

// DatabaseConnections 数据库连接数（Gauge）
var DatabaseConnections = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "database_connections",
		Help: "Number of database connections",
	},
	[]string{"state"}, // 标签：idle, in_use, open
)

// DatabaseQueryDuration 数据库查询耗时（Histogram）
var DatabaseQueryDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "database_query_duration_seconds",
		Help:    "Database query duration in seconds",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	},
	[]string{"operation"}, // 标签：select, insert, update, delete
)

// DatabaseSlowQueries 慢查询数（Counter）
var DatabaseSlowQueries = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "database_slow_queries_total",
		Help: "Total number of slow queries",
	},
	[]string{"operation"}, // 标签：select, insert, update, delete
)

// ============================================
// Redis 指标（高优先级）
// ============================================

// RedisOperationsTotal Redis 操作总数（Counter）
var RedisOperationsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "redis_operations_total",
		Help: "Total number of Redis operations",
	},
	[]string{"operation", "status"}, // 标签：操作类型（get, set, del等）、状态（success, error）
)

// RedisOperationDuration Redis 操作耗时（Histogram）
var RedisOperationDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "redis_operation_duration_seconds",
		Help:    "Redis operation duration in seconds",
		Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
	},
	[]string{"operation"},
)

// RedisCacheHits Redis 缓存命中数（Counter）
var RedisCacheHits = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "redis_cache_hits_total",
		Help: "Total number of Redis cache hits",
	},
)

// RedisCacheMisses Redis 缓存未命中数（Counter）
var RedisCacheMisses = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "redis_cache_misses_total",
		Help: "Total number of Redis cache misses",
	},
)

// ============================================
// 业务指标（高优先级）
// ============================================

// OrdersCreatedTotal 订单创建总数（Counter）
var OrdersCreatedTotal = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "orders_created_total",
		Help: "Total number of orders created",
	},
)

// OrdersCreatedValue 订单总金额（Counter）
var OrdersCreatedValue = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "orders_created_value_total",
		Help: "Total value of orders created",
	},
)

// PaymentsSuccessTotal 支付成功数（Counter）
var PaymentsSuccessTotal = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "payments_success_total",
		Help: "Total number of successful payments",
	},
)

// PaymentsFailedTotal 支付失败数（Counter）
var PaymentsFailedTotal = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "payments_failed_total",
		Help: "Total number of failed payments",
	},
)

// PaymentsValue 支付总金额（Counter）
var PaymentsValue = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "payments_value_total",
		Help: "Total value of payments",
	},
)
