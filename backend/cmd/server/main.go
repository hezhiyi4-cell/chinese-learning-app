package main

import (
	"log"
	"os"

	"chinese-learning-app/internal/config"
	"chinese-learning-app/internal/database"
	"chinese-learning-app/internal/handlers"
	"chinese-learning-app/internal/middleware"
	"chinese-learning-app/internal/repositories"
	"chinese-learning-app/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	if err := database.InitDB(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	db := database.GetDB()

	// 初始化 Repositories
	userRepo := repositories.NewUserRepository(db)
	courseRepo := repositories.NewCourseRepository(db)
	progressRepo := repositories.NewProgressRepository(db)
	paymentRepo := repositories.NewPaymentRepository(db)
	toneBattleRepo := repositories.NewToneBattleRepository(db)

	// 初始化 Services
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	courseService := services.NewCourseService(courseRepo)
	progressService := services.NewProgressService(progressRepo, userRepo, courseRepo)
	aiService := services.NewAIService(cfg.OpenAIAPIKey)
	paypalGateway := services.NewPayPalGateway(cfg.PayPalClientID, cfg.PayPalSecret, cfg.PayPalBaseURL)
	paymentService := services.NewPaymentService(paymentRepo, userRepo, cfg.FrontendBaseURL, paypalGateway)
	toneBattleService := services.NewToneBattleService(toneBattleRepo, userRepo, cfg.RedisAddr, cfg.RedisPassword)

	// 初始化 Handlers
	authHandler := handlers.NewAuthHandler(authService)
	courseHandler := handlers.NewCourseHandler(courseService, progressService)
	aiHandler := handlers.NewAIHandler(aiService)
	paymentHandler := handlers.NewPaymentHandler(paymentService)
	toneBattleHandler := handlers.NewToneBattleHandler(toneBattleService, userRepo, cfg.JWTSecret)

	// 填充初始数据
	if err := database.SeedInitialData(courseRepo); err != nil {
		log.Printf("Warning: Failed to seed initial data: %v", err)
	}
	if err := database.SeedToneBattleQuestions(toneBattleRepo); err != nil {
		log.Printf("Warning: Failed to seed tone battle questions: %v", err)
	}
	if err := database.EnsureDefaultAdmin(db, cfg.DefaultAdminEmail, cfg.DefaultAdminPassword); err != nil {
		log.Printf("Warning: Failed to ensure default admin: %v", err)
	}

	// 初始化 Gin
	r := gin.Default()
	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		log.Fatalf("Failed to prepare upload directory: %v", err)
	}
	r.Static("/uploads", cfg.UploadDir)

	// CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Chinese Learning App Backend is running!",
		})
	})

	// API v1 路由
	apiV1 := r.Group("/api/v1")
	{
		apiV1.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"version": "1.0",
				"message": "Welcome to Chinese Learning API",
			})
		})

		// 认证路由（公开）
		auth := apiV1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// 课程路由（公开）
		apiV1.GET("/courses", courseHandler.GetCourses)
		apiV1.GET("/courses/:id", courseHandler.GetCourseDetail)
		apiV1.GET("/lessons/:id", courseHandler.GetLesson)
		apiV1.GET("/lessons/:id/pronunciation", courseHandler.GetLessonPronunciation)
		apiV1.GET("/tone-battle/questions", toneBattleHandler.ListQuestions)
		apiV1.GET("/tone-battle/ws", toneBattleHandler.HandleWebSocket)

		// AI 路由（部分公开）
		ai := apiV1.Group("/ai")
		{
			ai.POST("/speech-to-text", aiHandler.SpeechToText)
			ai.POST("/evaluate", aiHandler.Evaluate)
		}

		// 需要认证的普通用户路由
		protected := apiV1.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			protected.GET("/progress", courseHandler.GetProgress)
			protected.POST("/progress/:lessonId", courseHandler.UpdateProgress)
			protected.GET("/stats", courseHandler.GetStats)
			protected.GET("/user/stats", courseHandler.GetStats)
			protected.GET("/leaderboard", courseHandler.GetLeaderboard)
			protected.GET("/payments/catalog", paymentHandler.GetCatalog)
			protected.POST("/payments/checkout", paymentHandler.CreateCheckout)
			protected.POST("/payments/orders/:id/confirm", paymentHandler.ConfirmCheckout)

			// 需要认证的 AI 路由
			protected.POST("/ai/chat", aiHandler.Chat)
			protected.GET("/ai/scenes", aiHandler.GetScenes)
		}

		// 管理员路由
		adminProtected := apiV1.Group("/")
		adminProtected.Use(middleware.AuthMiddleware(cfg.JWTSecret), middleware.AdminOnlyMiddleware())
		{
			adminProtected.GET("/admin/courses", courseHandler.GetCoursesForAdmin)
			adminProtected.PATCH("/admin/courses/:id/publish", courseHandler.UpdateCoursePublishStatus)
			adminProtected.POST("/admin/courses/:id/upload", courseHandler.UploadCourseThumbnail)
			adminProtected.DELETE("/admin/courses/:id/thumbnail", courseHandler.DeleteCourseThumbnail)
			adminProtected.PUT("/admin/courses/reorder", courseHandler.ReorderCourses)
			adminProtected.PUT("/admin/lessons/reorder", courseHandler.ReorderLessons)
			adminProtected.POST("/courses", courseHandler.CreateCourse)
			adminProtected.PUT("/courses/:id", courseHandler.UpdateCourse)
			adminProtected.DELETE("/courses/:id", courseHandler.DeleteCourse)
			adminProtected.POST("/courses/:id/lessons", courseHandler.CreateLesson)
			adminProtected.PUT("/lessons/:id", courseHandler.UpdateLesson)
			adminProtected.DELETE("/lessons/:id", courseHandler.DeleteLesson)
		}
	}

	log.Printf("🚀 Server starting on :%s...", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
