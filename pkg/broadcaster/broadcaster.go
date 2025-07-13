package broadcaster

import (
	"GRUniChat-Broadcaster/internal/config"
	"GRUniChat-Broadcaster/internal/message"
	"GRUniChat-Broadcaster/pkg/logger"
	"GRUniChat-Broadcaster/pkg/middleware"
	"GRUniChat-Broadcaster/pkg/router"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
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
	config      *config.Config
	logger      logger.Logger
	regexCache  map[string]*regexp.Regexp
	mu          sync.RWMutex
}

// NewBroadcaster 创建新的广播器
func NewBroadcaster(rt *router.Router, mw *middleware.MiddlewareChain, cfg *config.Config, log logger.Logger) *Broadcaster {
	return &Broadcaster{
		connections: make(map[string]Connection),
		router:      rt,
		middleware:  mw,
		config:      cfg,
		logger:      log,
		regexCache:  make(map[string]*regexp.Regexp),
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
	b.logger.Debugf("路由目标服务器: %v", targets)

	// 检查是否为指定服务器执行的命令
	if processedMsg.Type == "command" && processedMsg.Body.ExecuteAt != "" {
		// 验证指定的服务器是否在连接的服务器列表中
		executeAtServer := processedMsg.Body.ExecuteAt
		found := false
		for _, server := range connectedServers {
			if server == executeAtServer {
				found = true
				break
			}
		}

		if found {
			targets = []string{executeAtServer}
			b.logger.Infof("命令指定在服务器 '%s' 执行", executeAtServer)
		} else {
			b.logger.Errorf("指定的服务器 '%s' 未连接，命令无法执行", executeAtServer)
			return fmt.Errorf("指定的服务器 '%s' 未连接", executeAtServer)
		}
	}

	b.logger.Debugf("最终目标服务器: %v", targets)

	// 应用组级别的黑名单过滤
	filteredTargets := b.applyGroupBlacklist(processedMsg, targets)
	if len(filteredTargets) != len(targets) {
		b.logger.Debugf("黑名单过滤后的目标服务器: %v", filteredTargets)
	}

	// 发送消息
	return b.sendToTargets(messageBytes, filteredTargets)
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

// applyGroupBlacklist 应用组级别的黑名单过滤
func (b *Broadcaster) applyGroupBlacklist(msg *message.Message, targets []string) []string {
	filtered := make([]string, 0, len(targets))

	for _, target := range targets {
		if b.shouldBlockMessage(msg, target) {
			b.logger.Debugf("消息被黑名单过滤: from=%s to=%s, type=%s", msg.From, target, msg.Type)
			continue
		}
		filtered = append(filtered, target)
	}

	return filtered
}

// shouldBlockMessage 检查消息是否应该被阻止发送到指定目标
func (b *Broadcaster) shouldBlockMessage(msg *message.Message, target string) bool {
	// 查找目标服务器所属的组
	var targetGroup *config.BroadcastGroup
	for _, group := range b.config.Groups {
		for _, server := range group.Members {
			if server == target {
				targetGroup = &group
				break
			}
		}
		if targetGroup != nil {
			break
		}
	}

	if targetGroup == nil {
		return false // 如果找不到组，不过滤
	}

	// 检查黑名单规则
	for _, rule := range targetGroup.Blacklist {
		if !rule.Enabled {
			continue
		}

		if b.matchesBlacklistRule(msg, &rule, target) {
			b.logger.Debugf("消息匹配黑名单规则: %s", rule.Name)
			return true
		}
	}

	return false
}

// matchesBlacklistRule 检查消息是否匹配黑名单规则
func (b *Broadcaster) matchesBlacklistRule(msg *message.Message, rule *config.GroupBlacklistRule, target string) bool {
	// 检查源服务器匹配
	if len(rule.From) > 0 && !b.matchesPattern(msg.From, rule.From) {
		return false
	}

	// 检查目标服务器匹配
	if len(rule.To) > 0 && !b.matchesPattern(target, rule.To) {
		return false
	}

	// 检查内容关键词匹配
	if len(rule.Content) > 0 && !b.matchesContent(msg, rule.Content) {
		return false
	}

	return true
}

// matchesPattern 检查字符串是否匹配模式列表（支持通配符）
func (b *Broadcaster) matchesPattern(str string, patterns []string) bool {
	for _, pattern := range patterns {
		if b.matchesWildcard(str, pattern) {
			return true
		}
	}
	return false
}

// matchesWildcard 简单的通配符匹配 (支持 * 通配符)
func (b *Broadcaster) matchesWildcard(str, pattern string) bool {
	if pattern == "*" {
		return true
	}

	if strings.Contains(pattern, "*") {
		// 转换为正则表达式
		regexPattern := strings.ReplaceAll(regexp.QuoteMeta(pattern), `\*`, `.*`)
		regexPattern = "^" + regexPattern + "$"

		if regex, exists := b.regexCache[regexPattern]; exists {
			return regex.MatchString(str)
		}

		if regex, err := regexp.Compile(regexPattern); err == nil {
			b.regexCache[regexPattern] = regex
			return regex.MatchString(str)
		}
	}

	return str == pattern
}

// matchesContent 检查消息内容是否匹配关键词列表
func (b *Broadcaster) matchesContent(msg *message.Message, patterns []string) bool {
	// 获取消息文本内容
	content := b.getMessageContent(msg)
	if content == "" {
		return false
	}

	for _, pattern := range patterns {
		// 如果模式以^开头，当作正则表达式处理
		if strings.HasPrefix(pattern, "^") || strings.HasPrefix(pattern, ".*") {
			if regex, exists := b.regexCache[pattern]; exists {
				if regex.MatchString(content) {
					return true
				}
			} else {
				if regex, err := regexp.Compile(pattern); err == nil {
					b.regexCache[pattern] = regex
					if regex.MatchString(content) {
						return true
					}
				}
			}
		} else {
			// 简单的字符串包含匹配
			if strings.Contains(strings.ToLower(content), strings.ToLower(pattern)) {
				return true
			}
		}
	}
	return false
}

// getMessageContent 从消息中提取文本内容
func (b *Broadcaster) getMessageContent(msg *message.Message) string {
	// 根据消息类型提取相应的文本内容
	switch msg.Type {
	case "chat":
		return msg.Body.ChatMessage
	case "command":
		return msg.Body.Command
	case "event":
		return msg.Body.EventDetail
	default:
		// 对于其他类型，尝试获取任何非空字段
		if msg.Body.ChatMessage != "" {
			return msg.Body.ChatMessage
		}
		if msg.Body.Command != "" {
			return msg.Body.Command
		}
		if msg.Body.EventDetail != "" {
			return msg.Body.EventDetail
		}
	}
	return ""
}

// UpdateConfig 更新配置（用于热重载）
func (b *Broadcaster) UpdateConfig(cfg *config.Config) {
	b.config = cfg
	// 清空正则表达式缓存
	b.regexCache = make(map[string]*regexp.Regexp)
	b.logger.Info("广播器配置已更新")
}
