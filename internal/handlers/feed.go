package handlers

import "github.com/gofiber/fiber/v2"

func GetFeed(c *fiber.Ctx) error {
	// Logic to get feed
	return c.SendString("Get Feed")
}
