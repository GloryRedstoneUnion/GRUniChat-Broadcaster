package connection

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"GRUniChat-Broadcaster/internal/config"
	"GRUniChat-Broadcaster/internal/message"
	"GRUniChat-Broadcaster/pkg/broadcaster"
	"GRUniChat-Broadcaster/pkg/database"
	"GRUniChat-Broadcaster/pkg/logger"
	"GRUniChat-Broadcaster/pkg/middleware"
	"GRUniChat-Broadcaster/pkg/router"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境中应该更严格
	},
}

// WSConnection WebSocket连接实现
type WSConnection struct {
	ws              *websocket.Conn
	serverID        string
	send            chan []byte
	isAuthenticated bool
	logger          logger.Logger
}

// NewWSConnection 创建新的WebSocket连接
func NewWSConnection(ws *websocket.Conn, log logger.Logger) *WSConnection {
	return &WSConnection{
		ws:              ws,
		serverID:        "",
		send:            make(chan []byte, 256),
		isAuthenticated: false,
		logger:          log,
	}
}

// GetID 实现broadcaster.Connection接口
func (c *WSConnection) GetID() string {
	return c.serverID
}

// Send 实现broadcaster.Connection接口
func (c *WSConnection) Send(data []byte) error {
	select {
	case c.send <- data:
		return nil
	default:
		return ErrChannelFull
	}
}

// IsConnected 实现broadcaster.Connection接口
func (c *WSConnection) IsConnected() bool {
	return c.ws != nil && c.isAuthenticated
}

// ConnectionManager 管理所有WebSocket连接
type ConnectionManager struct {
	broadcaster  *broadcaster.Broadcaster
	config       *config.Config
	logger       logger.Logger
	messageStore database.MessageStoreInterface
	messageTTL   time.Duration
}

// NewConnectionManager 创建新的连接管理器
func NewConnectionManager(cfg *config.Config, log logger.Logger) (*ConnectionManager, error) {
	// 创建路由器
	rt := router.NewRouter(cfg, log)

	// 创建中间件链
	mw := middleware.NewMiddlewareChain(log)
	mw.Add(middleware.NewAuthMiddleware(log))
	mw.Add(middleware.NewValidationMiddleware(log))
	mw.Add(middleware.NewLoggingMiddleware(log))

	// 创建广播器
	bc := broadcaster.NewBroadcaster(rt, mw, log)

	// 创建消息存储
	messageStore, err := database.CreateMessageStore(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("创建消息存储失败: %v", err)
	}

	// 获取消息TTL
	messageTTL := database.GetMessageTTL(&cfg.Database)

	cm := &ConnectionManager{
		broadcaster:  bc,
		config:       cfg,
		logger:       log,
		messageStore: messageStore,
		messageTTL:   messageTTL,
	}

	log.Infof("消息存储初始化成功，类型: %s", cfg.Database.Type)
	return cm, nil
}

// Stop 停止连接管理器
func (cm *ConnectionManager) Stop() error {
	if cm.messageStore != nil {
		return cm.messageStore.Close()
	}
	return nil
}

// UpdateConfig 更新配置（用于热重载）
func (cm *ConnectionManager) UpdateConfig(newConfig *config.Config) error {
	cm.logger.Info("正在更新连接管理器配置...")

	// 创建新的路由器
	newRouter := router.NewRouter(newConfig, cm.logger)

	// 创建新的中间件链
	mw := middleware.NewMiddlewareChain(cm.logger)
	mw.Add(middleware.NewAuthMiddleware(cm.logger))
	mw.Add(middleware.NewValidationMiddleware(cm.logger))
	mw.Add(middleware.NewLoggingMiddleware(cm.logger))

	// 创建新的广播器
	newBroadcaster := broadcaster.NewBroadcaster(newRouter, mw, cm.logger)

	// 迁移现有连接到新的广播器
	connections := cm.broadcaster.GetAllConnections()
	for _, conn := range connections {
		newBroadcaster.AddConnection(conn)
		cm.broadcaster.RemoveConnection(conn.GetID())
	}

	// 更新广播器和配置
	cm.broadcaster = newBroadcaster
	cm.config = newConfig

	cm.logger.Info("连接管理器配置更新完成")
	return nil
}

// SetHotReloader 设置热重载器引用
func (cm *ConnectionManager) SetHotReloader(hr *config.HotReloader) {
	// 这里我们需要访问路由器，但目前路由器在broadcaster内部
	// 我们需要重新设计架构或者通过broadcaster获取路由器
	cm.logger.Info("热重载器已关联到连接管理器")
}

// HandleWebSocket 处理WebSocket连接
func (cm *ConnectionManager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		cm.logger.Errorf("WebSocket升级失败: %v", err)
		return
	}

	conn := NewWSConnection(ws, cm.logger)
	go conn.writePump(cm)
	go conn.readPump(cm)
}

func (c *WSConnection) readPump(cm *ConnectionManager) {
	defer func() {
		if c.isAuthenticated && c.serverID != "" {
			cm.broadcaster.RemoveConnection(c.serverID)
		}
		c.ws.Close()
	}()

	for {
		_, messageBytes, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Errorf("WebSocket错误: %v", err)
			}
			break
		}

		var msg message.Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			c.logger.Errorf("解析消息失败: %v", err)

			// 发送错误回复
			errorMsg := message.NewErrorMessage("", "消息格式错误", 400)
			if errorBytes, err := json.Marshal(errorMsg); err == nil {
				c.Send(errorBytes)
			}
			continue
		}

		// 生成消息TotalID（如果没有）
		msg.GenerateTotalID()

		// 验证消息基本格式
		if msg.From == "" || !msg.IsValidType() {
			c.logger.Errorf("无效消息格式: %+v", msg)

			// 发送错误回复
			errorMsg := message.NewErrorMessage(msg.TotalID, "消息格式验证失败", 400)
			if errorBytes, err := json.Marshal(errorMsg); err == nil {
				c.Send(errorBytes)
			}
			continue
		}

		// 更新时间戳
		msg.UpdateTimestamp()

		// 处理hello消息进行身份验证
		if msg.Type == "hello" && !c.isAuthenticated {
			c.serverID = msg.From
			c.isAuthenticated = true
			cm.broadcaster.AddConnection(c)
			c.logger.Infof("客户端 %s 已通过hello消息认证", c.serverID)

			// 发送确认消息
			ackMsg := message.NewAckMessage(msg.TotalID, "success", "认证成功")
			if ackBytes, err := json.Marshal(ackMsg); err == nil {
				c.Send(ackBytes)
			}
			continue
		}

		// 只有已认证的连接才能发送其他消息
		if !c.isAuthenticated {
			c.logger.Errorf("未认证的连接尝试发送消息: %s", msg.Type)

			// 发送错误回复
			errorMsg := message.NewErrorMessage(msg.TotalID, "未认证", 401)
			if errorBytes, err := json.Marshal(errorMsg); err == nil {
				c.Send(errorBytes)
			}
			continue
		}

		// 存储消息到数据库
		msgBytes, _ := json.Marshal(msg)
		if err := cm.messageStore.StoreMessage(msg.TotalID, msgBytes, cm.messageTTL); err != nil {
			c.logger.Errorf("存储消息到数据库失败: %v", err)
		}

		// 设置消息状态为处理中
		if err := cm.messageStore.SetMessageStatus(msg.TotalID, "processing", cm.messageTTL); err != nil {
			c.logger.Errorf("设置消息状态失败: %v", err)
		}

		// 广播消息
		if err := cm.broadcaster.Broadcast(msgBytes); err != nil {
			c.logger.Errorf("广播消息失败: %v", err)

			// 设置消息状态为失败
			cm.messageStore.SetMessageStatus(msg.TotalID, "failed", cm.messageTTL)

			// 发送错误回复
			errorMsg := message.NewErrorMessage(msg.TotalID, "广播失败", 500)
			if errorBytes, err := json.Marshal(errorMsg); err == nil {
				c.Send(errorBytes)
			}
		} else {
			// 设置消息状态为成功
			cm.messageStore.SetMessageStatus(msg.TotalID, "success", cm.messageTTL)

			// 发送确认消息
			ackMsg := message.NewAckMessage(msg.TotalID, "success", "消息已成功广播")
			if ackBytes, err := json.Marshal(ackMsg); err == nil {
				c.Send(ackBytes)
			}
		}
	}
}

func (c *WSConnection) writePump(cm *ConnectionManager) {
	defer c.ws.Close()

	for message := range c.send {
		if err := c.ws.WriteMessage(websocket.TextMessage, message); err != nil {
			c.logger.Errorf("发送消息失败: %v", err)
			return
		}
	}
	// Channel closed, send close message
	c.ws.WriteMessage(websocket.CloseMessage, []byte{})
}

// GetStats 获取连接管理器统计信息
func (cm *ConnectionManager) GetStats() map[string]interface{} {
	stats := cm.broadcaster.GetStats()

	// 添加数据库统计信息
	if cm.messageStore != nil {
		if dbStats, err := cm.messageStore.GetStats(); err == nil {
			stats["database"] = dbStats
		}
	}

	return stats
}

// GetMessageStatus 获取消息状态
func (cm *ConnectionManager) GetMessageStatus(messageID string) (string, error) {
	if cm.messageStore == nil {
		return "", fmt.Errorf("消息存储未初始化")
	}
	return cm.messageStore.GetMessageStatus(messageID)
}

// GetMessage 获取消息内容
func (cm *ConnectionManager) GetMessage(messageID string) ([]byte, error) {
	if cm.messageStore == nil {
		return nil, fmt.Errorf("消息存储未初始化")
	}
	return cm.messageStore.GetMessage(messageID)
}

// 错误定义
var (
	ErrChannelFull = fmt.Errorf("发送通道已满")
)
