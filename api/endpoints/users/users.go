package users

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/Krishap-s/keats-backend/configs"
	"github.com/Krishap-s/keats-backend/crud"
	"github.com/Krishap-s/keats-backend/firebaseclient"
	"github.com/Krishap-s/keats-backend/models"
	"github.com/Krishap-s/keats-backend/schemas"
	jwt "github.com/form3tech-oss/jwt-go"
	"github.com/go-pg/pg/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

// Non Handlers

func GetUID(c *fiber.Ctx) (string, error) {
	user := c.Locals("user").(*models.User)
	uidBytes, err := user.ID.MarshalText()
	if err != nil {
		return "", err
	}
	uid := string(uidBytes)
	return uid, nil
}

type IDTokenRequest struct {
	IDToken string `json:"id_token"`
}

func getPhoneNo(c *fiber.Ctx) (string, error) {
	req := new(IDTokenRequest)
	client, err := firebaseclient.GetClient()
	if err != nil {
		return "", err
	}
	err = c.BodyParser(req)
	if err != nil {
		return "", fmt.Errorf("malformed IDToken")
	}
	fireToken, err := client.VerifyIDToken(context.Background(), req.IDToken)
	if err != nil {
		return "", fmt.Errorf("IDToken verification failed")
	}
	phoneNumber, ok := fireToken.Claims["phone_number"].(string)
	if !ok {
		return "", fmt.Errorf("no phoneNo")
	}
	return phoneNumber, nil
}

func createJWT(user *models.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = user.ID
	signedToken, err := token.SignedString([]byte(configs.GetSecret()))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

// Handlers

func createUser(c *fiber.Ctx) error {
	phoneNumber, err := getPhoneNo(c)
	if err != nil {
		return err
	}
	u := &schemas.UserCreate{
		PhoneNo: phoneNumber,
	}

	created, err := crud.CreateUser(u)
	if err != nil {
		return err
	}

	signedToken, err := createJWT(created)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"token":   signedToken,
			"user_id": created.ID,
		},
	})
}

func updateUser(c *fiber.Ctx) error {
	u := new(schemas.UserUpdate)
	if err := c.BodyParser(u); err != nil {
		return fmt.Errorf("JSON Data Incorrect")
	}
	uid, err := GetUID(c)
	if err != nil {
		return err
	}
	u.ID = uid
	updated, err := crud.UpdateUser(u)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   updated,
	})
}

func updateUserProfilePic(c *fiber.Ctx) error {
	u := new(schemas.UserUpdate)
	r := new(struct {
		ProfilePic string `json:"profile_pic"`
	})
	if err := c.BodyParser(r); err != nil {
		return fmt.Errorf("JSON Data Incorrect")
	}

	user := c.Locals("user").(*models.User)
	uidBytes, err := user.ID.MarshalText()
	if err != nil {
		return err
	}
	u.ID = string(uidBytes)
	u.ProfilePic = r.ProfilePic
	updated, err := crud.UpdateUser(u)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   updated,
	})

}

func updateUserPhoneNo(c *fiber.Ctx) error {
	phoneNumber, err := getPhoneNo(c)
	if err != nil {
		return err
	}
	u := &schemas.UserUpdate{
		PhoneNo: phoneNumber,
	}
	u.ID, err = GetUID(c)
	if err != nil {
		return err
	}
	_, err = crud.UpdateUser(u)
	if err != nil {
		pgErr := err.(pg.Error)
		if pgErr.IntegrityViolation() {
			return fmt.Errorf("no phoneNo")
		}
		return err
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
	uid, err := GetUID(c)
	if err != nil {
		return err
	}
	clubs, err := crud.GetUserClub(uid)
	if err != nil {
		if err == pg.ErrNoRows {
			return fmt.Errorf("club not found")
		}
		return err
	}
	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"clubs": clubs,
			"user":  c.Locals("user"),
		},
	})

}

func closeFile(file multipart.File) {
	err := file.Close()
	if err != nil {
		log.Println("error:", err.Error())
	}
}

func uploadFile(c *fiber.Ctx) error {
	if form, err := c.MultipartForm(); err == nil {
		bucketName := viper.GetString("FIREBASE_BUCKET_NAME")
		// Open form file header
		fileHeader := form.File["file"][0]
		file, err := fileHeader.Open()
		if err != nil {
			return fmt.Errorf("file parse error")
		}
		// Close file when function ends
		defer closeFile(file)
		fileData := make([]byte, 512)
		_, err = io.ReadAtLeast(file, fileData, 512)
		if err != nil {
			return fmt.Errorf("file parse error")
		}
		// Resets file pointer
		_, err = file.Seek(0, 0)
		if err != nil {
			return err
		}
		contentType := http.DetectContentType(fileData)
		if !(contentType == "application/pdf" || contentType == "application/epub+xml" || contentType == "image/png" || contentType == "image/jpeg") {
			return fmt.Errorf("invalid file type")
		}
		bucketClient, err := firebaseclient.GetBucket()
		if err != nil {
			log.Println(err.Error())
			return err
		}
		bucket, err := bucketClient.Bucket(bucketName)
		if err != nil {
			return err
		}
		fid := uuid.NewString()
		filePath := "public/" + fid
		wc := bucket.Object(filePath).NewWriter(context.Background())
		if _, err = io.Copy(wc, file); err != nil {
			return err
		}
		if err = wc.Close(); err != nil {
			return err
		}

		fileURL := "https://firebasestorage.googleapis.com/v0/b/" + bucketName + "/o/public%2f" + fid + "?alt=media"
		return c.JSON(fiber.Map{
			"status": "success",
			"data":   fileURL,
		})
	}

	return fmt.Errorf("file parse error")

}

// MountRoutes mounts all routes declared here
func MountRoutes(app *fiber.App, middleware func(c *fiber.Ctx) error) {
	app.Post("/api/user", createUser)
	authGroup := app.Group("/api/user", middleware)
	authGroup.Patch("", updateUser)
	authGroup.Post("updateprofilepic", updateUserProfilePic)
	authGroup.Post("updatephone", updateUserPhoneNo)
	authGroup.Get("", getUser)
	authGroup.Get("clubs", getUserClubsAndDetails)
	app.Post("/api/uploadfile", uploadFile, middleware)
}
