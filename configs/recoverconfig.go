package configs

import (
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var RecoverConfig = recover.Config{
	EnableStackTrace: true,
}
