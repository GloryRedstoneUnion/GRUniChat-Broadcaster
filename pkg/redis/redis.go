package redis

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// EmbeddedRedis 内嵌Redis服务器
type EmbeddedRedis struct {
	port    int
	dataDir string
	cmd     *exec.Cmd
	client  *redis.Client
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewEmbeddedRedis 创建新的内嵌Redis实例
func NewEmbeddedRedis(port int, dataDir string) *EmbeddedRedis {
	ctx, cancel := context.WithCancel(context.Background())
	return &EmbeddedRedis{
		port:    port,
		dataDir: dataDir,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start 启动内嵌Redis服务器
func (r *EmbeddedRedis) Start() error {
	// 确保数据目录存在
	if err := os.MkdirAll(r.dataDir, 0755); err != nil {
		return fmt.Errorf("创建Redis数据目录失败: %v", err)
	}

	// 检查端口是否可用
	if !r.isPortAvailable(r.port) {
		return fmt.Errorf("端口 %d 已被占用", r.port)
	}

	// 获取Redis可执行文件路径
	redisServerPath, err := r.getRedisServerPath()
	if err != nil {
		return fmt.Errorf("获取Redis服务器路径失败: %v", err)
	}

	// 创建Redis配置
	configPath := filepath.Join(r.dataDir, "redis.conf")
	if err := r.createRedisConfig(configPath); err != nil {
		return fmt.Errorf("创建Redis配置失败: %v", err)
	}

	// 启动Redis服务器
	r.cmd = exec.CommandContext(r.ctx, redisServerPath, configPath)
	r.cmd.Dir = r.dataDir

	if err := r.cmd.Start(); err != nil {
		return fmt.Errorf("启动Redis服务器失败: %v", err)
	}

	// 等待Redis启动
	if err := r.waitForRedis(); err != nil {
		r.Stop()
		return fmt.Errorf("等待Redis启动失败: %v", err)
	}

	// 创建Redis客户端
	r.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%d", r.port),
		Password: "",
		DB:       0,
	})

	// 测试连接
	_, err = r.client.Ping(r.ctx).Result()
	if err != nil {
		r.Stop()
		return fmt.Errorf("Redis连接测试失败: %v", err)
	}

	return nil
}

// Stop 停止内嵌Redis服务器
func (r *EmbeddedRedis) Stop() error {
	if r.client != nil {
		r.client.Close()
	}

	if r.cmd != nil && r.cmd.Process != nil {
		r.cancel()

		// 尝试优雅关闭
		if err := r.cmd.Process.Signal(os.Interrupt); err != nil {
			// 如果优雅关闭失败，强制终止
			r.cmd.Process.Kill()
		}

		r.cmd.Wait()
	}

	return nil
}

// GetClient 获取Redis客户端
func (r *EmbeddedRedis) GetClient() *redis.Client {
	return r.client
}

// isPortAvailable 检查端口是否可用
func (r *EmbeddedRedis) isPortAvailable(port int) bool {
	address := fmt.Sprintf("localhost:%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	defer listener.Close()
	return true
}

// getRedisServerPath 获取Redis服务器可执行文件路径
func (r *EmbeddedRedis) getRedisServerPath() (string, error) {
	// 优先查找系统路径中的redis-server
	if path, err := exec.LookPath("redis-server"); err == nil {
		return path, nil
	}

	// 查找内嵌的Redis可执行文件
	execDir, err := os.Executable()
	if err != nil {
		return "", err
	}
	execDir = filepath.Dir(execDir)

	var redisServerName string
	switch runtime.GOOS {
	case "windows":
		redisServerName = "redis-server.exe"
	default:
		redisServerName = "redis-server"
	}

	// 在可执行文件同目录下查找
	embeddedPath := filepath.Join(execDir, "redis", redisServerName)
	if _, err := os.Stat(embeddedPath); err == nil {
		return embeddedPath, nil
	}

	// 在相对路径下查找
	relativePath := filepath.Join("redis", redisServerName)
	if _, err := os.Stat(relativePath); err == nil {
		return relativePath, nil
	}

	return "", fmt.Errorf("未找到Redis服务器可执行文件")
}

// createRedisConfig 创建Redis配置文件
func (r *EmbeddedRedis) createRedisConfig(configPath string) error {
	config := fmt.Sprintf(`# Redis内嵌配置
port %d
bind 127.0.0.1
dir %s
dbfilename dump.rdb
save 900 1
save 300 10
save 60 10000
maxmemory 128mb
maxmemory-policy allkeys-lru
appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec
# 禁用保护模式（仅本地使用）
protected-mode no
# 禁用日志输出到控制台
logfile ""
# 设置超时时间
timeout 0
# 启用键过期事件通知
notify-keyspace-events Ex
`, r.port, r.dataDir)

	return os.WriteFile(configPath, []byte(config), 0644)
}

// waitForRedis 等待Redis启动
func (r *EmbeddedRedis) waitForRedis() error {
	for i := 0; i < 30; i++ { // 最多等待30秒
		if r.isPortAvailable(r.port) {
			time.Sleep(1 * time.Second)
			continue
		}

		// 端口被占用说明Redis已启动
		time.Sleep(500 * time.Millisecond) // 再等待一点确保完全启动
		return nil
	}
	return fmt.Errorf("Redis启动超时")
}

// MessageStoreInterface 消息存储接口
type MessageStoreInterface interface {
	StoreMessage(messageID string, message []byte, ttl time.Duration) error
	GetMessage(messageID string) ([]byte, error)
	DeleteMessage(messageID string) error
	SetMessageStatus(messageID, status string, ttl time.Duration) error
	GetMessageStatus(messageID string) (string, error)
	IncrementCounter(key string) (int64, error)
	GetStats() (map[string]interface{}, error)
}

// MessageStore Redis消息存储接口
type MessageStore struct {
	client *redis.Client
	ctx    context.Context
}

// NewMessageStore 创建消息存储实例
func NewMessageStore(client *redis.Client) MessageStoreInterface {
	return &MessageStore{
		client: client,
		ctx:    context.Background(),
	}
}

// StoreMessage 存储消息
func (ms *MessageStore) StoreMessage(messageID string, message []byte, ttl time.Duration) error {
	key := fmt.Sprintf("msg:%s", messageID)
	return ms.client.Set(ms.ctx, key, message, ttl).Err()
}

// GetMessage 获取消息
func (ms *MessageStore) GetMessage(messageID string) ([]byte, error) {
	key := fmt.Sprintf("msg:%s", messageID)
	return ms.client.Get(ms.ctx, key).Bytes()
}

// DeleteMessage 删除消息
func (ms *MessageStore) DeleteMessage(messageID string) error {
	key := fmt.Sprintf("msg:%s", messageID)
	return ms.client.Del(ms.ctx, key).Err()
}

// SetMessageStatus 设置消息状态
func (ms *MessageStore) SetMessageStatus(messageID, status string, ttl time.Duration) error {
	key := fmt.Sprintf("status:%s", messageID)
	return ms.client.Set(ms.ctx, key, status, ttl).Err()
}

// GetMessageStatus 获取消息状态
func (ms *MessageStore) GetMessageStatus(messageID string) (string, error) {
	key := fmt.Sprintf("status:%s", messageID)
	return ms.client.Get(ms.ctx, key).Result()
}

// IncrementCounter 递增计数器
func (ms *MessageStore) IncrementCounter(key string) (int64, error) {
	return ms.client.Incr(ms.ctx, key).Result()
}

// GetStats 获取Redis统计信息
func (ms *MessageStore) GetStats() (map[string]interface{}, error) {
	info, err := ms.client.Info(ms.ctx, "memory", "keyspace", "stats").Result()
	if err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})
	stats["redis_info"] = info

	// 获取消息相关的键数量
	msgKeys, err := ms.client.Keys(ms.ctx, "msg:*").Result()
	if err == nil {
		stats["stored_messages"] = len(msgKeys)
	}

	statusKeys, err := ms.client.Keys(ms.ctx, "status:*").Result()
	if err == nil {
		stats["message_statuses"] = len(statusKeys)
	}

	return stats, nil
}

// MemoryStore 内存消息存储（降级方案）
type MemoryStore struct {
	messages map[string][]byte
	statuses map[string]string
	counters map[string]int64
	mutex    sync.RWMutex
}

// NewMemoryStore 创建内存存储实例
func NewMemoryStore() MessageStoreInterface {
	return &MemoryStore{
		messages: make(map[string][]byte),
		statuses: make(map[string]string),
		counters: make(map[string]int64),
	}
}

// StoreMessage 存储消息
func (ms *MemoryStore) StoreMessage(messageID string, message []byte, ttl time.Duration) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.messages[messageID] = message

	// 内存模式不实现TTL，仅简单存储
	return nil
}

// GetMessage 获取消息
func (ms *MemoryStore) GetMessage(messageID string) ([]byte, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	if msg, exists := ms.messages[messageID]; exists {
		return msg, nil
	}
	return nil, fmt.Errorf("消息不存在")
}

// DeleteMessage 删除消息
func (ms *MemoryStore) DeleteMessage(messageID string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	delete(ms.messages, messageID)
	return nil
}

// SetMessageStatus 设置消息状态
func (ms *MemoryStore) SetMessageStatus(messageID, status string, ttl time.Duration) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.statuses[messageID] = status
	return nil
}

// GetMessageStatus 获取消息状态
func (ms *MemoryStore) GetMessageStatus(messageID string) (string, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	if status, exists := ms.statuses[messageID]; exists {
		return status, nil
	}
	return "", fmt.Errorf("状态不存在")
}

// IncrementCounter 递增计数器
func (ms *MemoryStore) IncrementCounter(key string) (int64, error) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.counters[key]++
	return ms.counters[key], nil
}

// GetStats 获取统计信息
func (ms *MemoryStore) GetStats() (map[string]interface{}, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["type"] = "memory"
	stats["stored_messages"] = len(ms.messages)
	stats["message_statuses"] = len(ms.statuses)
	stats["counters"] = len(ms.counters)

	return stats, nil
}
