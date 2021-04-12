package configs

import "github.com/gofiber/fiber/v2"

func FiberConfig() fiber.Config {
	return fiber.Config{
		BodyLimit: 30 * 1024 * 1024,
	}
}
