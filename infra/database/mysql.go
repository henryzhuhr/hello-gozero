package database

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Mysql 配置结构体
type MysqlConfig struct {
	// 地址
	Host string `json:"Host"`

	// 端口
	Port int `json:"Port"`

	// 用户名
	User string `json:"User"`

	// 密码
	Password string `json:"Password"`

	// 数据库名称
	DB string `json:"DB"`

	// 最大打开连接数
	MaxOpenConns int `json:"MaxOpenConns"`

	// 最大空闲连接数
	MaxIdleConns int `json:"MaxIdleConns"`

	// 连接最大生命周期，单位秒
	ConnMaxLifetime int `json:"ConnMaxLifetime"`

	// 连接最大空闲时间，单位秒
	ConnMaxIdleTime int `json:"ConnMaxIdleTime"`
}

// MustNewMysql 初始化 MySQL 连接，失败时 panic
func MustNewMysql(config MysqlConfig, appLogger logx.Logger) *gorm.DB {
	// 初始化 Gorm 日志，接管 go-zero 日志
	gormLogger := NewGormLogger(appLogger)

	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DB,
	)
	db, err := gorm.Open(mysql.Open(dataSource), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		panic(err)
	}

	// 获取底层 SQL DB 实例
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		panic(err)
	}

	// 配置连接池
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(config.ConnMaxIdleTime) * time.Second)

	return db
}

// CloseMysql 关闭 MySQL 连接
func CloseMysql(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
