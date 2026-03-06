package main

import (
	"log"
	"runtime/debug"

	"ai-color-palette/config"
	"ai-color-palette/handler"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[PANIC] %v", r)
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
	// 加载配置
	config.LoadConfig()
	log.Println("[INFO] Config loaded")
	// 设置Gin为发布模式
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	// 配置CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173", "http://localhost:3000", "*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	router.Use(cors.New(config))
	router.GET("/api/health", handler.HealthHandler)
	router.POST("/api/generate-palette", handler.GeneratePaletteHandler)
	router.POST("/api/refine-palette", handler.RefinePaletteHandler)
	router.POST("/api/regenerate-color", handler.RegenerateSingleColorHandler)
	log.Println("[INFO] GIN Server ready")
	log.Println("[INFO] GIN Server starting on :5208")
	if err := router.Run(":5208"); err != nil {
		log.Fatalf("[FATAL] Server failed: %v", err)
	}
	log.Println("[INFO] Bind Ready")
}
