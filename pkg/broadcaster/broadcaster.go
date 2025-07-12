package broadcaster

import (
	"GRUniChat-Broadcaster/internal/message"
	"GRUniChat-Broadcaster/pkg/logger"
	"GRUniChat-Broadcaster/pkg/middleware"
	"GRUniChat-Broadcaster/pkg/router"
	"encoding/json"
	"sync"
)

// Connection 连接接口
type Connection interface {
	GetID() string
	Send(data []byte) error
	IsConnected() bool
}

// Broadcaster 消息广播器
type Broadcaster struct {
	connections map[string]Connection
	router      *router.Router
	middleware  *middleware.MiddlewareChain
	logger      logger.Logger
	mu          sync.RWMutex
}

// NewBroadcaster 创建新的广播器
func NewBroadcaster(rt *router.Router, mw *middleware.MiddlewareChain, log logger.Logger) *Broadcaster {
	return &Broadcaster{
		connections: make(map[string]Connection),
		router:      rt,
		middleware:  mw,
		logger:      log,
	}
}

// AddConnection 添加连接
func (b *Broadcaster) AddConnection(conn Connection) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.connections[conn.GetID()] = conn
	b.logger.Infof("添加连接: %s", conn.GetID())
}

// RemoveConnection 移除连接
func (b *Broadcaster) RemoveConnection(connID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.connections[connID]; exists {
		delete(b.connections, connID)
		b.logger.Infof("移除连接: %s", connID)
	}
}

// GetConnections 获取所有连接ID
func (b *Broadcaster) GetConnections() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	connections := make([]string, 0, len(b.connections))
	for id := range b.connections {
		connections = append(connections, id)
	}
	return connections
}

// GetConnectionCount 获取连接数量
func (b *Broadcaster) GetConnectionCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.connections)
}

// GetAllConnections 获取所有连接（用于热重载时迁移连接）
func (b *Broadcaster) GetAllConnections() []Connection {
	b.mu.RLock()
	defer b.mu.RUnlock()

	connections := make([]Connection, 0, len(b.connections))
	for _, conn := range b.connections {
		connections = append(connections, conn)
	}
	return connections
}

// Broadcast 广播消息
func (b *Broadcaster) Broadcast(messageBytes []byte) error {
	var msg message.Message
	if err := json.Unmarshal(messageBytes, &msg); err != nil {
		b.logger.Errorf("解析消息失败: %v", err)
		return err
	}

	// 通过中间件处理消息
	processedMsg, err := b.middleware.Process(&msg)
	if err != nil {
		b.logger.Errorf("中间件处理失败: %v", err)
		return err
	}

	if processedMsg == nil {
		b.logger.Debug("消息被中间件过滤")
		return nil
	}

	b.logger.Infof("广播消息: from=%s, type=%s", processedMsg.From, processedMsg.Type)

	// 获取已连接的服务器列表
	connectedServers := b.GetConnections()

	// 获取目标服务器
	targets := b.router.GetTargets(processedMsg.From, connectedServers)
	b.logger.Debugf("目标服务器: %v", targets)

	// 发送消息
	return b.sendToTargets(messageBytes, targets)
}

// sendToTargets 发送消息到目标服务器
func (b *Broadcaster) sendToTargets(messageBytes []byte, targets []string) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	successCount := 0
	for _, target := range targets {
		conn, exists := b.connections[target]
		if !exists {
			b.logger.Debugf("目标连接不存在: %s", target)
			continue
		}

		if !conn.IsConnected() {
			b.logger.Debugf("目标连接已断开: %s", target)
			continue
		}

		if err := conn.Send(messageBytes); err != nil {
			b.logger.Errorf("发送到 %s 失败: %v", target, err)
		} else {
			b.logger.Debugf("消息已发送到: %s", target)
			successCount++
		}
	}

	b.logger.Infof("消息发送完成: %d/%d 成功", successCount, len(targets))
	return nil
}

// GetStats 获取广播器统计信息
func (b *Broadcaster) GetStats() map[string]interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()

	stats := map[string]interface{}{
		"total_connections": len(b.connections),
		"connections":       b.GetConnections(),
		"router_info":       b.router.GetRouteInfo(),
	}

	return stats
}
