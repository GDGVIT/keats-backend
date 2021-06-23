package models

import "github.com/google/uuid"

type ClubUser struct {
	ID     uuid.UUID `pg:",pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	ClubID uuid.UUID `pg:"type:uuid,nopk,notnull,unique:clubuser" json:"room_id"`
	UserID uuid.UUID `pg:"type:uuid,nopk,notnull,unique:clubuser" json:"user_id"`
}
