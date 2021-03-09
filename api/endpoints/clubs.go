package endpoints

import (
	"github.com/Krishap-s/keats-backend/crud"
	"github.com/Krishap-s/keats-backend/errors"
	"github.com/Krishap-s/keats-backend/models"
	"github.com/Krishap-s/keats-backend/schemas"
	"github.com/go-pg/pg/v10"
	"github.com/gofiber/fiber/v2"
)

func createClub(c *fiber.Ctx) error {
	club := new(schemas.ClubCreate)
	if err := c.BodyParser(club); err != nil {
		return errors.UnprocessableEntityError(c, "form data in the incorrect format")
	}
	user := c.Locals("user").(*models.User)
	uidBytes, err := user.ID.MarshalText()
	if err != nil {
		return errors.InternalServerError(c, "")

	}
	club.HostID = string(uidBytes)
	created, err := crud.CreateClub(club)
	if err != nil {
		return errors.InternalServerError(c, err.Error())
	}
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   created,
	})
}

func joinClub(c *fiber.Ctx) error {
	clubId := c.Params("club_id")
	club, err := crud.GetClub(clubId)
	if err != nil {
		return errors.NotFoundError(c, "Club not found")
	}
	users, err := crud.GetClubUser(clubId)
	user := c.Locals("user").(*models.User)
	uidBytes, err := user.ID.MarshalText()
	if err != nil {
		return errors.InternalServerError(c, "")
	}
	_, err = crud.CreateClubUser(clubId, string(uidBytes))
	if err != nil {
		pgerr := err.(pg.Error)
		if pgerr.IntegrityViolation() {
			return errors.ConflictError(c, "You are already a member of this club")
		}
		return errors.InternalServerError(c, "")
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"club":  club,
			"users": users,
		},
		"message": "Club joined successfully",
	})
}

func MountClubRoutes(app *fiber.App, middleware func(c *fiber.Ctx) error) {
	authGroup := app.Group("/api/clubs", middleware)
	authGroup.Post("", createClub)
	authGroup.Post("joinclub/:club_id", joinClub)
}