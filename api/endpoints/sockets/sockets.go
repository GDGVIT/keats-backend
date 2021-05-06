package sockets

import (
	"fmt"
	"github.com/Krishap-s/keats-backend/api/ws"
	"github.com/Krishap-s/keats-backend/configs"
	"github.com/Krishap-s/keats-backend/crud"
	"github.com/form3tech-oss/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

func MountWebsockets(app *fiber.App, middleware func(c *fiber.Ctx) error) {
	wsRoutes := app.Group("/api/ws", middleware)
	wsRoutes.Use("", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	wsRoutes.Get(":id", websocket.New(func(conn *websocket.Conn) {
		clubID := conn.Params("id")
		_, err := crud.GetClub(clubID)
		if err != nil {
			_ = conn.WriteJSON(fiber.Map{
				"action":  "error",
				"message": "Club not found",
			})
			return
		}
		usersList, err := crud.GetClubUser(clubID)
		tokenstring := conn.Query("token")
		token, err := jwt.Parse(tokenstring, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return []byte(configs.GetSecret()), nil
		})
		if err != nil {
			if err.Error() == "Missing or malformed JWT" {
				_ = conn.WriteJSON(fiber.Map{
					"action":  "error",
					"message": "Missing or malformed JWT",
				})
				return
			}
			_ = conn.WriteJSON(fiber.Map{
				"action":  "error",
				"message": "Invalid or Expired JWT",
			})
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		uid := claims["id"].(string)
		userID, err := uuid.Parse(uid)
		if err != nil {
			_ = conn.WriteJSON(fiber.Map{
				"action":  "error",
				"message": "Invalid or Expired JWT",
			})
			return
		}
		var isMember = false
		for _, clubUser := range usersList {
			if clubUser.ID == userID {
				isMember = true
				break
			}
		}
		if !isMember {
			_ = conn.WriteJSON(fiber.Map{
				"action":  "error",
				"message": "You are not a member of this club",
			})
			return
		}
		ws.ServeWs(conn, uid, clubID)
	}))
}
