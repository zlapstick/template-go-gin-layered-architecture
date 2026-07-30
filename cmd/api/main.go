// @title Gin Golang Layered Architecture Template
// @version 1.0
// @description By Fardan Maula Azizi
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Ketik "Bearer" diikuti spasi dan access token.

package main

import (
	"go-gin-layered-architecture/internal/config"
	"go-gin-layered-architecture/internal/handler"
	"go-gin-layered-architecture/internal/helper"
	"go-gin-layered-architecture/internal/middleware"
	"go-gin-layered-architecture/internal/repository"
	"go-gin-layered-architecture/internal/scheduler"
	"go-gin-layered-architecture/internal/service"
	"go-gin-layered-architecture/internal/websocket"
	"time"

	_ "go-gin-layered-architecture/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func main() {
	_ = godotenv.Load()
	appConfig := config.AppLoad()
	db := config.DBLoad(appConfig.DBDsn)
	defer db.Close()
	docsConfig := config.DocsLoad()
	geminiConfig := config.GeminiLoad()
	bunnyConfig := config.BunnyLoad()

	healthHandler := handler.NewHealthHandler()
	docsHandler := handler.NewDocsHandler(docsConfig)

	userLogger := helper.Logger("user")
	userRepository := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(userService, userLogger)

	authLogger := helper.Logger("auth")
	revokedRefreshTokenRepository := repository.NewRevokedRefreshTokenRepository(db)
	authService := service.NewAuthService(userRepository, revokedRefreshTokenRepository, appConfig.JWTSecret)
	authHandler := handler.NewAuthHandler(authService, authLogger)

	thirdPartyLogger := helper.Logger("third_party")
	thirdPartyKeyRepository := repository.NewThirdPartyKeyRepository(db)
	authMiddleware := middleware.NewAuthMiddleware(appConfig.JWTSecret, thirdPartyKeyRepository, thirdPartyLogger)

	chatbotLogger := helper.Logger("chatbot")
	geminiService := service.NewGeminiService(geminiConfig)
	chatbotHandler := handler.NewChatbotHandler(geminiService, chatbotLogger)

	bunnyLogger := helper.Logger("bunny")
	bunnyService := service.NewBunnyService(bunnyConfig)
	bunnyHandler := handler.NewBunnyHandler(bunnyService, bunnyLogger)

	logLogger := helper.Logger("logs")
	logRepository := repository.NewLogRepository()
	logService := service.NewLogService(logRepository)
	logHandler := handler.NewLogHandler(logService, logLogger)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))
	r.GET("/docs", docsHandler.Auth(), docsHandler.GetDocs)
	r.GET("/api/health", healthHandler.GetHealth)
	{
		api := r.Group("api")
		{
			auth := api.Group("auth")
			auth.POST("login", authHandler.Login)
			auth.POST("register", authHandler.Register)
			auth.POST("refresh", authHandler.Refresh)
			{
				authAuthorized := auth.Group("")
				authAuthorized.Use(authMiddleware.RequiredAuth())
				authAuthorized.GET("me", authHandler.Me)
				authAuthorized.POST("logout", authHandler.Logout)
				authAuthorized.PUT("profile", authHandler.UpdateProfile)
				authAuthorized.PUT("password", authHandler.ChangePassword)
			}
		}
		{
			users := api.Group("users")
			{
				authUsers := users.Group("")
				authUsers.Use(authMiddleware.RequiredAuth("jwt", "third_party:smartchat,tes"))
				authUsers.GET("", userHandler.ListUsers)
				authUsers.GET(":id", userHandler.GetById)
				authUsers.POST("", userHandler.CreateUser)
				authUsers.PUT(":id", userHandler.UpdateUser)
				authUsers.DELETE(":id", userHandler.DeleteUser)
			}
		}
		{
			chatbot := api.Group("chatbot")
			{
				authChatbot := chatbot.Group("")
				authChatbot.POST("ask", chatbotHandler.Ask)
				authChatbot.POST("generate-image", chatbotHandler.GenerateImage)
			}
		}
		{
			files := api.Group("files")
			{
				authFiles := files.Group("")
				authFiles.Use(authMiddleware.RequiredAuth())
				authFiles.POST("", bunnyHandler.UploadFile)
				authFiles.DELETE("", bunnyHandler.DeleteFile)
			}
		}
		{
			logs := api.Group("logs")
			logs.GET("", logHandler.GetAllLogs)
			logs.GET(":file_name", logHandler.GetLogsByFileName)
		}
	}

	testingJob := scheduler.NewTestingJob()

	// Scheduler
	scheduler := cron.New()
	scheduler.AddFunc("* * * * *", testingJob.Run)
	scheduler.Start()
	defer scheduler.Stop()

	// Websocket
	r.GET("/ws/testing", websocket.HandleTestingConnections)
	go websocket.HandleTestingBroadcast()

	r.Run(":" + appConfig.Port)
}
