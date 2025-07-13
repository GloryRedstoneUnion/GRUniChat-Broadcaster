package message

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Message 消息结构
type Message struct {
	From        string `json:"from"`
	Type        string `json:"type"`
	Body        Body   `json:"body"`
	TotalID     string `json:"totalId"` // 作为消息唯一ID使用
	CurrentTime string `json:"currentTime"`
}

// Body 消息体
type Body struct {
	Sender      string `json:"sender"`
	ChatMessage string `json:"chatMessage"`
	Command     string `json:"command"`
	ExecuteAt   string `json:"executeAt,omitempty"` // 当type=command时，指定命令在哪个服务器执行
	EventDetail string `json:"eventDetail"`
}

// AckMessage 消息确认结构
type AckMessage struct {
	TotalID   string `json:"totalId"`   // 原消息TotalID
	Type      string `json:"type"`      // 固定为 "ack"
	Status    string `json:"status"`    // "success" 或 "error"
	Message   string `json:"message"`   // 状态描述
	Timestamp string `json:"timestamp"` // 确认时间戳
}

// ErrorMessage 错误消息结构
type ErrorMessage struct {
	TotalID   string `json:"totalId"`   // 原消息TotalID（如果有）
	Type      string `json:"type"`      // 固定为 "error"
	Error     string `json:"error"`     // 错误描述
	Code      int    `json:"code"`      // 错误代码
	Timestamp string `json:"timestamp"` // 错误时间戳
}

// GetContent 获取消息内容摘要
func (m *Message) GetContent() string {
	if m.Body.ChatMessage != "" {
		return fmt.Sprintf("聊天消息: %s", m.Body.ChatMessage)
	}
	if m.Body.Command != "" {
		return fmt.Sprintf("命令: %s", m.Body.Command)
	}
	if m.Body.EventDetail != "" {
		return fmt.Sprintf("事件: %s", m.Body.EventDetail)
	}
	return fmt.Sprintf("消息类型: %s", m.Type)
}

// IsValidType 检查消息类型是否有效
func (m *Message) IsValidType() bool {
	validTypes := []string{"chat", "command", "event", "hello", "ping", "pong"}
	for _, validType := range validTypes {
		if m.Type == validType {
			return true
		}
	}
	return false
}

// UpdateTimestamp 更新时间戳
func (m *Message) UpdateTimestamp() {
	m.CurrentTime = time.Now().Format("2006-01-02 15:04:05")
}

// GenerateTotalID 生成TotalID（如果为空）
func (m *Message) GenerateTotalID() {
	if m.TotalID == "" {
		m.TotalID = GenerateMessageID()
	}
}

// GetMessageID 获取消息ID（使用TotalID）
func (m *Message) GetMessageID() string {
	return m.TotalID
}

// GenerateMessageID 生成唯一的消息ID
func GenerateMessageID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// NewAckMessage 创建新的确认消息
func NewAckMessage(totalID, status, message string) *AckMessage {
	return &AckMessage{
		TotalID:   totalID,
		Type:      "ack",
		Status:    status,
		Message:   message,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}
}

// NewErrorMessage 创建新的错误消息
func NewErrorMessage(totalID, error string, code int) *ErrorMessage {
	return &ErrorMessage{
		TotalID:   totalID,
		Type:      "error",
		Error:     error,
		Code:      code,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}
}

// ParseCommand 解析命令
func (b *Body) ParseCommand() (string, string) {
	if b.Command == "" {
		return "", ""
	}
	parts := strings.SplitN(b.Command, " ", 2)
	command := parts[0]
	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}
	return command, args
}

// IsPingPong 检查是否为Ping/Pong消息
func (m *Message) IsPingPong() bool {
	return m.Type == "ping" || m.Type == "pong"
}

// MatchEventDetail 匹配事件详情
func (b *Body) MatchEventDetail(pattern string) (string, bool) {
	if b.EventDetail == "" {
		return "", false
	}
	matched, err := regexp.MatchString(pattern, b.EventDetail)
	if err != nil {
		return "", false
	}
	if matched {
		return b.EventDetail, true
	}
	return "", false
}

// ValidateExecuteAt 验证 ExecuteAt 字段是否有效
func (m *Message) ValidateExecuteAt(validServers []string) error {
	if m.Type != "command" {
		return nil // 非命令消息不需要验证
	}

	if m.Body.ExecuteAt == "" {
		return fmt.Errorf("command类型消息必须指定executeAt字段")
	}

	// 如果提供了有效服务器列表，检查是否在列表中
	if len(validServers) > 0 {
		for _, server := range validServers {
			if server == m.Body.ExecuteAt {
				return nil
			}
		}
		return fmt.Errorf("指定的服务器 '%s' 不在有效服务器列表中", m.Body.ExecuteAt)
	}

	return nil
}
