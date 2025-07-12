package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// MessageStoreInterface 消息存储接口
type MessageStoreInterface interface {
	StoreMessage(messageID string, message []byte, ttl time.Duration) error
	GetMessage(messageID string) ([]byte, error)
	DeleteMessage(messageID string) error
	SetMessageStatus(messageID, status string, ttl time.Duration) error
	GetMessageStatus(messageID string) (string, error)
	IncrementCounter(key string) (int64, error)
	GetStats() (map[string]interface{}, error)
	Close() error
}

// MemoryStore 内存消息存储
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

// Close 关闭存储
func (ms *MemoryStore) Close() error {
	return nil
}

// RedisStore Redis消息存储
type RedisStore struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisStore 创建Redis存储实例
func NewRedisStore(addr, password string, db int) (MessageStoreInterface, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 测试连接
	ctx := context.Background()
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("连接Redis失败: %v", err)
	}

	return &RedisStore{
		client: client,
		ctx:    ctx,
	}, nil
}

// StoreMessage 存储消息
func (rs *RedisStore) StoreMessage(messageID string, message []byte, ttl time.Duration) error {
	key := fmt.Sprintf("msg:%s", messageID)
	return rs.client.Set(rs.ctx, key, message, ttl).Err()
}

// GetMessage 获取消息
func (rs *RedisStore) GetMessage(messageID string) ([]byte, error) {
	key := fmt.Sprintf("msg:%s", messageID)
	return rs.client.Get(rs.ctx, key).Bytes()
}

// DeleteMessage 删除消息
func (rs *RedisStore) DeleteMessage(messageID string) error {
	key := fmt.Sprintf("msg:%s", messageID)
	return rs.client.Del(rs.ctx, key).Err()
}

// SetMessageStatus 设置消息状态
func (rs *RedisStore) SetMessageStatus(messageID, status string, ttl time.Duration) error {
	key := fmt.Sprintf("status:%s", messageID)
	return rs.client.Set(rs.ctx, key, status, ttl).Err()
}

// GetMessageStatus 获取消息状态
func (rs *RedisStore) GetMessageStatus(messageID string) (string, error) {
	key := fmt.Sprintf("status:%s", messageID)
	return rs.client.Get(rs.ctx, key).Result()
}

// IncrementCounter 递增计数器
func (rs *RedisStore) IncrementCounter(key string) (int64, error) {
	return rs.client.Incr(rs.ctx, key).Result()
}

// GetStats 获取Redis统计信息
func (rs *RedisStore) GetStats() (map[string]interface{}, error) {
	info, err := rs.client.Info(rs.ctx, "memory", "keyspace", "stats").Result()
	if err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})
	stats["type"] = "redis"
	stats["redis_info"] = info

	// 获取消息相关的键数量
	msgKeys, err := rs.client.Keys(rs.ctx, "msg:*").Result()
	if err == nil {
		stats["stored_messages"] = len(msgKeys)
	}

	statusKeys, err := rs.client.Keys(rs.ctx, "status:*").Result()
	if err == nil {
		stats["message_statuses"] = len(statusKeys)
	}

	return stats, nil
}

// Close 关闭Redis连接
func (rs *RedisStore) Close() error {
	return rs.client.Close()
}

// SQLStore SQL数据库存储（MySQL/PostgreSQL）
type SQLStore struct {
	db       *sql.DB
	dbType   string
	counters map[string]int64
	mutex    sync.RWMutex
}

// NewSQLStore 创建SQL存储实例
func NewSQLStore(dbType, dsn string) (MessageStoreInterface, error) {
	db, err := sql.Open(dbType, dsn)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %v", err)
	}

	store := &SQLStore{
		db:       db,
		dbType:   dbType,
		counters: make(map[string]int64),
	}

	// 初始化表结构
	if err := store.initTables(); err != nil {
		return nil, fmt.Errorf("初始化表结构失败: %v", err)
	}

	return store, nil
}

// initTables 初始化数据库表
func (ss *SQLStore) initTables() error {
	// 消息表
	createMessagesTable := `
		CREATE TABLE IF NOT EXISTS ws_messages (
			id VARCHAR(255) PRIMARY KEY,
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP
		)`

	// 状态表
	createStatusTable := `
		CREATE TABLE IF NOT EXISTS ws_message_status (
			message_id VARCHAR(255) PRIMARY KEY,
			status VARCHAR(50) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP
		)`

	// PostgreSQL使用不同的语法
	if ss.dbType == "postgres" {
		createMessagesTable = `
			CREATE TABLE IF NOT EXISTS ws_messages (
				id VARCHAR(255) PRIMARY KEY,
				content TEXT NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				expires_at TIMESTAMP
			)`
		createStatusTable = `
			CREATE TABLE IF NOT EXISTS ws_message_status (
				message_id VARCHAR(255) PRIMARY KEY,
				status VARCHAR(50) NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				expires_at TIMESTAMP
			)`
	}

	if _, err := ss.db.Exec(createMessagesTable); err != nil {
		return err
	}

	if _, err := ss.db.Exec(createStatusTable); err != nil {
		return err
	}

	return nil
}

// StoreMessage 存储消息
func (ss *SQLStore) StoreMessage(messageID string, message []byte, ttl time.Duration) error {
	var expiresAt *time.Time
	if ttl > 0 {
		expiry := time.Now().Add(ttl)
		expiresAt = &expiry
	}

	query := `INSERT INTO ws_messages (id, content, expires_at) VALUES (?, ?, ?) 
			  ON DUPLICATE KEY UPDATE content = VALUES(content), expires_at = VALUES(expires_at)`

	if ss.dbType == "postgres" {
		query = `INSERT INTO ws_messages (id, content, expires_at) VALUES ($1, $2, $3) 
				 ON CONFLICT (id) DO UPDATE SET content = EXCLUDED.content, expires_at = EXCLUDED.expires_at`
	}

	_, err := ss.db.Exec(query, messageID, string(message), expiresAt)
	return err
}

// GetMessage 获取消息
func (ss *SQLStore) GetMessage(messageID string) ([]byte, error) {
	query := `SELECT content FROM ws_messages WHERE id = ? AND (expires_at IS NULL OR expires_at > NOW())`
	if ss.dbType == "postgres" {
		query = `SELECT content FROM ws_messages WHERE id = $1 AND (expires_at IS NULL OR expires_at > NOW())`
	}

	var content string
	err := ss.db.QueryRow(query, messageID).Scan(&content)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("消息不存在")
		}
		return nil, err
	}

	return []byte(content), nil
}

// DeleteMessage 删除消息
func (ss *SQLStore) DeleteMessage(messageID string) error {
	query := `DELETE FROM ws_messages WHERE id = ?`
	if ss.dbType == "postgres" {
		query = `DELETE FROM ws_messages WHERE id = $1`
	}

	_, err := ss.db.Exec(query, messageID)
	return err
}

// SetMessageStatus 设置消息状态
func (ss *SQLStore) SetMessageStatus(messageID, status string, ttl time.Duration) error {
	var expiresAt *time.Time
	if ttl > 0 {
		expiry := time.Now().Add(ttl)
		expiresAt = &expiry
	}

	query := `INSERT INTO ws_message_status (message_id, status, expires_at) VALUES (?, ?, ?) 
			  ON DUPLICATE KEY UPDATE status = VALUES(status), expires_at = VALUES(expires_at)`

	if ss.dbType == "postgres" {
		query = `INSERT INTO ws_message_status (message_id, status, expires_at) VALUES ($1, $2, $3) 
				 ON CONFLICT (message_id) DO UPDATE SET status = EXCLUDED.status, expires_at = EXCLUDED.expires_at`
	}

	_, err := ss.db.Exec(query, messageID, status, expiresAt)
	return err
}

// GetMessageStatus 获取消息状态
func (ss *SQLStore) GetMessageStatus(messageID string) (string, error) {
	query := `SELECT status FROM ws_message_status WHERE message_id = ? AND (expires_at IS NULL OR expires_at > NOW())`
	if ss.dbType == "postgres" {
		query = `SELECT status FROM ws_message_status WHERE message_id = $1 AND (expires_at IS NULL OR expires_at > NOW())`
	}

	var status string
	err := ss.db.QueryRow(query, messageID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("状态不存在")
		}
		return "", err
	}

	return status, nil
}

// IncrementCounter 递增计数器
func (ss *SQLStore) IncrementCounter(key string) (int64, error) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()
	ss.counters[key]++
	return ss.counters[key], nil
}

// GetStats 获取统计信息
func (ss *SQLStore) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	stats["type"] = ss.dbType

	// 获取消息数量
	var msgCount int
	query := `SELECT COUNT(*) FROM ws_messages WHERE expires_at IS NULL OR expires_at > NOW()`
	if err := ss.db.QueryRow(query).Scan(&msgCount); err == nil {
		stats["stored_messages"] = msgCount
	}

	// 获取状态数量
	var statusCount int
	query = `SELECT COUNT(*) FROM ws_message_status WHERE expires_at IS NULL OR expires_at > NOW()`
	if err := ss.db.QueryRow(query).Scan(&statusCount); err == nil {
		stats["message_statuses"] = statusCount
	}

	ss.mutex.RLock()
	stats["counters"] = len(ss.counters)
	ss.mutex.RUnlock()

	return stats, nil
}

// Close 关闭数据库连接
func (ss *SQLStore) Close() error {
	return ss.db.Close()
}
