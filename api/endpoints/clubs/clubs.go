package clubs

import (
	"fmt"
	"mime/multipart"
	"strconv"
	"time"

	"github.com/Krishap-s/keats-backend/api/endpoints/users"
	"github.com/Krishap-s/keats-backend/crud"
	"github.com/Krishap-s/keats-backend/firebaseclient"
	"github.com/Krishap-s/keats-backend/models"
	"github.com/Krishap-s/keats-backend/redisclient"
	"github.com/Krishap-s/keats-backend/schemas"
	"github.com/Krishap-s/keats-backend/utils"
	"github.com/go-pg/pg/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

// Non Handlers

type clubRequests struct {
	ID string `json:"club_id"`
}

func parseClubIDRequest(c *fiber.Ctx) (*clubRequests, error) {
	r := new(clubRequests)
	if err := c.BodyParser(r); err != nil || r.ID == "" {
		return nil, fmt.Errorf("JSON Data Incorrect")
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
		return fmt.Errorf("club not found")
	}
	if !check {
		return fmt.Errorf("not host")
	}
	return nil
}

func updateClubFiles(c *fiber.Ctx) (string, string, error) {
	var clubPicURL, fileURL string
	//nolint
	clubPicFileHeader, _ := c.FormFile("club_pic")
	if clubPicFileHeader != nil {
		clubPicFile, err := clubPicFileHeader.Open()
		if err != nil {
			return "", "", fmt.Errorf("file parse error")
		}
		defer utils.CloseFile(clubPicFile)
		acceptedTypes := []string{"image/png", "image/jpeg"}
		clubPicURL, err = firebaseclient.WriteObject(&clubPicFile, acceptedTypes)
		if err != nil {
			return "", "", err
		}
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return "", "", fmt.Errorf("form Data Incorrect")
	}
	if fileHeader != nil {
		var fileFile multipart.File
		fileFile, err = fileHeader.Open()
		if err != nil {
			return "", "", fmt.Errorf("file parse error")
		}
		defer utils.CloseFile(fileFile)
		acceptedTypes := []string{"application/pdf", "application/epub+xml", "application/epub+zip", "application/zip"}
		fileURL, err = firebaseclient.WriteObject(&fileFile, acceptedTypes)
		if err != nil {
			return "", "", err
		}
	}
	return clubPicURL, fileURL, nil
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
	maxClubCreated := viper.GetInt64("MAX_NUMBER_OF_CLUBS_CREATED")
	timePeriod := viper.GetFloat64("TIME_PERIOD_CLUB_CREATED_LIMIT")
	timePeriodHour := time.Hour * time.Duration(timePeriod)
	rdb, err := redisclient.GetRedisClient()
	if err != nil {
		return err
	}
	pipe := rdb.TxPipeline()
	r := new(schemas.ClubCreate)
	if err = c.BodyParser(r); err != nil {
		return fmt.Errorf("form Data Incorrect")
	}
	var uid string
	uid, err = users.GetUID(c)
	if err != nil {
		return err
	}
	counterKey := uid + "_create_club"
	count := pipe.Incr(c.Context(), counterKey)
	pipe.Expire(c.Context(), counterKey, timePeriodHour)
	_, err = pipe.Exec(c.Context())
	if err != nil {
		return err
	}
	if count.Val() >= maxClubCreated {
		return fmt.Errorf("max clubs created")
	}
	r.HostID = uid
	r.ClubPic, r.FileURL, err = updateClubFiles(c)
	if err != nil {
		rdb.Decr(c.Context(), counterKey)
		return err
	}
	created, err := crud.CreateClub(r)
	if err != nil {
		rdb.Decr(c.Context(), counterKey)
		return err
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
		return fmt.Errorf("club not found")
	}
	usersList, err := crud.GetClubUser(clubID)
	if err != nil {
		return err
	}
	// Shows host user as the first user
	for i, clubUser := range usersList {
		if clubUser.ID.String() == club.HostID {
			usersList[0], usersList[i] = usersList[i], usersList[0]
			break
		}
	}
	user := c.Locals("user").(*models.User)
	uidBytes, err := user.ID.MarshalText()
	if err != nil {
		return err
	}
	_, err = crud.CreateClubUser(clubID, string(uidBytes))
	if err != nil {
		pgErr := err.(pg.Error)
		if pgErr.IntegrityViolation() {
			return fmt.Errorf("already member")
		}
		return err
	}
	chatMessages, err := crud.GetChatMessage(clubID)
	if err != nil {
		return err
	}

	comments, err := crud.GetComment(clubID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"club":     club,
			"users":    usersList,
			"comments": comments,
			"chat":     chatMessages,
		},
		"message": "Club joined successfully",
	})
}

func listClubs(c *fiber.Ctx) error {
	var n int
	uid, err := users.GetUID(c)
	if err != nil {
		return err
	}
	n, err = strconv.Atoi(c.Query("page", "0"))
	if err != nil || n < 1 {
		n = 1
	}
	clubs, err := crud.ListClub(uid, n)
	if err != nil {
		return err
	}
	if clubs == nil {
		return fmt.Errorf("no public")
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
		return fmt.Errorf("not member")
	}
	if err != nil {
		return err
	}
	club, err := crud.GetClub(clubID)
	if err != nil {
		return err
	}
	// Shows host user as the first user
	for i, clubUser := range usersList {
		if clubUser.ID.String() == club.HostID {
			usersList[0], usersList[i] = usersList[i], usersList[0]
			break
		}
	}
	chatMessages, err := crud.GetChatMessage(clubID)
	if err != nil {
		return err
	}

	comments, err := crud.GetComment(clubID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"club":     club,
			"users":    usersList,
			"comments": comments,
			"chat":     chatMessages,
		},
	})
}

func updateClub(c *fiber.Ctx) error {
	r := new(schemas.ClubUpdate)
	if err := c.BodyParser(r); err != nil || r.ID == "" {
		return fmt.Errorf("JSON Data Incorrect")
	}
	if err := prepUpdate(c, r.ID); err != nil {
		return err
	}
	var err error
	r.ClubPic, r.FileURL, err = updateClubFiles(c)
	if err != nil {
		return err
	}
	updated, err := crud.UpdateClub(r)
	if err != nil {
		return err
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
		return err
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
		return err
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
		return fmt.Errorf("JSON Data Incorrect")
	}
	clubID := r.ClubID
	deviantID := r.UserID
	club, err := crud.GetClub(clubID)
	if err != nil {
		return fmt.Errorf("already member")
	}
	uid, err := users.GetUID(c)
	if err != nil {
		return err
	}
	if uid != club.HostID {
		return fmt.Errorf("not host")
	} else if uid == deviantID {
		return fmt.Errorf("self kick")
	}
	_, err = crud.DeleteClubUser(clubID, r.UserID)
	if err != nil {
		return err
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
		return err
	}
	_, err = crud.DeleteClubUser(clubID, uid)
	if err != nil {
		if err == pg.ErrNoRows {
			return fmt.Errorf("not member")
		}
		return err
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
