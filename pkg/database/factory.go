package database

import (
	"fmt"
	"time"
	"websocket_broadcaster/internal/config"

	_ "github.com/go-sql-driver/mysql" // MySQL驱动
	_ "github.com/lib/pq"              // PostgreSQL驱动
)

// CreateMessageStore 根据配置创建消息存储实例
func CreateMessageStore(cfg *config.DatabaseConfig) (MessageStoreInterface, error) {
	switch cfg.Type {
	case "memory", "":
		return NewMemoryStore(), nil

	case "redis":
		addr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
		return NewRedisStore(addr, cfg.Redis.Password, cfg.Redis.DB)

	case "mysql":
		if cfg.MySQL.User == "" || cfg.MySQL.Database == "" {
			return nil, fmt.Errorf("MySQL配置不完整：需要用户名和数据库名")
		}
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4",
			cfg.MySQL.User, cfg.MySQL.Password, cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Database)
		return NewSQLStore("mysql", dsn)

	case "postgresql", "postgres":
		if cfg.PostgreSQL.User == "" || cfg.PostgreSQL.Database == "" {
			return nil, fmt.Errorf("PostgreSQL配置不完整：需要用户名和数据库名")
		}
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.PostgreSQL.Host, cfg.PostgreSQL.Port, cfg.PostgreSQL.User,
			cfg.PostgreSQL.Password, cfg.PostgreSQL.Database, cfg.PostgreSQL.SSLMode)
		return NewSQLStore("postgres", dsn)

	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", cfg.Type)
	}
}

// GetMessageTTL 获取消息TTL
func GetMessageTTL(cfg *config.DatabaseConfig) time.Duration {
	if cfg.MessageTTL <= 0 {
		return time.Hour // 默认1小时
	}
	return time.Duration(cfg.MessageTTL) * time.Second
}
