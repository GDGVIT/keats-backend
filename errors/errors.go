package errors

import (
	"github.com/gofiber/fiber/v2"
	"log"
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
	log.Println("ip", c.IP(), "user:", c.Locals("user"), err)
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
