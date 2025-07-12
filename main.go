package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"encoding/json"
	"websocket_broadcaster/internal/config"
	"websocket_broadcaster/internal/connection"
	"websocket_broadcaster/pkg/logger"
)

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
	fmt.Println("                         WebSocket 消息广播器 v1.0.0                          ")
	fmt.Println()
	fmt.Println("                      作者: Glory Redstone Union - caikun233                         ")
	fmt.Println("                     描述: 用于跨平台消息同步的WebSocket广播服务               ")
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
	flag.Parse()

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
