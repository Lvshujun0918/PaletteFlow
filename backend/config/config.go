package config

import (
	"os"
	"strconv"

	"ai-color-palette/logging"

	"github.com/joho/godotenv"
)

type Config struct {
	AIAPIKey     string
	AIAPIBaseURL string
	AIModel      string
	AITimeout    int
}

var AppConfig *Config

func LoadConfig() {
	// 加载.env文件
	if err := godotenv.Load(); err != nil {
		logging.Warn("config.env", ".env file not found, using environment variables", nil)
	} else {
		logging.Info("config.env", ".env loaded", nil)
	}

	timeout := 30
	if timeoutStr := os.Getenv("AI_TIMEOUT"); timeoutStr != "" {
		if val, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = val
		}
	}

	AppConfig = &Config{
		AIAPIKey:     os.Getenv("AI_API_KEY"),
		AIAPIBaseURL: getEnvOrDefault("AI_API_BASE_URL", "https://open.bigmodel.cn/api/paas/v4"),
		AIModel:      getEnvOrDefault("AI_MODEL", "glm-4.7-flash"),
		AITimeout:    timeout,
	}

	logging.Info("config.values", "config values prepared", logging.Fields{
		"ai_base_url": AppConfig.AIAPIBaseURL,
		"ai_model":    AppConfig.AIModel,
		"ai_timeout":  AppConfig.AITimeout,
		"has_api_key": AppConfig.AIAPIKey != "",
	})

	if AppConfig.AIAPIKey == "" {
		logging.Error("config.validation", "AI_API_KEY is not set in environment variables", nil)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	logging.Info("config.default", "env key is empty, fallback to default", logging.Fields{
		"key":     key,
		"default": defaultValue,
	})
	return defaultValue
}
