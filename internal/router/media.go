package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ormasia/swiftstream/internal/handlers"
)

func RegMediaRoutes(router fiber.Router) {
	router.Get("/media", handlers.GetMedia)
}
