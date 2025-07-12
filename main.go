package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"websocket_broadcaster/internal/config"
	"websocket_broadcaster/internal/connection"
	"websocket_broadcaster/pkg/logger"
)

// 版本信息变量，通过编译时 -ldflags 注入
var (
	Version   = "dev"     // 版本号
	BuildTime = "unknown" // 编译时间
)

// GitHubRelease GitHub API 响应结构
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
}

// checkForUpdates 检查更新
func checkForUpdates() {
	if Version == "dev" {
		return // 开发版本跳过更新检查
	}

	fmt.Print(">>> 检查更新中...")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/GloryRedstoneUnion/GRUniChat-MCDR/releases/latest")
	if err != nil {
		fmt.Printf(" 检查失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf(" 检查失败: HTTP %d\n", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf(" 读取响应失败: %v\n", err)
		return
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		fmt.Printf(" 解析响应失败: %v\n", err)
		return
	}

	latestVersion := release.TagName
	currentVersion := Version

	// 简单的版本比较
	if compareVersions(currentVersion, latestVersion) < 0 {
		fmt.Printf(" 发现新版本!\n")
		fmt.Printf(">>> 当前版本: %s\n", currentVersion)
		fmt.Printf(">>> 最新版本: %s\n", latestVersion)
		fmt.Printf(">>> 下载地址: %s\n\n", release.HTMLURL)
	} else {
		fmt.Printf(" 已是最新版本\n")
	}
}

// compareVersions 比较版本号，返回 -1(当前版本较老), 0(相同), 1(当前版本较新)
func compareVersions(current, latest string) int {
	// 移除 'v' 前缀
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	// 提取版本号部分 (去除 pre-release 标识)
	re := regexp.MustCompile(`^(\d+\.\d+\.\d+)`)
	currentMatch := re.FindString(current)
	latestMatch := re.FindString(latest)

	if currentMatch == "" || latestMatch == "" {
		return 0 // 无法解析版本号
	}

	currentParts := strings.Split(currentMatch, ".")
	latestParts := strings.Split(latestMatch, ".")

	for i := 0; i < 3; i++ {
		var currentNum, latestNum int
		if i < len(currentParts) {
			fmt.Sscanf(currentParts[i], "%d", &currentNum)
		}
		if i < len(latestParts) {
			fmt.Sscanf(latestParts[i], "%d", &latestNum)
		}

		if currentNum < latestNum {
			return -1
		} else if currentNum > latestNum {
			return 1
		}
	}
	return 0
}

// printBanner 打印启动横幅
func printBanner() {
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("    ____  ____   _   _         _   ____  _             _                       ")
	fmt.Println("   / ___||  _ \\ | | | | _ __  (_) / ___|| |__    __ _ | |_                     ")
	fmt.Println("  | |  _ | |_) || | | || '_ \\ | || |    | '_ \\  / _` || __|                    ")
	fmt.Println("  | |_| ||  _ < | |_| || | | || || |___ | | | || (_| || |_                     ")
	fmt.Println("   \\____||_| \\_\\_\\___/ |_| |_||_| \\____||_| |_| \\__,_| \\__|      _             ")
	fmt.Println("              | __ )  _ __  ___    __ _   __| |  ___  __ _  ___ | |_  ___  _ __ ")
	fmt.Println("              |  _ \\ | '__|/ _ \\  / _` | / _` | / __|/ _` |/ __|| __|/ _ \\| '__|")
	fmt.Println("              | |_) || |  | (_) || (_| || (_| || (__| (_| |\\__ \\| |_|  __/| |  ")
	fmt.Println("              |____/ |_|   \\___/  \\__,_| \\__,_| \\___|\\__,_||___/ \\__|\\___||_|  ")
	fmt.Println()
	fmt.Printf("                         WebSocket 消息广播器 %s                          \n", Version)
	fmt.Println()
	fmt.Println("                         作者: Glory Redstone Union                           ")
	fmt.Println("                     描述: 用于跨平台消息同步的WebSocket广播服务               ")
	if BuildTime != "unknown" {
		fmt.Printf("                         编译时间: %s                        \n", BuildTime)
	}
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════")
	fmt.Println()
}

func main() {
	// 打印启动横幅
	printBanner()

	var configFile = flag.String("config", "config.yaml", "配置文件路径")
	var debugMode = flag.Bool("debug", false, "启用调试模式")
	var hotReload = flag.Bool("hot-reload", true, "启用配置热重载")
	var interactive = flag.Bool("interactive", true, "启用交互式热重载确认")
	var noCheckUpdate = flag.Bool("no-check-update", false, "跳过版本检查")
	flag.Parse()

	// 检查更新（如果未被禁用）
	if !*noCheckUpdate && Version != "dev" {
		checkForUpdates()
	}

	// 初始化日志器
	log := logger.NewDefaultLogger(*debugMode)

	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Errorf("加载配置文件失败: %v", err)
		os.Exit(1)
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		log.Errorf("配置验证失败: %v", err)
		os.Exit(1)
	}

	// 创建热重载管理器
	var hotReloader *config.HotReloader
	if *hotReload {
		hotReloader, err = config.NewHotReloader(*configFile, cfg, log, *interactive)
		if err != nil {
			log.Errorf("创建热重载管理器失败: %v", err)
			os.Exit(1)
		}
	}

	cm, err := connection.NewConnectionManager(cfg, log)
	if err != nil {
		log.Errorf("创建连接管理器失败: %v", err)
		os.Exit(1)
	}

	// 设置热重载器与路由器的关联
	if hotReloader != nil {
		cm.SetHotReloader(hotReloader)
	}

	// 设置路由
	http.HandleFunc(cfg.Server.Path, cm.HandleWebSocket)

	// 统计信息API
	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		stats := cm.GetStats()
		if data, err := json.Marshal(stats); err == nil {
			w.Write(data)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"统计信息序列化失败"}`))
		}
	})

	// 消息状态查询API
	http.HandleFunc("/api/message/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// 提取消息ID
		messageID := r.URL.Path[len("/api/message/"):]
		if messageID == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"消息ID不能为空"}`))
			return
		}

		// 查询消息状态
		status, err := cm.GetMessageStatus(messageID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"消息未找到"}`))
			return
		}

		response := map[string]interface{}{
			"message_id": messageID,
			"status":     status,
			"timestamp":  time.Now().UnixMilli(),
		}

		if data, err := json.Marshal(response); err == nil {
			w.Write(data)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"响应序列化失败"}`))
		}
	})

	server := &http.Server{
		Addr: cfg.GetServerAddr(),
	}

	// 设置热重载回调
	if hotReloader != nil {
		hotReloader.SetReloadCallback(func(newConfig *config.Config) error {
			// 更新连接管理器的配置
			if err := cm.UpdateConfig(newConfig); err != nil {
				return fmt.Errorf("更新连接管理器配置失败: %v", err)
			}
			return nil
		})

		// 启动热重载监听
		hotReloader.Start()
	}

	fmt.Printf(">>> 服务器启动成功: %s\n", cfg.GetWebSocketURL())
	if *hotReload {
		fmt.Printf(">>> 配置热重载: 已启用\n")
	}
	if *debugMode {
		fmt.Printf(">>> 调试模式: 已启用\n")
	}
	fmt.Printf(">>> 按 Ctrl+C 停止服务器\n\n")

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorf("服务器启动失败: %v", err)
			os.Exit(1)
		}
	}()

	<-quit
	fmt.Printf("\n>>> 正在关闭服务器...\n")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 先停止热重载器
	if hotReloader != nil {
		if err := hotReloader.Stop(); err != nil {
			log.Errorf("停止热重载器失败: %v", err)
		}
	}

	// 停止连接管理器
	if err := cm.Stop(); err != nil {
		log.Errorf("停止连接管理器失败: %v", err)
	}

	if err := server.Shutdown(ctx); err != nil {
		log.Errorf("服务器关闭失败: %v", err)
	} else {
		fmt.Printf(">>> 服务器已关闭\n")
	}
}
