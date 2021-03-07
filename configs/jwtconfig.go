package configs

//goland:noinspection SpellCheckingInspection
import (
	"fmt"
	"log"

	jwt "github.com/form3tech-oss/jwt-go"
	"github.com/go-pg/pg/v10"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v2"
	"github.com/spf13/viper"

	"github.com/Krishap-s/keats-backend/crud"
	"github.com/Krishap-s/keats-backend/errors"
)

func GetSecret() string {
	secret, ok := viper.Get("JWT_SECRET").(string)
	if !ok {
		log.Panic(fmt.Errorf("jwt secret not found"))
	}
	return secret
}

var JWTConfig = jwtware.Config{
	ErrorHandler: func(c *fiber.Ctx, err error) error {
		if err.Error() == "Missing or malformed JWT" {
			return errors.BadRequestError(c, "Missing or malformed JWT")
		} else {
			return errors.UnauthorizedError(c, "Invalid JWT")
		}
	},
	SuccessHandler: func(c *fiber.Ctx) error {
		token := c.Locals("user").(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)
		id := claims["id"].(string)
		user, err := crud.GetUser(id)
		if err != nil {
			if err == pg.ErrNoRows {
				return errors.UnauthorizedError(c, "Invalid JWT")
			} else {
				return errors.InternalServerError(c, "")
			}
		}
		c.Locals("user", user)
		return c.Next()
	},
	SigningMethod: "HS256",
}
