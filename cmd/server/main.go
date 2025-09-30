package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/YaoAzure/wsgateway/pkg/config"
	"github.com/YaoAzure/wsgateway/pkg/jwt"
	"github.com/YaoAzure/wsgateway/pkg/log"
	"github.com/YaoAzure/wsgateway/pkg/redis"
	"github.com/YaoAzure/wsgateway/pkg/session"
	"github.com/gofiber/fiber/v3"
	"github.com/samber/do/v2"
)

func main() {
	// Parse command line flags
	configPath := parseFlags()

	// Load configuration first
	loader := config.NewLoader(configPath)
	conf, err := loader.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load configuration: %v", err))
	}

	// Create DI container with all packages
	injector := do.New(
		config.NewPackage(conf), // 配置包 - 使用 Eager Loading
		log.Package,             // Log 包 - 使用 Lazy Loading
		redis.Package,           // Redis 包 - 使用 Lazy Loading
		jwt.Package,             // JWT 包 - 使用 Lazy Loading
		session.Package,         // Session 包 - 使用 Lazy Loading
	)
	defer injector.Shutdown()

	// Get configured logger from DI container
	logger, err := do.Invoke[*log.Logger](injector)
	if err != nil {
		panic(fmt.Sprintf("Failed to get logger from DI container: %v", err))
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: conf.App.Name,
	})

	// healty check
	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Start server
	logger.Info("Starting server", "service", conf.App.Name, "addr", conf.App.Addr)
	if err := app.Listen(conf.App.Addr); err != nil {
		logger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}

// parseFlags 解析命令行参数并返回配置文件路径
func parseFlags() string {
	var configPath = flag.String("config", "configs/config.yaml", "配置文件路径")
	var showHelp = flag.Bool("help", false, "显示帮助信息")
	flag.Parse()

	// Show help if requested
	if *showHelp {
		flag.Usage()
		return ""
	}

	return *configPath
}
