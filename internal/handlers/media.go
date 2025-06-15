package handlers

import "github.com/gofiber/fiber/v2"

func GetMedia(c *fiber.Ctx) error {
	// Logic to get media
	return c.SendString("Get Media")
}
