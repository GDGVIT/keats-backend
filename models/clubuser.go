package models

import (
	"context"
	"encoding/json"
	"github.com/Krishap-s/keats-backend/redisclient"
	"github.com/go-pg/pg/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ClubUser struct {
	ID     uuid.UUID `pg:",pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	ClubID uuid.UUID `pg:"type:uuid,nopk,notnull,unique:clubuser" json:"room_id"`
	UserID uuid.UUID `pg:"type:uuid,nopk,notnull,unique:clubuser" json:"user_id"`
}

var _ pg.AfterInsertHook = (*ClubUser)(nil)

// AfterInsert hook publishes to websocket clients that a user has joined the club
func (c *ClubUser) AfterInsert(ctx context.Context) error {
	rdb, err := redisclient.GetRedisClient()
	if err != nil {
		return err
	}
	userID := c.UserID.String()
	clubID := c.ClubID.String()
	byteData, _ := json.Marshal(fiber.Map{
		"action": "User Join",
		"data":   userID,
	})
	rdb.Publish(ctx, clubID, byteData)
	return nil
}

var _ pg.AfterDeleteHook = (*ClubUser)(nil)

// AfterDelete hook publishes to websocket clients that a user has left the club
func (c *ClubUser) AfterDelete(ctx context.Context) error {
	rdb, err := redisclient.GetRedisClient()
	if err != nil {
		return err
	}
	userID := c.UserID.String()
	clubID := c.ClubID.String()
	byteData, _ := json.Marshal(fiber.Map{
		"action": "User Leave",
		"data":   userID,
	})
	rdb.Publish(ctx, clubID, byteData)
	return nil
}
