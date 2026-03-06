package main

import (
	"log"
	"runtime/debug"

	"ai-color-palette/config"
	"ai-color-palette/handler"
	"ai-color-palette/logging"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			logging.Error("server.panic", "panic recovered in main", logging.Fields{"panic": r})
			debug.PrintStack()
		}
	}()
	log.Println("8888888b.          888          888    888            8888888888 888                        ")
	log.Println("888   Y88b         888          888    888            888        888                        ")
	log.Println("888    888         888          888    888            888        888                        ")
	log.Println("888   d88P 8888b.  888  .d88b.  888888 888888 .d88b.  8888888    888  .d88b.  888  888  888 ")
	log.Println("8888888P       88b 888 d8P  Y8b 888    888   d8P  Y8b 888        888 d88  88b 888  888  888 ")
	log.Println("888       .d888888 888 88888888 888    888   88888888 888        888 888  888 888  888  888 ")
	log.Println("888       888  888 888 Y8b.     Y88b.  Y88b. Y8b.     888        888 Y88..88P Y88b 888 d88P ")
	log.Println("888        Y888888 888   Y8888    Y888   Y888  Y8888  888        888   Y88P     Y8888888P   ")
	log.Println("                                                                                            ")
	logging.Info("server.start", "backend process starting", logging.Fields{"port": 5208})
	// 加载配置
	config.LoadConfig()
	logging.Info("config.loaded", "config loaded", nil)
	// 设置Gin为发布模式
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(logging.RequestLoggerMiddleware())
	// 配置CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173", "http://localhost:3000", "*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	router.Use(cors.New(config))
	logging.Info("server.routes", "routes registered", logging.Fields{
		"routes": []string{
			"GET /api/health",
			"POST /api/generate-palette",
			"POST /api/refine-palette",
			"POST /api/regenerate-color",
			"POST /api/apply-image-palette",
			"GET /api/apply-image-palette/task/:taskId",
			"GET /api/apply-image-palette/task/:taskId/result",
		},
	})
	router.GET("/api/health", handler.HealthHandler)
	router.POST("/api/generate-palette", handler.GeneratePaletteHandler)
	router.POST("/api/refine-palette", handler.RefinePaletteHandler)
	router.POST("/api/regenerate-color", handler.RegenerateSingleColorHandler)
	router.POST("/api/apply-image-palette", handler.ApplyImagePaletteHandler)
	router.GET("/api/apply-image-palette/task/:taskId", handler.GetImagePaletteTaskHandler)
	router.GET("/api/apply-image-palette/task/:taskId/result", handler.DownloadImagePaletteTaskResultHandler)
	logging.Info("server.ready", "gin server ready", logging.Fields{"port": 5208})
	logging.Info("server.listen", "gin server listening", logging.Fields{"addr": ":5208"})
	if err := router.Run(":5208"); err != nil {
		logging.Error("server.exit", "server failed to start", logging.Fields{"error": err.Error()})
		log.Fatalf("server failed: %v", err)
	}
	logging.Info("server.stop", "server stopped", nil)
}
