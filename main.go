package main

//goland:noinspection SpellCheckingInspection
import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	jwtware "github.com/gofiber/jwt/v2"
	"github.com/spf13/viper"

	"github.com/Krishap-s/keats-backend/api/endpoints"
	"github.com/Krishap-s/keats-backend/configs"
	"github.com/Krishap-s/keats-backend/pgdb"
)

func healthCheck(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func main() {
	// Set global configuration
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Panicln(fmt.Errorf("fatal error config file: %s", err))
	}

	app := fiber.New()

	// Use Middleware
	app.Use(limiter.New(configs.LimiterConfig))
	app.Use(logger.New(configs.LoggerConfig))
	app.Use(recover.New(configs.RecoverConfig))

	// Set Up JWT middleware
	jwtconf := configs.JWTConfig
	jwtconf.SigningKey = []byte(configs.GetSecret())
	app.Use(jwtware.New(jwtconf))

	app.Get("/", healthCheck)

	// Run pgdb migrations
	log.Println("Running database migrations")
	if err := pgdb.Migrate(); err != nil {
		log.Panic(err)
	}

	endpoints.MountRoutes(app)

	if err := app.Listen(":3000"); err != nil {
		log.Panic(err)
	}
}
