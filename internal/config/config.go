package config

import (
	"fmt"
	"os"
	"strings"
	"time"

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

// createDefaultConfig 创建默认配置
func createDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: "8765",
			Path: "/ws",
		},
		Database: DatabaseConfig{
			Type:       "memory",
			MessageTTL: 3600,
		},
		Groups: []BroadcastGroup{
			{
				Name:         "全平台互通",
				Members:      []string{"creative", "survival", "test_client", "debug_client", "qq_bot"},
				MessageTypes: []string{"chat"},
				Enabled:      true,
				Transform: &Transform{
					PrefixChat: "",
				},
			},
			{
				Name:         "事件广播",
				Members:      []string{"creative", "survival", "test_client", "qq_bot"},
				MessageTypes: []string{"event"},
				Enabled:      true,
				Transform: &Transform{
					PrefixEvent: "【事件】 ",
				},
			},
		},
		Rules: []BroadcastRule{
			{
				Name:         "监控转发",
				FromSources:  []string{"*"},
				ToTargets:    []string{"monitor_system"},
				MessageTypes: []string{"event"},
				Enabled:      false,
				Transform: &Transform{
					PrefixEvent: "[监控] ",
				},
			},
		},
		Clients: []ClientConfig{
			{
				Name:              "minecraft_server",
				URL:               "ws://localhost:8766/ws",
				AutoReconnect:     true,
				ReconnectInterval: 5,
			},
			{
				Name:              "qq_bot",
				URL:               "ws://localhost:8767/ws",
				AutoReconnect:     true,
				ReconnectInterval: 5,
			},
		},
	}
}

// writeDefaultConfig 将默认配置写入文件
func writeDefaultConfig(filename string) error {
	config := createDefaultConfig()

	// 转换为YAML格式
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("生成默认配置失败: %v", err)
	}

	// 添加配置文件头部注释
	header := `# WebSocket 广播器配置文件 (自动生成)
# 此文件由系统自动创建，包含推荐的默认配置
# 请根据您的实际需求修改相关设置

`

	// 写入文件
	fullData := header + string(data)
	err = os.WriteFile(filename, []byte(fullData), 0644)
	if err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// Load 加载配置文件
func Load(filename string) (*Config, error) {
	// 检查配置文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// 配置文件不存在，显示醒目提示
		separator := strings.Repeat("=", 60)
		fmt.Printf("\n%s\n", separator)
		fmt.Printf("[警告] 配置文件未找到: %s\n", filename)
		fmt.Printf("[信息] 正在创建默认配置文件...\n")

		// 创建默认配置文件
		if err := writeDefaultConfig(filename); err != nil {
			fmt.Printf("[错误] 创建默认配置文件失败: %v\n", err)
			fmt.Printf("%s\n\n", separator)
			return nil, fmt.Errorf("创建默认配置文件失败: %v", err)
		}

		fmt.Printf("[成功] 默认配置文件已创建: %s\n", filename)
		fmt.Printf("[信息] 请查看并根据需要修改配置，然后重新启动广播器\n")
		fmt.Printf("%s\n\n", separator)

		// 倒计时5秒
		for i := 5; i > 0; i-- {
			fmt.Printf("[倒计时] %d秒后退出...\r", i)
			time.Sleep(1 * time.Second)
		}
		fmt.Printf("\n[退出] 程序已退出，请重新启动\n")

		// 退出程序
		os.Exit(0)
	}

	// 配置文件存在，正常加载
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
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
