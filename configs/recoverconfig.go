package configs

import (
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func RecoverConfig() recover.Config {
	return recover.Config{
		EnableStackTrace: true,
	}
}
