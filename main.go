package main

//goland:noinspection SpellCheckingInspection
import (
	"fmt"
	"github.com/Krishap-s/keats-backend/api/sockets"
	"log"

	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	jwtware "github.com/gofiber/jwt/v2"
	"github.com/spf13/viper"

	"github.com/Krishap-s/keats-backend/api/endpoints/clubs"
	"github.com/Krishap-s/keats-backend/api/endpoints/users"
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

	app := fiber.New(configs.FiberConfig())

	// Use Middleware
	app.Use(limiter.New(configs.LimiterConfig()))
	app.Use(logger.New(configs.LoggerConfig()))
	app.Use(recover.New(configs.RecoverConfig()))
	app.Use(cors.New(configs.CORSConfig()))

	// Setting up jwt config
	jwtconf := configs.JWTConfig()

	app.Get("/", healthCheck)

	// Run pgdb migrations
	log.Println("Running database migrations")
	if err := pgdb.Migrate(); err != nil {
		log.Panic(err)
	}

	users.MountRoutes(app, jwtware.New(jwtconf))
	clubs.MountRoutes(app, jwtware.New(jwtconf))
	sockets.MountWebsockets(app, jwtware.New(jwtconf))

	if err := app.Listen("0.0.0.0:" + viper.GetString("PORT")); err != nil {
		log.Panic(err)
	}
}
