package crud

import (
	"fmt"
	"time"

	"github.com/Krishap-s/keats-backend/models"
	"github.com/Krishap-s/keats-backend/pgdb"
	"github.com/Krishap-s/keats-backend/schemas"
	"github.com/google/uuid"
)

// CreateChatMessage creates a chatmessage in the database or returns an error
func CreateChatMessage(objIn *schemas.ChatMessageCreate) (*models.ChatMessage, error) {
	db := pgdb.GetDB()
	cid, err := uuid.Parse(objIn.ClubID)
	if err != nil {
		return nil, err
	}
	uid, err := uuid.Parse(objIn.UserID)
	if err != nil {
		return nil, err
	}
	if len(objIn.Message) > 150 {
		return nil, fmt.Errorf("max string length")
	}
	chatmessage := &models.ChatMessage{
		Message:     objIn.Message,
		ClubID:      cid,
		UserID:      uid,
		TimeCreated: time.Now(),
	}
	_, err = db.Model(chatmessage).Returning("*").Insert()
	if err != nil {
		return nil, err
	}

	return chatmessage, nil
}

// GetChatMessage gets chatmessages from a room or returns an error
func GetChatMessage(cid string) ([]*schemas.ChatMessage, error) {
	db := pgdb.GetDB()
	var chatmessages []*schemas.ChatMessage
	err := db.Model((*models.ChatMessage)(nil)).
		Where("club_id = ?", cid).
		Order("time_created ASC").
		Select(&chatmessages)
	if err != nil {
		return nil, err
	}
	return chatmessages, nil
}

// AddChatMessageLike increments likes field of chatmessage
func AddChatMessageLike(id string) error {
	db := pgdb.GetDB()
	chatmessage := models.ChatMessage{}
	_, err := db.Model(&chatmessage).
		Set("likes = likes + 1").
		Where("id = ?", id).
		Returning("id").
		UpdateNotZero()
	if err != nil {
		return err
	}
	return nil
}
