package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"shangqing/internal/config"
	"shangqing/internal/dao"
	"shangqing/internal/handler"
	"shangqing/internal/middleware"
	"shangqing/internal/service"
)

var (
	configPath = flag.String("c", "config/config.yaml", "config file path")
)

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config error: %v", err)
	}

	// 初始化数据库
	db, err := dao.NewDB(&cfg.Database)
	if err != nil {
		log.Fatalf("init db error: %v", err)
	}
	log.Println("✅ MySQL connected")

	// 初始化 Redis
	redis, err := dao.NewRedis(&cfg.Redis)
	if err != nil {
		log.Fatalf("init redis error: %v", err)
	}
	log.Println("✅ Redis connected")

	// 初始化服务
	svc := service.NewServices(cfg, db, redis)

	// 初始化处理器
	h := handler.NewHandler(svc)

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 创建引擎
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// 注册路由
	setupRoutes(r, h, svc)

	// 创建服务器
	addr := ":" + cfg.Server.Port
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 启动服务器（后台运行）
	go func() {
		fmt.Printf(`
╔══════════════════════════════════════════════╗
║         🌌 熵清 V5 Backend 启动中...          ║
╠══════════════════════════════════════════════╣
║  Port: %s
║  Mode: %s
║  IERFT: S = B / J
╚══════════════════════════════════════════════╝
`, addr, cfg.Server.Mode)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// 等待中断信号优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited")
}

func setupRoutes(r *gin.Engine, h *handler.Handler, svc *service.Services) {
	// 健康检查
	r.GET("/health", h.Health)
	r.GET("/debug/secret", h.DebugSecret)

	// API v1
	v1 := r.Group("/api/v1")
	{
		// 公开接口
		v1.POST("/user/register", h.Register)
		v1.POST("/user/login", h.Login)

		// 需要认证的接口
		auth := v1.Group("")
		auth.Use(middleware.JWTAuth(svc.User))
		{
			// 用户
			auth.GET("/user/profile", h.GetProfile)

			// 对话
			auth.POST("/conversations", h.CreateConversation)
			auth.GET("/conversations", h.ListConversations)
			auth.GET("/conversations/:id", h.GetConversation)
			auth.POST("/conversations/:id/chat", h.Chat)

			// 分析
			auth.POST("/analyze", h.Analyze)

			// LLM
			auth.GET("/providers", h.ListProviders)
			auth.PUT("/providers/config", h.UpdateLLMConfig)
		}
	}
}
