package main

import (
	"log"
	"os"
	"os/signal"
	"search-ec2/internal/config"
	"search-ec2/internal/handlers"
	"search-ec2/internal/services"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// 加载配置
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 设置日志
	setupLogger()

	logrus.Info("Search EC2 - Natural Language Product Search System")
	logrus.Info("Starting server...")

	// 初始化服务
	serviceManager, err := services.NewServiceManager()
	if err != nil {
		logrus.Fatalf("Failed to initialize services: %v", err)
	}

	// 设置优雅关闭
	defer func() {
		if err := serviceManager.Close(); err != nil {
			logrus.Errorf("Error closing services: %v", err)
		}
	}()

	// 设置 Gin 模式
	gin.SetMode(config.AppConfig.Server.Mode)

	// 创建 Gin 引擎
	r := gin.New()

	// 设置中间件
	handlers.SetupMiddleware(r)

	// 设置路由（传入服务管理器）
	handlers.SetupRoutesWithServices(r, serviceManager)

	// 启动服务器
	address := config.AppConfig.GetAddress()
	logrus.Infof("Server starting on %s", address)

	// 设置信号处理
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logrus.Info("Received shutdown signal")
		os.Exit(0)
	}()

	if err := r.Run(address); err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}

// setupLogger 设置日志
func setupLogger() {
	// 设置日志级别
	level, err := logrus.ParseLevel(config.AppConfig.Logging.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// 设置日志格式
	if config.AppConfig.Logging.Format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	logrus.Info("Logger initialized")
}
