package models

import "github.com/google/uuid"

// Comment represents a commnent in the database
type Comment struct {
	ID       uuid.UUID `pg:",pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	ClubID   uuid.UUID `pg:"type:uuid,notnull,nopk" json:"club_id"`
	ParentID uuid.UUID `pg:"type:uuid,nopk" json:"parent_id"`
	UserID   uuid.UUID `pg:"type:uuid,notnull,nopk" json:"user_id"`
	PageNo   int       `pg:",notnull" json:"page_no"`
	Message  string    `pg:",notnull"`
	Likes    int       `pg:",notnull,default:0" json:"likes"`
}
