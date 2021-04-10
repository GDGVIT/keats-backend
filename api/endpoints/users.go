package endpoints

import (
	"context"
	jwt "github.com/form3tech-oss/jwt-go"
	"github.com/go-pg/pg/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/Krishap-s/keats-backend/configs"
	"github.com/Krishap-s/keats-backend/crud"
	"github.com/Krishap-s/keats-backend/errors"
	"github.com/Krishap-s/keats-backend/firebaseclient"
	"github.com/Krishap-s/keats-backend/models"
	"github.com/Krishap-s/keats-backend/schemas"
)

type IDTokenRequest struct {
	IDToken string `json:"id_token"`
}

func createJWT(user *models.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = user.ID
	signedtoken, err := token.SignedString([]byte(configs.GetSecret()))
	if err != nil {
		return "", err
	}
	return signedtoken, nil
}

func createUser(c *fiber.Ctx) error {
	req := new(IDTokenRequest)
	client, err := firebaseclient.GetClient()
	if err != nil {
		return errors.InternalServerError(c, "")
	}
	err = c.BodyParser(req)
	if err != nil {
		return errors.BadRequestError(c, "Missing or Malformed IDToken")
	}
	firetoken, err := client.VerifyIDToken(context.Background(), req.IDToken)
	if err != nil {
		return errors.UnauthorizedError(c, "IDToken Verification failed or IDToken expired")
	}
	phone_number, ok := firetoken.Claims["phone_number"].(string)
	if !ok {
		return errors.BadRequestError(c, "IDToken missing phone_number")
	}
	u := &schemas.UserCreate{
		PhoneNo: phone_number,
	}

	created, err := crud.CreateUser(u)
	if err != nil {
		return errors.InternalServerError(c, "")
	}

	signedtoken, err := createJWT(created)
	if err != nil {
		return errors.InternalServerError(c, "")
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"token":   signedtoken,
			"user_id": created.ID,
		},
	})
}

func updateUser(c *fiber.Ctx) error {
	u := new(schemas.UserUpdate)

	if err := c.BodyParser(u); err != nil {
		return errors.UnprocessableEntityError(c, "JSON in the incorrect format")
	}

	user := c.Locals("user").(*models.User)
	uidBytes, err := user.ID.MarshalText()
	if err != nil {
		return errors.InternalServerError(c, "")
	}
	u.ID = string(uidBytes)
	u.ProfilePic = user.ProfilePic
	u.PhoneNo = user.PhoneNo

	updated, err := crud.UpdateUser(u)
	if err != nil {
		return errors.InternalServerError(c, err.Error())
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   updated,
	})
}

func updateUserPhoneNo(c *fiber.Ctx) error {
	req := new(IDTokenRequest)
	client, err := firebaseclient.GetClient()
	if err != nil {
		return errors.InternalServerError(c, "")
	}
	err = c.BodyParser(req)
	if err != nil {
		return errors.BadRequestError(c, "Missing or Malformed IDToken")
	}
	firetoken, err := client.VerifyIDToken(context.Background(), req.IDToken)
	if err != nil {
		return errors.UnauthorizedError(c, "IDToken Verification failed or IDToken expired")
	}
	phone_number, ok := firetoken.Claims["phone_number"].(string)
	if !ok {
		return errors.BadRequestError(c, "IDToken missing phone_number")
	}
	user := c.Locals("user").(*models.User)
	u := &schemas.UserUpdate{
		PhoneNo: phone_number,
	}
	uidBytes, err := user.ID.MarshalText()
	if err != nil {
		return errors.InternalServerError(c, "")
	}
	u.ID = string(uidBytes)
	_, err = crud.UpdateUser(u)
	if err != nil {
		pgerr := err.(pg.Error)
		if pgerr.IntegrityViolation() {
			return errors.ConflictError(c, "phone number already exists")
		}
		return errors.InternalServerError(c, "")
	}
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Phone number updated",
	})

}

func getUser(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   c.Locals("user"),
	})
}

func getUserClubsAndDetails(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	uidBytes, err := user.ID.MarshalText()
	if err != nil {
		return errors.InternalServerError(c, "")
	}
	uid := string(uidBytes)
	clubs, err := crud.GetUserClub(uid)
	if err != nil {
		if err == pg.ErrNoRows {
			return errors.NotFoundError(c, "No clubs found")
		}
		return errors.InternalServerError(c, err.Error())
	}
	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"clubs": clubs,
			"user":  c.Locals("user"),
		},
	})

}

// MountUserRoutes mounts all routes declared here
func MountUserRoutes(app *fiber.App, middleware func(c *fiber.Ctx) error) {
	app.Post("/api/user", createUser)
	authGroup := app.Group("/api/", middleware)
	authGroup.Patch("user", updateUser)
	authGroup.Post("user/updatephone", updateUserPhoneNo)
	authGroup.Get("user", getUser)
	authGroup.Get("user/clubs", getUserClubsAndDetails)
}
