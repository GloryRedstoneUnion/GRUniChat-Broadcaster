package utils

import (
	"strings"
)

// MatchesAny 检查值是否匹配任一模式
func MatchesAny(value string, patterns []string) bool {
	if len(patterns) == 0 {
		return true // 空列表表示匹配所有
	}

	for _, pattern := range patterns {
		if pattern == "*" || pattern == value || strings.Contains(value, pattern) {
			return true
		}
	}
	return false
}

// Contains 检查字符串数组是否包含指定值
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Remove 从字符串数组中移除指定值
func Remove(slice []string, item string) []string {
	result := make([]string, 0)
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// RemoveExcept 从群组成员中移除指定项，返回其他成员
func RemoveExcept(slice []string, except string) []string {
	result := make([]string, 0)
	for _, s := range slice {
		if s != except {
			result = append(result, s)
		}
	}
	return result
}

// RemoveDuplicates 移除字符串数组中的重复项
func RemoveDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// IsValidMessageType 检查是否为有效的消息类型
func IsValidMessageType(msgType string) bool {
	validTypes := []string{"chat", "event", "command", "hello", "response", "error"}
	return Contains(validTypes, msgType)
}
