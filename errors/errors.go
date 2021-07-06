package errors

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func TooManyRequestsError(c *fiber.Ctx) error {
	return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
		"status":  "error",
		"message": "Too Many Requests",
	})
}

func BadRequestError(c *fiber.Ctx, err string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"status":  "error",
		"message": err,
	})
}

func UnauthorizedError(c *fiber.Ctx, err string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"status":  "error",
		"message": err,
	})
}

func InternalServerError(c *fiber.Ctx, err string) error {
	if err == "" {
		err = "Something went wrong"
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"status":  "error",
		"message": err,
	})
}
func UnprocessableEntityError(c *fiber.Ctx, err string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
		"status":  "error",
		"message": err,
	})
}

func ConflictError(c *fiber.Ctx, err string) error {
	return c.Status(fiber.StatusConflict).JSON(fiber.Map{
		"status":  "error",
		"message": err,
	})
}

func NotFoundError(c *fiber.Ctx, err string) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"status":  "error",
		"message": err,
	})
}

func ConstraintError(c *fiber.Ctx, err string) error {
	return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
		"status":  "error",
		"message": err,
	})
}

func MaxCreated(c *fiber.Ctx, err string) error {
	return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
		"status":  "error",
		"message": err,
	})
}
func ErrorHandler(c *fiber.Ctx, err error) error {
	switch err.Error() {
	case "form Data Incorrect":
		return UnprocessableEntityError(c, "Form Data In Incorrect Format")
	case "JSON Data Incorrect":
		return UnprocessableEntityError(c, "JSON Data In Incorrect Format")
	case "not member":
		return UnauthorizedError(c, "You are not a member of this club")
	case "already member":
		return ConflictError(c, "You are already a member of this club")
	case "club not found":
		return NotFoundError(c, "Club not found")
	case "not host":
		return UnauthorizedError(c, "You are not the host of this club")
	case "self kick":
		return ConflictError(c, "You cannot kick yourself out of the club")
	case "no public":
		return NotFoundError(c, "No public clubs found")
	case "malformed IDToken":
		return UnprocessableEntityError(c, "Missing or Malformed IDToken")
	case "no phoneNo":
		return BadRequestError(c, "IDToken missing phone_number")
	case "IDToken verification failed":
		return UnauthorizedError(c, "IDToken verification failed or IDToken expired")
	case "phoneNo exists":
		return ConflictError(c, "Phone Number already exists")
	case "file parse error":
		return BadRequestError(c, "Error finding or parsing file")
	case "invalid file type":
		return BadRequestError(c, "Invalid file type")
	case "malformed jwt":
		return BadRequestError(c, "Missing or malformed JWT")
	case "invalid jwt":
		return UnauthorizedError(c, "Invalid or Expired JWT")
	case "max string length":
		return ConstraintError(c, "One of your string inputs are too large")
	case "max clubs created":
		return MaxCreated(c, "You have exceeded maximum number of clubs created per user")
	}
	log.Println("Uncaught Error:", err.Error())
	return InternalServerError(c, "")
}
