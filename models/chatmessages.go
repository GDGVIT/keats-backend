package models

import (
	"github.com/google/uuid"
	"time"
)

// ChatMessage represents a chatmessage in the database
type ChatMessage struct {
	ID      uuid.UUID `pg:",pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	ClubID  uuid.UUID `pg:"type:uuid,notnull,nopk" json:"club_id"`
	UserID  uuid.UUID `pg:"type:uuid,notnull,nopk" json:"user_id"`
	Message string    `pg:",notnull" json:"message"`
	Likes   int       `pg:",notnull,default:0" json:"likes"`
	TimeCreated time.Time `pg:",notnull,default:now()" json:"time_created"`
}
