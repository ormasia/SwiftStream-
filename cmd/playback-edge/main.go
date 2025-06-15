package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ormasia/swiftstream/internal/router"
)

func main() {
	app := fiber.New(fiber.Config{
		AppName: "SwiftStream",
	})

	router.RegisterRoutes(app, router.Deps{})

	app.Listen(":3000")
}
