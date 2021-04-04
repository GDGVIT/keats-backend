package endpoints

import (
	"github.com/Krishap-s/keats-backend/crud"
	"github.com/Krishap-s/keats-backend/errors"
	"github.com/Krishap-s/keats-backend/models"
	"github.com/Krishap-s/keats-backend/schemas"
	"github.com/go-pg/pg/v10"
	"github.com/gofiber/fiber/v2"
)

type clubRequests struct {
	ClubID string `json:"club_id"`
}

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
	r := new(clubRequests)
	if err := c.BodyParser(r); err != nil || r.ClubID == "" {
		return errors.UnprocessableEntityError(c, "JSON in the incorrect format")
	}
	clubId := r.ClubID
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
			"club":     club,
			"users":    users,
			"comments": "{}",
			"chat":     "{}",
		},
		"message": "Club joined successfully",
	})
}

func getClub(c *fiber.Ctx) error {
	clubId := c.Query("club_id")
	users, err := crud.GetClubUser(clubId)
	user := c.Locals("user").(*models.User)
	var isMember = false
	for _, clubUser := range users {
		if clubUser.ID == user.ID {
			isMember = true
			break
		}
	}
	if !isMember {
		return errors.UnauthorizedError(c, "You are not a member of this club")
	}
	if err != nil {
		return errors.InternalServerError(c, err.Error())
	}
	club, err := crud.GetClub(clubId)
	if err != nil {
		return errors.InternalServerError(c, err.Error())
	}
	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"club":     club,
			"users":    users,
			"comments": "{}",
			"chat":     "{}",
		},
	})
}

func updateClub(c *fiber.Ctx) error {
	r := new(schemas.ClubUpdate)
	if err := c.BodyParser(r); err != nil || r.ID == "" {
		return errors.UnprocessableEntityError(c, "JSON in the incorrect format")
	}
	clubId := r.ID
	club, err := crud.GetClub(clubId)
	if err != nil {
		return errors.NotFoundError(c, "Club not found")
	}
	user := c.Locals("user").(*models.User)
	uidBytes, err := user.ID.MarshalText()
	if err != nil {
		return errors.InternalServerError(c, "")
	}
	uid := string(uidBytes)
	if uid != club.HostID {
		return errors.UnauthorizedError(c, "You are not the host of this group")
	}
	updated, err := crud.UpdateClub(r)
	if err != nil {
		return errors.InternalServerError(c, "")
	}
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   updated,
	})
}

func kickUser(c *fiber.Ctx) error {
	r := new(struct {
		UserID string `json:"user_id"`
		ClubID string `json:"club_id"`
	})
	if err := c.BodyParser(r); err != nil || r.UserID == "" || r.ClubID == "" {
		return errors.UnprocessableEntityError(c, "JSON in the incorrect format")
	}
	clubId := r.ClubID
	deviantId := r.UserID
	club, err := crud.GetClub(clubId)
	if err != nil {
		return errors.NotFoundError(c, "Club not found")
	}
	user := c.Locals("user").(*models.User)
	uidBytes, err := user.ID.MarshalText()
	if err != nil {
		return errors.InternalServerError(c, "")
	}
	uid := string(uidBytes)
	if uid != club.HostID {
		return errors.UnauthorizedError(c, "You are not the host of this group")
	} else if uid == deviantId {
		return errors.ConflictError(c, "You cannot kick yourself out of the club")
	}
	_, err = crud.DeleteClubUser(clubId, r.UserID)
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "User has been kicked from the club",
	})
}

func leaveClub(c *fiber.Ctx) error {
	r := new(clubRequests)
	if err := c.BodyParser(r); err != nil || r.ClubID == "" {
		return errors.UnprocessableEntityError(c, "JSON in the incorrect format")
	}
	clubId := r.ClubID
	user := c.Locals("user").(*models.User)
	uidBytes, err := user.ID.MarshalText()
	if err != nil {
		return errors.InternalServerError(c, "")
	}
	_, err = crud.DeleteClubUser(clubId, string(uidBytes))
	if err != nil {
		if err == pg.ErrNoRows {
			return errors.ConflictError(c, "You are not a member of this club")
		}
		return errors.InternalServerError(c, err.Error())
	}
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "You have left the club",
	})
}

func MountClubRoutes(app *fiber.App, middleware func(c *fiber.Ctx) error) {
	authGroup := app.Group("/api/clubs", middleware)
	authGroup.Post("createclub", createClub)
	authGroup.Post("joinclub", joinClub)
	authGroup.Get("", getClub)
	authGroup.Patch("updateclub", updateClub)
	authGroup.Post("kickuser", kickUser)
	authGroup.Post("leaveclub", leaveClub)
}
