package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 主配置结构
type Config struct {
	Server   ServerConfig     `yaml:"server"`
	Database DatabaseConfig   `yaml:"database"`
	Rules    []BroadcastRule  `yaml:"rules,omitempty"`
	Groups   []BroadcastGroup `yaml:"groups,omitempty"`
	Clients  []ClientConfig   `yaml:"clients,omitempty"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type       string      `yaml:"type"`        // 数据库类型: memory, redis, mysql, postgresql
	Redis      RedisConfig `yaml:"redis"`       // Redis配置
	MySQL      MySQLConfig `yaml:"mysql"`       // MySQL配置
	PostgreSQL PgSQLConfig `yaml:"postgresql"`  // PostgreSQL配置
	MessageTTL int         `yaml:"message_ttl"` // 消息TTL（秒）
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `yaml:"host"`     // Redis主机地址
	Port     int    `yaml:"port"`     // Redis端口
	Password string `yaml:"password"` // Redis密码
	DB       int    `yaml:"db"`       // Redis数据库编号
}

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Host     string `yaml:"host"`     // MySQL主机地址
	Port     int    `yaml:"port"`     // MySQL端口
	User     string `yaml:"user"`     // 用户名
	Password string `yaml:"password"` // 密码
	Database string `yaml:"database"` // 数据库名
}

// PgSQLConfig PostgreSQL配置
type PgSQLConfig struct {
	Host     string `yaml:"host"`     // PostgreSQL主机地址
	Port     int    `yaml:"port"`     // PostgreSQL端口
	User     string `yaml:"user"`     // 用户名
	Password string `yaml:"password"` // 密码
	Database string `yaml:"database"` // 数据库名
	SSLMode  string `yaml:"sslmode"`  // SSL模式
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	Path string `yaml:"path"`
}

// BroadcastRule 广播规则
type BroadcastRule struct {
	Name         string     `yaml:"name"`
	FromSources  []string   `yaml:"from_sources"`
	ToTargets    []string   `yaml:"to_targets"`
	MessageTypes []string   `yaml:"message_types"`
	Enabled      bool       `yaml:"enabled"`
	Transform    *Transform `yaml:"transform,omitempty"`
}

// BroadcastGroup 群组配置 - 简化多平台互通
type BroadcastGroup struct {
	Name         string     `yaml:"name"`
	Members      []string   `yaml:"members"`
	MessageTypes []string   `yaml:"message_types"`
	Enabled      bool       `yaml:"enabled"`
	Transform    *Transform `yaml:"transform,omitempty"`
}

// Transform 消息转换规则
type Transform struct {
	PrefixChat  string `yaml:"prefix_chat,omitempty"`
	PrefixEvent string `yaml:"prefix_event,omitempty"`
	ChangeFrom  string `yaml:"change_from,omitempty"`
}

// ClientConfig 客户端配置
type ClientConfig struct {
	Name              string `yaml:"name"`
	URL               string `yaml:"url"`
	AutoReconnect     bool   `yaml:"auto_reconnect"`
	ReconnectInterval int    `yaml:"reconnect_interval"`
}

// Load 加载配置文件
func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// GetServerAddr 获取服务器地址
func (c *Config) GetServerAddr() string {
	return c.Server.Host + ":" + c.Server.Port
}

// GetWebSocketURL 获取WebSocket完整URL
func (c *Config) GetWebSocketURL() string {
	return "ws://" + c.GetServerAddr() + c.Server.Path
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Port == "" {
		c.Server.Port = "8765"
	}
	if c.Server.Path == "" {
		c.Server.Path = "/ws"
	}

	// 数据库配置默认值
	if c.Database.Type == "" {
		c.Database.Type = "memory" // 默认使用内存存储
	}

	// Redis配置默认值
	if c.Database.Redis.Port == 0 {
		c.Database.Redis.Port = 6379
	}
	if c.Database.Redis.Host == "" {
		c.Database.Redis.Host = "localhost"
	}

	// MySQL配置默认值
	if c.Database.MySQL.Port == 0 {
		c.Database.MySQL.Port = 3306
	}
	if c.Database.MySQL.Host == "" {
		c.Database.MySQL.Host = "localhost"
	}

	// PostgreSQL配置默认值
	if c.Database.PostgreSQL.Port == 0 {
		c.Database.PostgreSQL.Port = 5432
	}
	if c.Database.PostgreSQL.Host == "" {
		c.Database.PostgreSQL.Host = "localhost"
	}
	if c.Database.PostgreSQL.SSLMode == "" {
		c.Database.PostgreSQL.SSLMode = "disable"
	}

	// 消息TTL默认值
	if c.Database.MessageTTL == 0 {
		c.Database.MessageTTL = 3600 // 默认1小时
	}

	return nil
}
