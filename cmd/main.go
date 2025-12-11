package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lowercode-go-server/api/controller"
	"lowercode-go-server/api/route"
	"lowercode-go-server/bootstrap"
	"lowercode-go-server/internal/ws"
	"lowercode-go-server/repository"
	"lowercode-go-server/usecase"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("🚀 LowCode Go Server 启动中...")

	// ========== 1. 加载环境变量 ==========
	env := bootstrap.LoadEnv()

	// ========== 2. 初始化 Clerk ==========
	bootstrap.InitClerk()

	// ========== 3. 连接数据库 ==========
	db := bootstrap.NewDatabase(env.DatabaseURL)

	// ========== 4. 依赖注入 ==========
	// Repository 层
	pageRepo := repository.NewPageRepository(db)
	userRepo := repository.NewUserRepository(db)

	// WebSocket Hub（需要 PageService 接口，pageRepo 实现了它）
	// 类型断言：pageRepo 同时实现了 domain.PageRepository 和 ws.PageService
	hub := ws.NewHub(pageRepo.(ws.PageService))

	// UseCase 层
	pageUseCase := usecase.NewPageUseCase(pageRepo, hub)

	// Controller 层
	pageController := controller.NewPageController(pageUseCase)
	wsHandler := controller.NewWSHandler(hub, []string{
		"https://xxmudcloudxx.github.io", // 生产环境前端
	})
	webhookController := controller.NewWebhookController(userRepo, env.WebhookSecret)

	// ========== 5. 启动 Hub 事件循环 ==========
	go hub.Run()

	// ========== 6. 配置 Gin 路由 ==========
	router := gin.Default()

	// CORS 配置
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://xxmudcloudxx.github.io", "http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 设置路由
	route.Setup(router, &route.Dependencies{
		PageController:    pageController,
		WSHandler:         wsHandler,
		WebhookController: webhookController,
	})

	// ========== 7. 启动 HTTP 服务 ==========
	srv := &http.Server{
		Addr:    ":" + env.Port,
		Handler: router,
	}

	// 在 goroutine 中启动服务，主线程等待中断信号
	go func() {
		log.Printf("✅ 服务已启动: http://localhost:%s", env.Port)
		log.Printf("📖 API 文档:")
		log.Printf("   GET  /health              - 健康检查")
		log.Printf("   GET  /api/pages/:pageId   - 获取页面")
		log.Printf("   POST /api/pages           - 创建页面")
		log.Printf("   DELETE /api/pages/:pageId - 删除页面")
		log.Printf("   GET  /ws?pageId=xxx&token=xxx - WebSocket 连接")
		log.Printf("   POST /webhook/clerk       - Clerk Webhook")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ 服务启动失败: %v", err)
		}
	}()

	// ========== 8. 优雅停机 ==========
	quit := make(chan os.Signal, 1)
	// 监听 SIGINT (Ctrl+C) 和 SIGTERM (容器停止)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 收到停机信号，正在优雅关闭...")

	// 给 5 秒时间处理剩余请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ 服务强制关闭: %v", err)
	}

	// Hub 和 Room 的清理会在 srv.Shutdown 后自动触发
	// Room.Stop() 会调用 flushToDB，确保数据不丢失

	log.Println("✅ 服务已安全停止")
}
