package models

import "github.com/google/uuid"

// ChatMessage represents a chatmessage in the database
type ChatMessage struct {
	ID      uuid.UUID `pg:",pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	RoomID  uuid.UUID `pg:"type:uuid,notnull,nopk" json:"room_id"`
	UserID  uuid.UUID `pg:"type:uuid,notnull,nopk" json:"user_id"`
	Message string    `pg:",notnull"`
	Likes   int       `pg:",notnull,default:0" json:"likes"`
}
