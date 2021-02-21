package models

import "github.com/google/uuid"

type RoomUser struct {
	ID     uuid.UUID `pg:"pk,type:uuid,default:generate_uuid_v4()" json:"id"`
	RoomID uuid.UUID `pg:"type:uuid,nopk,notnull" json:"room_id"`
	UserID uuid.UUID `pg:"type:uuid,nopk,notnull" json:"user_id"`
}
