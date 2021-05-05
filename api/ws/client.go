package ws

import (
	"bytes"
	"context"
	"github.com/Krishap-s/keats-backend/redisclient"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/websocket/v2"
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
	space   = []byte{' '}
)

// Client is a middleman between the websocket connection and the pubsub connection.
type Client struct {
	// Redis PubSub channel
	PubSub *redis.PubSub

	// Channel the client is subscribed to
	channel string

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send <-chan *redis.Message
}

// readPump pumps messages from the websocket connection to the pubsub channel.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump(rdb *redis.Client) {
	defer func() {
		c.PubSub.Close()
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		ctx := context.Background()
		// Publish to websocket channel
		rdb.Publish(ctx, c.channel, message)
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
			if !ok {
				// The pubsub closed the channel.
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write([]byte(message.Payload))

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write([]byte((<-c.send).String()))
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWs handles websocket requests from the peer.
func ServeWs(conn *websocket.Conn, clubID string) {

	ctx := context.Background()
	rdb, err := redisclient.GetRedisClient()
	if err != nil {
		log.Println(err)
		return
	}
	pubsub := rdb.Subscribe(ctx, clubID)

	// Wait for confirmation that subscription is created before publishing anything.
	_, err = pubsub.Receive(ctx)
	if err != nil {
		panic(err)
	}
	c := pubsub.Channel()
	client := &Client{channel: clubID, PubSub: pubsub, conn: conn, send: c}
	client.conn.SetReadLimit(maxMessageSize)
	_ = client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error { _ = client.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	client.readPump(rdb)
}
