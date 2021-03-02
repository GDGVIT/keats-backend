package endpoints

import (
	"net/http"
	"context"

	"github.com/go-pg/pg/v10"
	"github.com/gofiber/fiber/v2"
	jwt "github.com/form3tech-oss/jwt-go"

	"github.com/Krishap-s/keats-backend/firebaseclient"
	"github.com/Krishap-s/keats-backend/crud"
	"github.com/Krishap-s/keats-backend/schemas"
	"github.com/Krishap-s/keats-backend/configs"
	"github.com/Krishap-s/keats-backend/errors"
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

func createUser(c *fiber.Ctx) error {
	req := new(createUserRequest)
	client ,err := firebaseclient.GetClient()
	if err != nil {
		return errors.InternalServerError(c, "")
	}
	err = c.BodyParser(req);if err != nil {
		return errors.BadRequestError(c, "Missing or Malformed IDToken")
	}
	firetoken, err := client.VerifyIDToken(context.Background(),req.IDToken)
	if err != nil {
		return errors.UnauthorizedError(c, "IDToken Verification failed or IDToken expired")
	}
	phone_number, ok := firetoken.Claims["phone_number"].(string)
	if !ok {
		return errors.BadRequestError(c,"IDToken missing phone_number")
	}
	user := &schemas.UserCreate{
		PhoneNo: phone_number,
	}

	created, err := crud.CreateUser(user)
	if err != nil{
		return errors.InternalServerError(c,"")
	}

	token := jwt.New(jwt.SigningMethodRS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = created.Username
	claims["phone_number"] = created.PhoneNo
	claims["email"] = created.Email
	claims["bio"] = created.Bio
	claims["profile_pic"] = created.ProfilePic

	signedtoken,err := token.SignedString(configs.GetKey())

	if err != nil {
		return errors.InternalServerError(c, "")
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"data": fiber.Map{"user":created,"jwt_token":signedtoken},
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

	updated, err := crud.UpdateUser(u)
	if err != nil {
		err = userErrHandler(c, err)
		return err
	}

	return c.JSON(fiber.Map{
		"msg":  "user updated successfully",
		"user": updated,
	})
}

func getUser(c *fiber.Ctx) error {
	userID := c.Params("id")
	user, err := crud.GetUser(userID)
	if err != nil {
		err = userErrHandler(c, err)
		return err
	}

	return c.JSON(user)
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
	app.Get("/api/user/:id", getUser)
}
