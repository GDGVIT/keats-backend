package errors

import (
	"github.com/gofiber/fiber/v2"
)

func TooManyRequestsError(c *fiber.Ctx) error {
	return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
		"status":"error",
		"error":"Too Many Requests",
	})
}

