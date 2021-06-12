package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Krishap-s/keats-backend/crud"
	"github.com/Krishap-s/keats-backend/redisclient"
	"github.com/Krishap-s/keats-backend/schemas"
	"github.com/go-pg/pg/v10"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"log"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
)

// Client is a middleman between the websocket connection and the pubsub connection.
type Client struct {
	// User Id of client
	UserID string

	// Redis PubSub channel
	PubSub *redis.PubSub

	// Channel the client is subscribed to
	ClubID string

	// The websocket connection.
	conn *websocket.Conn

	// Buffered ClubID of outbound messages.
	send <-chan *redis.Message
}

// readPump pumps messages from the websocket connection to the pubsub channel.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump(rdb *redis.Client) {
	defer func() {
		_ = c.PubSub.Close()
		_ = c.conn.Close()
	}()
	for {
		var publishMessage *fiber.Map
		var jsonMessage map[string]interface{}
		err := c.conn.ReadJSON(&jsonMessage)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
				log.Printf("error: %v", err)
			}
			break
		}

		if jsonMessage["action"] == "" {
			_ = c.conn.WriteJSON(fiber.Map{
				"status":  "error",
				"message": "Invalid message format",
			})
			continue
		}
		switch jsonMessage["action"] {
		case "chatmessage":
			text, ok := jsonMessage["data"].(string)
			if !ok {
				_ = c.conn.WriteJSON(fiber.Map{
					"action":  "error",
					"message": "data in incorrect format",
				})
				continue
			}

			chatmessage := &schemas.ChatMessageCreate{
				UserID:  c.UserID,
				ClubID:  c.ClubID,
				Message: text,
				Likes:   0,
			}
			createdchatmessage, err := crud.CreateChatMessage(chatmessage)
			if err != nil {
				_ = c.conn.WriteJSON(fiber.Map{
					"action":  "error",
					"message": "Something went wrong",
				})
				continue
			}
			publishMessage = &fiber.Map{
				"user_id": c.UserID,
				"action":  "chatmessage",
				"data":    createdchatmessage,
			}
		case "like_chatmessage":
			id, ok := jsonMessage["data"].(string)
			_, err = uuid.Parse(id)
			if !ok || err != nil {
				_ = c.conn.WriteJSON(fiber.Map{
					"action":  "error",
					"message": "data in incorrect format",
				})
				continue
			}

			publishMessage = &fiber.Map{
				"user_id":        c.UserID,
				"action":         "like_chatmessage",
				"chatmessage_id": id,
			}
			err = crud.AddChatMessageLike(id)
			if err != nil {
				if err == pg.ErrNoRows {
					_ = c.conn.WriteJSON(fiber.Map{
						"action":  "error",
						"message": "Chatmessage not found",
					})
					continue
				}
				_ = c.conn.WriteJSON(fiber.Map{
					"action":  "error",
					"message": "Something went wrong",
				})
				continue
			}
		case "comment":
			commentJSON, err := json.Marshal(jsonMessage["data"])
			var comment schemas.CommentCreate
			err = json.Unmarshal(commentJSON, &comment)
			if err != nil || comment.Message == "" || comment.PageNo == 0 {
				_ = c.conn.WriteJSON(fiber.Map{
					"action":  "error",
					"message": "data in incorrect format",
				})
				continue
			}
			comment.UserID = c.UserID
			comment.ClubID = c.ClubID
			createdcomment, err := crud.CreateComment(&comment)
			if err != nil {
				_ = c.conn.WriteJSON(fiber.Map{
					"action":  "error",
					"message": err.Error(),
				})
				continue
			}
			publishMessage = &fiber.Map{
				"user_id": c.UserID,
				"action":  "comment",
				"data":    createdcomment,
			}
		case "like_comment":
			id, ok := jsonMessage["data"].(string)
			_, err = uuid.Parse(id)
			if !ok || err != nil {
				_ = c.conn.WriteJSON(fiber.Map{
					"action":  "error",
					"message": "data in incorrect format",
				})
				continue
			}

			publishMessage = &fiber.Map{
				"user_id":    c.UserID,
				"action":     "like_comment",
				"comment_id": id,
			}
			err = crud.AddCommentLike(id)
			if err != nil {
				if err == pg.ErrNoRows {
					_ = c.conn.WriteJSON(fiber.Map{
						"action":  "error",
						"message": "Comment not found",
					})
					continue
				}
				_ = c.conn.WriteJSON(fiber.Map{
					"action":  "error",
					"message": "Something went wrong",
				})
				continue
			}
		default:
			_ = c.conn.WriteJSON(fiber.Map{
				"status":  "error",
				"message": fmt.Sprint(jsonMessage["action"], " is not an action"),
			})
		}
		bytePublishMessage, err := json.Marshal(publishMessage)
		// Publish to websocket ClubID
		ctx := context.Background()
		rdb.Publish(ctx, c.ClubID, bytePublishMessage)
	}
}

// writePump pumps messages from the pubsub channel to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case message, ok := <-c.send:
			var jsonMessage map[string]interface{}
			if !ok {
				// The pubsub closed the ClubID.
				return
			}
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			byteMessage := []byte(message.Payload)
			_ = json.Unmarshal(byteMessage, &jsonMessage)
			_, _ = w.Write([]byte(message.Payload))

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, _ = w.Write(newline)
				_, _ = w.Write([]byte((<-c.send).String()))
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWs handles websocket requests from the peer.
func ServeWs(conn *websocket.Conn, userID string, clubID string) {

	ctx := context.Background()
	rdb, err := redisclient.GetRedisClient()
	if err != nil {
		log.Println(err)
		return
	}
	pubsub := rdb.Subscribe(ctx, clubID)
	c := pubsub.Channel()
	client := &Client{UserID: userID, ClubID: clubID, PubSub: pubsub, conn: conn, send: c}
	client.conn.SetReadLimit(maxMessageSize)
	_ = client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error { _ = client.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	client.readPump(rdb)
}
