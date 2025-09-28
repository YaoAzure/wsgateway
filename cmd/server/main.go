package main

import (
	"log"

	"github.com/YaoAzure/wsgateway/pkg/config"
	"github.com/gofiber/fiber/v3"
	"github.com/samber/do/v2"
)

func main() {
	// Create DI container
	injector := do.New()
	defer injector.Shutdown()

	loader := config.NewLoader("")
	config, err := loader.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: config.App.Name,
	})

	// healty check
	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})


	// Start server
	log.Printf("Starting %s server on %s", config.App.Name, config.App.Addr)
	if err := app.Listen(config.App.Addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
