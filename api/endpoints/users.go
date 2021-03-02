package endpoints

import (
	"context"
	"net/http"

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

func userErrHandler(c *fiber.Ctx, err error) error {
	if err == pg.ErrNoRows {
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"msg": "user with this username already exists",
		})
	}

	return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
		"msg": "something went wrong",
		"err": err.Error(),
	})

}

type createUserRequest struct {
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
	req := new(createUserRequest)
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
		"data":   signedtoken,
	})
}

func updateUser(c *fiber.Ctx) error {
	u := new(schemas.UserUpdate)

	if err := c.BodyParser(u); err != nil {
		return c.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
			"msg": "invalid json",
			"err": err.Error(),
		})
	}

	user := c.Locals("user").(*models.User)
	u.PhoneNo = user.PhoneNo

	updated, err := crud.UpdateUser(u)
	if err != nil {
		return errors.InternalServerError(c, "")
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   updated,
	})
}

func getUser(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   c.Locals("user"),
	})
}

func deleteUser(c *fiber.Ctx) error {
	userID := c.Params("id")
	deleted, err := crud.DeleteUser(userID)
	if err != nil {
		err = userErrHandler(c, err)
		return err
	}

	return c.JSON(fiber.Map{
		"msg":  "successfully deleted user",
		"user": deleted,
	})
}

// MountRoutes mounts all routes declared here
func MountRoutes(app *fiber.App) {
	app.Post("/api/user", createUser)
	app.Patch("/api/user", updateUser)
	app.Delete("/api/user", deleteUser)
	app.Get("/api/user", getUser)
}
