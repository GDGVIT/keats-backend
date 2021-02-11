package endpoints

import (
	"net/http"

	"github.com/go-pg/pg/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/Krishap-s/keats-backend/crud"
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

func createUser(c *fiber.Ctx) error {
	u := new(schemas.UserCreate)

	if err := c.BodyParser(u); err != nil {
		return c.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
			"msg": "invalid json",
			"err": err.Error(),
		})
	}

	created, err := crud.CreateUser(u)
	if err != nil {
		err = userErrHandler(c, err)
		return err
	}

	return c.JSON(fiber.Map{
		"msg":  "user created successfully",
		"user": created,
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
