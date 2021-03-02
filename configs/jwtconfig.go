package configs

import(
	"io/ioutil"
	"crypto/rsa"

	"golang.org/x/crypto/pkcs12"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v2"
	jwt "github.com/form3tech-oss/jwt-go"
	"github.com/go-pg/pg/v10"
	"github.com/spf13/viper"

	"github.com/Krishap-s/keats-backend/crud"
	"github.com/Krishap-s/keats-backend/errors"
)

var key *rsa.PrivateKey = nil

// Get private key from pfx file
func GetKey() *rsa.PrivateKey {
	if key != nil {
		return key
	}
	pfxData, err := ioutil.ReadFile(viper.GetString("PFX_FILE_LOCATION"))
	if err != nil {
		panic(err)
	}
	keyData, _,err := pkcs12.Decode(pfxData,viper.GetString("PFX_FILE_PASSWORD"))
	if err !=nil {
		panic(err)
	}
	key = keyData.(*rsa.PrivateKey)
	return key
}


var JWTConfig= jwtware.Config{
	Filter:func(c *fiber.Ctx) bool{
		if c.Method() == "POST" && c.Path() == "/api/user"{
			return true
		}
		return false
	},
	ErrorHandler: func(c *fiber.Ctx,err error) error {
			if err.Error() == "Missing or malformed JWT"{
				return errors.BadRequestError(c,"Missing or malformed JWT")
			} else {
				return errors.UnauthorizedError(c,"Invalid JWT")
			}
		},
	SuccessHandler: func(c *fiber.Ctx) error {
				token := c.Locals("user").(*jwt.Token)
				claims := token.Claims.(jwt.MapClaims)
				phone_no := claims["phone_number"].(string)
				user ,err:= crud.GetUser(phone_no)
				if err != nil {
					if err == pg.ErrNoRows{
						return errors.UnauthorizedError(c,"Invalid JWT")
					} else {
					return errors.InternalServerError(c,"")
				}
				}
				c.Locals("user",user)
				return c.Next()
			},
	SigningMethod: "RS256",
}


