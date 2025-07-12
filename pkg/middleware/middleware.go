package middleware

import (
	"GRUniChat-Broadcaster/internal/message"
	"GRUniChat-Broadcaster/pkg/logger"
)

// Middleware WebSocket消息中间件接口
type Middleware interface {
	Process(msg *message.Message) (*message.Message, error)
}

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	logger logger.Logger
}

func NewAuthMiddleware(log logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{logger: log}
}

func (m *AuthMiddleware) Process(msg *message.Message) (*message.Message, error) {
	if msg.From == "" {
		m.logger.Error("消息缺少发送者信息")
		return nil, nil
	}
	return msg, nil
}

// ValidationMiddleware 消息验证中间件
type ValidationMiddleware struct {
	logger logger.Logger
}

func NewValidationMiddleware(log logger.Logger) *ValidationMiddleware {
	return &ValidationMiddleware{logger: log}
}

func (m *ValidationMiddleware) Process(msg *message.Message) (*message.Message, error) {
	// 验证消息格式
	if msg.Type == "" {
		m.logger.Error("消息类型不能为空")
		return nil, nil
	}

	// Body是结构体，不需要nil检查
	m.logger.Debugf("消息验证通过: type=%s", msg.Type)

	return msg, nil
}

// LoggingMiddleware 日志中间件
type LoggingMiddleware struct {
	logger logger.Logger
}

func NewLoggingMiddleware(log logger.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{logger: log}
}

func (m *LoggingMiddleware) Process(msg *message.Message) (*message.Message, error) {
	m.logger.Debugf("处理消息: from=%s, type=%s", msg.From, msg.Type)
	return msg, nil
}

// MiddlewareChain 中间件链
type MiddlewareChain struct {
	middlewares []Middleware
	logger      logger.Logger
}

func NewMiddlewareChain(log logger.Logger) *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: make([]Middleware, 0),
		logger:      log,
	}
}

func (c *MiddlewareChain) Add(middleware Middleware) {
	c.middlewares = append(c.middlewares, middleware)
}

func (c *MiddlewareChain) Process(msg *message.Message) (*message.Message, error) {
	current := msg
	var err error

	for _, middleware := range c.middlewares {
		if current == nil {
			break
		}
		current, err = middleware.Process(current)
		if err != nil {
			c.logger.Errorf("中间件处理失败: %v", err)
			return nil, err
		}
	}

	return current, nil
}
