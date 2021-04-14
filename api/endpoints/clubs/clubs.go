package clubs

import (
	"github.com/Krishap-s/keats-backend/api/endpoints/users"
	"github.com/Krishap-s/keats-backend/crud"
	"github.com/Krishap-s/keats-backend/errors"
	"github.com/Krishap-s/keats-backend/models"
	"github.com/Krishap-s/keats-backend/schemas"
	"github.com/go-pg/pg/v10"
	"github.com/gofiber/fiber/v2"
)

// Non Handlers

type clubRequests struct {
	ID string `json:"club_id"`
}

func parseClubIDRequest(c *fiber.Ctx) (*clubRequests, error) {
	r := new(clubRequests)
	if err := c.BodyParser(r); err != nil || r.ID == "" {
		return nil, errors.BadRequestError(c, "JSON in the incorrect format")
	}
	return r, nil
}

func checkIfHost(userID string, clubID string) (bool, error) {
	club, err := crud.GetClub(clubID)
	if err != nil {
		return false, err
	}
	if userID != club.HostID {
		return false, nil
	}
	return true, nil
}

func prepUpdate(c *fiber.Ctx, userID string) error {
	uid, err := users.GetUID(c)
	if err != nil {
		return err
	}
	check, err := checkIfHost(uid, userID)
	if err != nil {
		return errors.NotFoundError(c, "Club not found")
	}
	if !check {
		return errors.UnauthorizedError(c, "You are not the host of this group")
	}
	return nil
}

func prepToggle(c *fiber.Ctx) (*clubRequests, error) {
	r, err := parseClubIDRequest(c)
	if err != nil || r == nil {
		return nil, err
	}
	if err = prepUpdate(c, r.ID); err != nil {
		return nil, err
	}
	return r, nil
}

// Handlers

func createClub(c *fiber.Ctx) error {
	r := new(schemas.ClubCreate)
	if err := c.BodyParser(r); err != nil {
		return errors.UnprocessableEntityError(c, "form data in the incorrect format")
	}
	uid, err := users.GetUID(c)
	if err != nil {
		return err
	}
	r.HostID = uid
	created, err := crud.CreateClub(r)
	if err != nil {
		return errors.InternalServerError(c, "")
	}
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   created,
	})
}

func joinClub(c *fiber.Ctx) error {
	r, err := parseClubIDRequest(c)
	if err != nil {
		return err
	}
	clubID := r.ID
	club, err := crud.GetClub(clubID)
	if err != nil {
		return errors.NotFoundError(c, "Club not found")
	}
	usersList, err := crud.GetClubUser(clubID)
	if err != nil {
		return errors.InternalServerError(c, "")
	}
	user := c.Locals("user").(*models.User)
	uidBytes, err := user.ID.MarshalText()
	if err != nil {
		return errors.InternalServerError(c, "")
	}
	_, err = crud.CreateClubUser(clubID, string(uidBytes))
	if err != nil {
		pgErr := err.(pg.Error)
		if pgErr.IntegrityViolation() {
			return errors.ConflictError(c, "You are already a member of this club")
		}
		return errors.InternalServerError(c, "")
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"club":     club,
			"users":    usersList,
			"comments": "{}",
			"chat":     "{}",
		},
		"message": "Club joined successfully",
	})
}

func listClubs(c *fiber.Ctx) error {
	uid, err := users.GetUID(c)
	if err != nil {
		return err
	}
	clubs, err := crud.ListClub(uid)
	if err != nil {
		return errors.InternalServerError(c, "")
	}
	if clubs == nil {
		return errors.NotFoundError(c, "No public clubs found")
	}
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   clubs,
	})
}

func getClub(c *fiber.Ctx) error {
	clubID := c.Query("club_id")
	usersList, err := crud.GetClubUser(clubID)
	user := c.Locals("user").(*models.User)

	var isMember = false
	for _, clubUser := range usersList {
		if clubUser.ID == user.ID {
			isMember = true
			break
		}
	}
	if !isMember {
		return errors.UnauthorizedError(c, "You are not a member of this club")
	}
	if err != nil {
		return errors.InternalServerError(c, "")
	}
	club, err := crud.GetClub(clubID)
	if err != nil {
		return errors.InternalServerError(c, "")
	}
	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"club":     club,
			"users":    usersList,
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
	if err := prepUpdate(c, r.ID); err != nil {
		return err
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

func togglePrivate(c *fiber.Ctx) error {
	r, err := prepToggle(c)
	if err != nil || r == nil {
		return err
	}
	if err := crud.TogglePrivate(r.ID); err != nil {
		return errors.InternalServerError(c, "")
	}
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Club private feature has been toggled",
	})
}

func toggleSync(c *fiber.Ctx) error {
	r, err := prepToggle(c)
	if err != nil || r == nil {
		return err
	}
	if err := crud.ToggleSync(r.ID); err != nil {
		return errors.InternalServerError(c, "")
	}
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Club page sync feature has been toggled",
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
	clubID := r.ClubID
	deviantID := r.UserID
	club, err := crud.GetClub(clubID)
	if err != nil {
		return errors.NotFoundError(c, "Club not found")
	}
	uid, err := users.GetUID(c)
	if err != nil {
		return err
	}
	if uid != club.HostID {
		return errors.UnauthorizedError(c, "You are not the host of this group")
	} else if uid == deviantID {
		return errors.ConflictError(c, "You cannot kick yourself out of the club")
	}
	_, err = crud.DeleteClubUser(clubID, r.UserID)
	if err != nil {
		return errors.InternalServerError(c, "")
	}
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "User has been kicked from the club",
	})
}

func leaveClub(c *fiber.Ctx) error {
	r, err := parseClubIDRequest(c)
	if err != nil {
		return err
	}
	clubID := r.ID
	uid, err := users.GetUID(c)
	if err != nil {
		return errors.InternalServerError(c, "")
	}
	_, err = crud.DeleteClubUser(clubID, uid)
	if err != nil {
		if err == pg.ErrNoRows {
			return errors.ConflictError(c, "You are not a member of this club")
		}
		return errors.InternalServerError(c, "")
	}
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "You have left the club",
	})
}

func MountRoutes(app *fiber.App, middleware func(c *fiber.Ctx) error) {
	authGroup := app.Group("/api/clubs", middleware)
	authGroup.Get("", getClub)
	authGroup.Get("list", listClubs)
	authGroup.Post("create", createClub)
	authGroup.Post("join", joinClub)
	authGroup.Patch("update", updateClub)
	authGroup.Post("toggleprivate", togglePrivate)
	authGroup.Post("togglesync", toggleSync)
	authGroup.Post("kickuser", kickUser)
	authGroup.Post("leave", leaveClub)
}
