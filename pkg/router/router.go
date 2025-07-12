package router

import (
	"GRUniChat-Broadcaster/internal/config"
	"GRUniChat-Broadcaster/pkg/logger"
	"GRUniChat-Broadcaster/pkg/utils"
	"sync"
)

// Router 消息路由器
type Router struct {
	config      *config.Config
	logger      logger.Logger
	mu          sync.RWMutex
	hotReloader *config.HotReloader // 热重载器引用
}

// NewRouter 创建新的路由器
func NewRouter(cfg *config.Config, log logger.Logger) *Router {
	return &Router{
		config: cfg,
		logger: log,
	}
}

// SetHotReloader 设置热重载器引用
func (r *Router) SetHotReloader(hr *config.HotReloader) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hotReloader = hr
}

// GetTargets 获取消息目标列表
func (r *Router) GetTargets(fromServer string, connectedServers []string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 检查路由是否被暂停
	if r.hotReloader != nil {
		if isPaused, reason := r.hotReloader.IsRoutingPaused(); isPaused {
			r.logger.Infof("路由已暂停，跳过消息路由: %s", reason)
			return []string{} // 返回空列表，不路由任何消息
		}
	}

	r.logger.Debugf("路由查询: from=%s, connected=%v", fromServer, connectedServers)

	// 优先检查groups配置
	for _, group := range r.config.Groups {
		if utils.Contains(group.Members, fromServer) {
			targets := r.filterConnectedServers(utils.RemoveExcept(group.Members, fromServer), connectedServers)
			r.logger.Debugf("群组路由结果: %v", targets)
			return targets
		}
	}

	// 回退到rules配置
	var targets []string
	for _, rule := range r.config.Rules {
		if !rule.Enabled {
			continue
		}

		// 检查是否匹配来源
		if utils.MatchesAny(fromServer, rule.FromSources) {
			ruleTargets := r.resolveTargets(rule.ToTargets, fromServer, connectedServers)
			targets = append(targets, ruleTargets...)
			r.logger.Debugf("规则 '%s' 匹配，添加目标: %v", rule.Name, ruleTargets)
		}
	}

	// 去重
	targets = utils.RemoveDuplicates(targets)
	r.logger.Debugf("最终路由结果: %v", targets)
	return targets
}

// resolveTargets 解析目标列表，处理通配符
func (r *Router) resolveTargets(toTargets []string, fromServer string, connectedServers []string) []string {
	var resolved []string

	for _, target := range toTargets {
		if target == "*" {
			// "*" 表示所有其他已连接的客户端
			for _, server := range connectedServers {
				if server != fromServer {
					resolved = append(resolved, server)
				}
			}
		} else {
			resolved = append(resolved, target)
		}
	}

	// 只返回已连接的服务器
	return r.filterConnectedServers(resolved, connectedServers)
}

// filterConnectedServers 过滤出已连接的服务器
func (r *Router) filterConnectedServers(targets []string, connectedServers []string) []string {
	var filtered []string
	for _, target := range targets {
		if utils.Contains(connectedServers, target) {
			filtered = append(filtered, target)
		}
	}
	return filtered
}

// IsValidRoute 检查路由是否有效
func (r *Router) IsValidRoute(fromServer, toServer string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 检查groups
	for _, group := range r.config.Groups {
		if utils.Contains(group.Members, fromServer) && utils.Contains(group.Members, toServer) {
			return true
		}
	}

	// 检查rules
	for _, rule := range r.config.Rules {
		if !rule.Enabled {
			continue
		}
		if utils.MatchesAny(fromServer, rule.FromSources) {
			if utils.Contains(rule.ToTargets, "*") || utils.Contains(rule.ToTargets, toServer) {
				return true
			}
		}
	}

	return false
}

// GetRouteInfo 获取路由信息用于调试
func (r *Router) GetRouteInfo() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return map[string]interface{}{
		"groups_count": len(r.config.Groups),
		"rules_count":  len(r.config.Rules),
		"groups":       r.config.Groups,
		"rules":        r.config.Rules,
	}
}
