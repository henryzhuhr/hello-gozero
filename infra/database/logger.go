package database

import (
	"context"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormLogger wraps go-zero's logx to GORM's logger interface.
// It allows GORM to log SQL statements using go-zero's logging system.
type GormLogger struct {
	logger  logx.Logger
	level   logger.LogLevel
	slowSQL time.Duration
}

func NewGormLogger(l logx.Logger) *GormLogger {
	return &GormLogger{
		logger:  l,           // 因为封装了一层，所以跳过一层调用栈
		level:   logger.Warn, // or logger.Info if you want to log all SQL
		slowSQL: 200 * time.Millisecond,
	}
}

// LogMode 实现 GORM 的 LogMode 方法
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return &GormLogger{
		logger:  l.logger,
		level:   level,
		slowSQL: l.slowSQL,
	}
}

// Info logs info messages
func (l *GormLogger) Info(ctx context.Context, format string, args ...interface{}) {
	if l.level >= logger.Info {
		l.logger.WithContext(ctx).Infof(format, args...)
	}
}

// Warn logs warn messages
func (l *GormLogger) Warn(ctx context.Context, format string, args ...interface{}) {
	if l.level >= logger.Warn {
		l.logger.WithContext(ctx).Infof(format, args...)
	}
}

// Error logs error messages
func (l *GormLogger) Error(ctx context.Context, format string, args ...interface{}) {
	if l.level >= logger.Error {
		l.logger.WithContext(ctx).Errorf(format, args...)
	}
}

// Trace logs SQL with source, duration, etc.
// It implements GORM's Trace method [gorm.io/gorm/logger.Interface].
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && !isIgnorableError(err) && l.level >= logger.Error:
		l.logger.Errorf("SQL=%s, rows=%d, time=%v, error=%v", sql, rows, elapsed, err)
	case err != nil && isIgnorableError(err) && l.level >= logger.Info:
		// 将 "record not found" 降级为 info（或 warn），避免污染 error 日志
		l.logger.Infof("SQL=%s, rows=%d, time=%v, error=record not found", sql, rows, elapsed)
	case l.slowSQL != 0 && elapsed > l.slowSQL && l.level >= logger.Warn:
		l.logger.WithContext(ctx).Info("SLOW SQL=%s, rows=%d, time=%v", sql, rows, elapsed)
	case l.level >= logger.Info:
		l.logger.WithContext(ctx).Infof("SQL=%s, rows=%d, time=%v", sql, rows, elapsed)
	}
}

// isIgnorableError 判断是否为可忽略的错误。
// 例如 [gorm.ErrRecordNotFound] 不应该被视为错误日志记录的一部分。
// 目前仅包含 [gorm.ErrRecordNotFound]，未来可根据需要扩展
func isIgnorableError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, gorm.ErrRecordNotFound) ||
		errors.Is(err, gorm.ErrInvalidTransaction) || // 示例：按需添加
		// 未来可继续追加，比如自定义的业务 error
		false
}
