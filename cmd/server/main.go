package main

import (
	"log"

	"github.com/YaoAzure/wsgateway/pkg/config"
	"github.com/YaoAzure/wsgateway/pkg/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/samber/do/v2"
)

func main() {
	// Create DI container
	injector := do.New()
	defer injector.Shutdown()

	loader := config.NewLoader("")
	conf, err := loader.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	do.ProvideValue(injector, conf)
	config.RegisterConfigService(injector, conf)
	jwt.RegisterJWTService(injector)
	
	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: conf.App.Name,
	})

	// healty check
	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})


	// Start server
	log.Printf("Starting %s server on %s", conf.App.Name, conf.App.Addr)
	if err := app.Listen(conf.App.Addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
