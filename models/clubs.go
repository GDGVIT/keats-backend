package models

import (
	"context"
	"encoding/json"
	"github.com/Krishap-s/keats-backend/redisclient"
	"github.com/go-pg/pg/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Room represents a room in the database
type Club struct {
	ClubName string    `pg:",notnull" json:"clubname"`
	ClubPic  string    `json:"club_pic"`
	FileURL  string    `pg:",notnull" json:"file_url"`
	PageNo   int       `json:"page_no"`
	ID       uuid.UUID `pg:",pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	HostID   uuid.UUID `pg:",type:uuid" json:"host_id"`
	PageSync bool      `pg:",use_zero" json:"page_sync"`
	Private  bool      `pg:",use_zero" json:"private"`
}

var _ pg.AfterUpdateHook = (*Club)(nil)

// AfterUpdate hook publishes club update notifications to websocket users
func (c *Club) AfterUpdate(ctx context.Context) error {
	rdb, err := redisclient.GetRedisClient()
	if err != nil {
		return err
	}
	clubID := c.ID.String()
	byteData, _ := json.Marshal(fiber.Map{
		"action": "club_update",
		"data":   c,
	})
	rdb.Publish(ctx, clubID, byteData)
	return nil
}
