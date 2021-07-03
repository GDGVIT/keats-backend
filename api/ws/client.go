package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/Krishap-s/keats-backend/crud"
	"github.com/Krishap-s/keats-backend/models"
	"github.com/Krishap-s/keats-backend/redisclient"
	"github.com/Krishap-s/keats-backend/schemas"
	"github.com/go-pg/pg/v10"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 20 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 10 * time.Second

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

	// Kill switch channel to synchronise closing of both readPump and writePump
	killChannel chan bool
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
		c.killChannel <- true
	}()
	for {
		var publishMessage *fiber.Map
		var jsonMessage map[string]interface{}
		err := c.conn.ReadJSON(&jsonMessage)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
				log.Printf("error: %v", err)
				break
			}
			err = c.conn.WriteJSON(fiber.Map{
				"status":  "error",
				"message": "Invalid message format",
			})
			log.Println("Websocket error:", err)
			continue
		}

		if jsonMessage["action"] == "" {
			err = c.conn.WriteJSON(fiber.Map{
				"status":  "error",
				"message": "Invalid message format",
			})
			log.Println("Websocket error:", err)
			continue
		}
		switch jsonMessage["action"] {
		case "chatmessage":
			text, ok := jsonMessage["data"].(string)
			if !ok {
				err = c.conn.WriteJSON(fiber.Map{
					"action":  "error",
					"message": "data in incorrect format",
				})
				log.Println("Websocket error:", err)
				continue
			}

			chatmessage := &schemas.ChatMessageCreate{
				UserID:  c.UserID,
				ClubID:  c.ClubID,
				Message: text,
				Likes:   0,
			}
			var createdchatmessage *models.ChatMessage
			createdchatmessage, err = crud.CreateChatMessage(chatmessage)
			if err != nil {
				err = c.conn.WriteJSON(fiber.Map{
					"action":  "error",
					"message": "Something went wrong",
				})
				log.Println("Websocket error:", err)
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
				err = c.conn.WriteJSON(fiber.Map{
					"action":  "error",
					"message": "data in incorrect format",
				})
				log.Println("Websocket error:", err)
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
					err = c.conn.WriteJSON(fiber.Map{
						"action":  "error",
						"message": "Chatmessage not found",
					})
					log.Println("Websocket error:", err)
					continue
				}
				err = c.conn.WriteJSON(fiber.Map{
					"action":  "error",
					"message": "Something went wrong",
				})
				log.Println("Websocket error:", err)
				continue
			}
		case "comment":
			var commentJSON []byte
			commentJSON, err = json.Marshal(jsonMessage["data"])
			log.Println("Websocket error:", err)
			var comment schemas.CommentCreate
			err = json.Unmarshal(commentJSON, &comment)
			if err != nil || comment.Message == "" || comment.PageNo == 0 {
				err = c.conn.WriteJSON(fiber.Map{
					"action":  "error",
					"message": "data in incorrect format",
				})
				log.Println("Websocket error:", err)
				continue
			}
			comment.UserID = c.UserID
			comment.ClubID = c.ClubID
			var createdcomment *models.Comment
			createdcomment, err = crud.CreateComment(&comment)
			if err != nil {
				err = c.conn.WriteJSON(fiber.Map{
					"action":  "error",
					"message": err.Error(),
				})
				log.Println("Websocket error:", err)
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
				err = c.conn.WriteJSON(fiber.Map{
					"action":  "error",
					"message": "data in incorrect format",
				})
				log.Println("Websocket error:", err)
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
					err = c.conn.WriteJSON(fiber.Map{
						"action":  "error",
						"message": "Comment not found",
					})
					log.Println("Websocket error:", err)
					continue
				}
				err = c.conn.WriteJSON(fiber.Map{
					"action":  "error",
					"message": "Something went wrong",
				})
				log.Println("Websocket error:", err)
				continue
			}
		default:
			err = c.conn.WriteJSON(fiber.Map{
				"status":  "error",
				"message": fmt.Sprint(jsonMessage["action"], " is not an action"),
			})
			log.Println("Websocket error:", err)
		}
		var bytePublishMessage []byte
		bytePublishMessage, err = json.Marshal(publishMessage)
		log.Println("Websocket error:", err)
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
				break
			}
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			log.Println("Websocket error:", err)

			var w io.WriteCloser
			w, err = c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				break
			}
			byteMessage := []byte(message.Payload)
			err = json.Unmarshal(byteMessage, &jsonMessage)
			log.Println("Websocket error:", err)
			_, err = w.Write([]byte(message.Payload))
			log.Println("Websocket error:", err)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, err = w.Write(newline)
				log.Println("Websocket error:", err)
				_, err = w.Write([]byte((<-c.send).String()))
				log.Println("Websocket error:", err)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			log.Println("Websocket error:", err)
			if err = c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.killChannel:
			break
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
	err = client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error {
		err = client.conn.SetReadDeadline(time.Now().Add(pongWait))
		log.Println("Websockets error:", err)
		return nil
	})
	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	client.readPump(rdb)
}
