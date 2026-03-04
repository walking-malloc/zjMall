package database

import (
	"database/sql"
	"time"

	"zjMall/internal/common/metrics"

	"gorm.io/gorm"
)

const slowQueryThreshold = 1 * time.Second // 慢查询阈值：1秒

// InitDatabaseMetrics 初始化数据库监控
// 需要在 InitMySQL 之后调用
func InitDatabaseMetrics(db *gorm.DB) error {
	// 注册 GORM 回调来监控查询
	db.Callback().Query().Before("gorm:query").Register("prometheus:before_query", beforeQuery)
	db.Callback().Query().After("gorm:query").Register("prometheus:after_query", afterQuery)
	db.Callback().Create().Before("gorm:create").Register("prometheus:before_create", beforeCreate)
	db.Callback().Create().After("gorm:create").Register("prometheus:after_create", afterCreate)
	db.Callback().Update().Before("gorm:update").Register("prometheus:before_update", beforeUpdate)
	db.Callback().Update().After("gorm:update").Register("prometheus:after_update", afterUpdate)
	db.Callback().Delete().Before("gorm:delete").Register("prometheus:before_delete", beforeDelete)
	db.Callback().Delete().After("gorm:delete").Register("prometheus:after_delete", afterDelete)

	// 启动定期更新连接池指标
	go updateConnectionPoolMetrics(db)

	return nil
}

// updateConnectionPoolMetrics 定期更新连接池指标
func updateConnectionPoolMetrics(db *gorm.DB) {
	ticker := time.NewTicker(10 * time.Second) // 每10秒更新一次
	defer ticker.Stop()

	for range ticker.C {
		sqlDB, err := db.DB()
		if err != nil {
			continue
		}

		stats := sqlDB.Stats()
		metrics.DatabaseConnections.WithLabelValues("open").Set(float64(stats.OpenConnections))
		metrics.DatabaseConnections.WithLabelValues("in_use").Set(float64(stats.InUse))
		metrics.DatabaseConnections.WithLabelValues("idle").Set(float64(stats.Idle))
	}
}

// beforeQuery 查询前回调
func beforeQuery(db *gorm.DB) {
	db.InstanceSet("prometheus_start_time", time.Now())
}

// afterQuery 查询后回调
func afterQuery(db *gorm.DB) {
	startTime, ok := db.InstanceGet("prometheus_start_time")
	if !ok {
		return
	}

	duration := time.Since(startTime.(time.Time)).Seconds()
	metrics.DatabaseQueryDuration.WithLabelValues("select").Observe(duration)

	// 记录慢查询
	if duration > slowQueryThreshold.Seconds() {
		metrics.DatabaseSlowQueries.WithLabelValues("select").Inc()
	}
}

// beforeCreate 创建前回调
func beforeCreate(db *gorm.DB) {
	db.InstanceSet("prometheus_start_time", time.Now())
}

// afterCreate 创建后回调
func afterCreate(db *gorm.DB) {
	startTime, ok := db.InstanceGet("prometheus_start_time")
	if !ok {
		return
	}

	duration := time.Since(startTime.(time.Time)).Seconds()
	metrics.DatabaseQueryDuration.WithLabelValues("insert").Observe(duration)

	if duration > slowQueryThreshold.Seconds() {
		metrics.DatabaseSlowQueries.WithLabelValues("insert").Inc()
	}
}

// beforeUpdate 更新前回调
func beforeUpdate(db *gorm.DB) {
	db.InstanceSet("prometheus_start_time", time.Now())
}

// afterUpdate 更新后回调
func afterUpdate(db *gorm.DB) {
	startTime, ok := db.InstanceGet("prometheus_start_time")
	if !ok {
		return
	}

	duration := time.Since(startTime.(time.Time)).Seconds()
	metrics.DatabaseQueryDuration.WithLabelValues("update").Observe(duration)

	if duration > slowQueryThreshold.Seconds() {
		metrics.DatabaseSlowQueries.WithLabelValues("update").Inc()
	}
}

// beforeDelete 删除前回调
func beforeDelete(db *gorm.DB) {
	db.InstanceSet("prometheus_start_time", time.Now())
}

// afterDelete 删除后回调
func afterDelete(db *gorm.DB) {
	startTime, ok := db.InstanceGet("prometheus_start_time")
	if !ok {
		return
	}

	duration := time.Since(startTime.(time.Time)).Seconds()
	metrics.DatabaseQueryDuration.WithLabelValues("delete").Observe(duration)

	if duration > slowQueryThreshold.Seconds() {
		metrics.DatabaseSlowQueries.WithLabelValues("delete").Inc()
	}
}

// GetConnectionPoolStats 获取连接池统计信息（用于调试）
func GetConnectionPoolStats(db *gorm.DB) (*sql.DBStats, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	stats := sqlDB.Stats()
	return &stats, nil
}
