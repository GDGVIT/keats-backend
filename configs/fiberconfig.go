package configs

import (
	"github.com/Krishap-s/keats-backend/errors"
	"github.com/gofiber/fiber/v2"
)

func FiberConfig() fiber.Config {
	return fiber.Config{
		BodyLimit:    30 * 1024 * 1024,
		ErrorHandler: errors.ErrorHandler,
	}
}
