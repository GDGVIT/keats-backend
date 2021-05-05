package configs

//goland:noinspection SpellCheckingInspection
import (
	"fmt"
	"github.com/gofiber/websocket/v2"
	"log"

	jwt "github.com/form3tech-oss/jwt-go"
	"github.com/go-pg/pg/v10"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v2"
	"github.com/spf13/viper"

	"github.com/Krishap-s/keats-backend/crud"
)

func GetSecret() string {
	secret, ok := viper.Get("JWT_SECRET").(string)
	if !ok {
		log.Panic(fmt.Errorf("jwt secret not found"))
	}
	return secret
}

func JWTConfig() jwtware.Config {
	return jwtware.Config{
		Filter: func(c *fiber.Ctx) bool {
			if websocket.IsWebSocketUpgrade(c) {
				return true
			}
			return false
		},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if err.Error() == "Missing or malformed JWT" {
				return fmt.Errorf("malformed jwt")
			}

			return fmt.Errorf("invalid jwt")
		},
		SuccessHandler: func(c *fiber.Ctx) error {
			token := c.Locals("user").(*jwt.Token)
			claims := token.Claims.(jwt.MapClaims)
			id := claims["id"].(string)
			user, err := crud.GetUser(id)
			if err != nil {
				if err == pg.ErrNoRows {
					return fmt.Errorf("invalid jwt")
				}
				return err
			}
			c.Locals("user", user)
			return c.Next()
		},
		SigningKey:    []byte(GetSecret()),
		SigningMethod: "HS256",
	}
}
