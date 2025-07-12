package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"GRUniChat-Broadcaster/pkg/logger"
)

// HotReloader 配置热重载管理器
type HotReloader struct {
	configPath    string
	config        *Config
	lastModTime   time.Time
	logger        logger.Logger
	mu            sync.RWMutex
	onReload      func(*Config) error
	isPaused      bool
	pauseReason   string
	isInteractive bool // 是否启用交互式重载确认
	stopChan      chan bool
}

// NewHotReloader 创建新的热重载管理器
func NewHotReloader(configPath string, config *Config, log logger.Logger, interactive bool) (*HotReloader, error) {
	// 获取配置文件的初始修改时间
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		return nil, fmt.Errorf("获取配置文件信息失败: %v", err)
	}

	hr := &HotReloader{
		configPath:    configPath,
		config:        config,
		lastModTime:   fileInfo.ModTime(),
		logger:        log,
		isInteractive: interactive,
		stopChan:      make(chan bool),
	}

	return hr, nil
}

// SetReloadCallback 设置重载回调函数
func (hr *HotReloader) SetReloadCallback(callback func(*Config) error) {
	hr.onReload = callback
}

// Start 开始监听配置文件变化
func (hr *HotReloader) Start() {
	go hr.watchLoop()
	hr.logger.Info("配置热重载监听已启动")
}

// Stop 停止监听
func (hr *HotReloader) Stop() error {
	close(hr.stopChan)
	return nil
}

// GetConfig 获取当前配置（线程安全）
func (hr *HotReloader) GetConfig() *Config {
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	return hr.config
}

// PauseRouting 暂停路由处理
func (hr *HotReloader) PauseRouting(reason string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	hr.isPaused = true
	hr.pauseReason = reason
	hr.logger.Infof("路由处理已暂停: %s", reason)
}

// ResumeRouting 恢复路由处理
func (hr *HotReloader) ResumeRouting() {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	hr.isPaused = false
	hr.pauseReason = ""
	hr.logger.Info("路由处理已恢复")
}

// IsRoutingPaused 检查路由是否已暂停
func (hr *HotReloader) IsRoutingPaused() (bool, string) {
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	return hr.isPaused, hr.pauseReason
}

// watchLoop 监听文件变化的主循环（使用轮询方式）
func (hr *HotReloader) watchLoop() {
	ticker := time.NewTicker(2 * time.Second) // 每2秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fileInfo, err := os.Stat(hr.configPath)
			if err != nil {
				hr.logger.Errorf("检查配置文件状态失败: %v", err)
				continue
			}

			// 检查文件是否被修改
			if fileInfo.ModTime().After(hr.lastModTime) {
				hr.logger.Infof("检测到配置文件变化: %s", hr.configPath)
				hr.lastModTime = fileInfo.ModTime()

				// 暂停路由处理
				hr.PauseRouting("配置文件已变化，等待重载确认")

				// 延迟一下，确保文件写入完成
				time.Sleep(100 * time.Millisecond)

				if hr.isInteractive {
					hr.handleInteractiveReload()
				} else {
					hr.handleAutoReload()
				}
			}

		case <-hr.stopChan:
			return
		}
	}
}

// handleInteractiveReload 处理交互式重载
func (hr *HotReloader) handleInteractiveReload() {
	fmt.Println("\n[配置变化] 检测到配置文件变化！")
	fmt.Println("[警告] 路由处理已暂停，WebSocket连接保持活跃")
	fmt.Print("是否重载配置文件？(y/n/preview): ")

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		response := strings.ToLower(strings.TrimSpace(scanner.Text()))

		switch response {
		case "y", "yes":
			hr.performReload()
		case "n", "no":
			fmt.Println("[取消] 重载已取消，恢复路由处理")
			hr.ResumeRouting()
		case "p", "preview":
			hr.previewConfig()
			hr.handleInteractiveReload() // 递归调用以再次询问
		default:
			fmt.Println("[提示] 无效输入，请输入 y/n/preview")
			hr.handleInteractiveReload() // 递归调用以再次询问
		}
	}
}

// handleAutoReload 处理自动重载
func (hr *HotReloader) handleAutoReload() {
	hr.logger.Info("自动重载配置文件...")
	hr.performReload()
}

// previewConfig 预览新配置文件内容
func (hr *HotReloader) previewConfig() {
	fmt.Println("\n[预览] 配置文件预览:")
	fmt.Println("=" + strings.Repeat("=", 50))

	// 尝试加载新配置
	newConfig, err := Load(hr.configPath)
	if err != nil {
		fmt.Printf("[错误] 配置文件加载失败: %v\n", err)
		return
	}

	// 验证新配置
	if err := newConfig.Validate(); err != nil {
		fmt.Printf("[错误] 配置验证失败: %v\n", err)
		return
	}

	// 显示配置摘要
	fmt.Printf("[服务器] %s\n", newConfig.GetWebSocketURL())
	fmt.Printf("[数据库] %s\n", newConfig.Database.Type)

	if len(newConfig.Groups) > 0 {
		fmt.Printf("[群组] 数量: %d\n", len(newConfig.Groups))
		for i, group := range newConfig.Groups {
			if group.Enabled {
				fmt.Printf("   %d. %s (%d个成员)\n", i+1, group.Name, len(group.Members))
			}
		}
	}

	if len(newConfig.Rules) > 0 {
		enabledRules := 0
		for _, rule := range newConfig.Rules {
			if rule.Enabled {
				enabledRules++
			}
		}
		fmt.Printf("[规则] %d个启用/%d个总计\n", enabledRules, len(newConfig.Rules))
	}

	fmt.Println("=" + strings.Repeat("=", 50))
}

// performReload 执行重载操作
func (hr *HotReloader) performReload() {
	hr.logger.Info("开始重载配置文件...")

	// 加载新配置
	newConfig, err := Load(hr.configPath)
	if err != nil {
		hr.logger.Errorf("重载失败 - 配置文件加载错误: %v", err)
		fmt.Printf("[错误] 重载失败: %v\n", err)
		fmt.Println("[恢复] 恢复路由处理，继续使用旧配置")
		hr.ResumeRouting()
		return
	}

	// 验证新配置
	if err := newConfig.Validate(); err != nil {
		hr.logger.Errorf("重载失败 - 配置验证错误: %v", err)
		fmt.Printf("[错误] 重载失败: %v\n", err)
		fmt.Println("[恢复] 恢复路由处理，继续使用旧配置")
		hr.ResumeRouting()
		return
	}

	// 调用重载回调
	if hr.onReload != nil {
		if err := hr.onReload(newConfig); err != nil {
			hr.logger.Errorf("重载失败 - 回调错误: %v", err)
			fmt.Printf("[错误] 重载失败: %v\n", err)
			fmt.Println("[恢复] 恢复路由处理，继续使用旧配置")
			hr.ResumeRouting()
			return
		}
	}

	// 更新配置
	hr.mu.Lock()
	hr.config = newConfig
	hr.mu.Unlock()

	// 恢复路由处理
	hr.ResumeRouting()

	hr.logger.Info("配置重载成功")
	if hr.isInteractive {
		fmt.Println("[成功] 配置重载成功！路由处理已恢复")
	}
}
