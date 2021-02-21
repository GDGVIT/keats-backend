package models

import "github.com/google/uuid"

// Room represents a room in the database
type Room struct {
	ID       uuid.UUID `pg:",pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	PageSync string    `pg:",notnull,unique" json:"page_sync"`
	FileURL  bool      `pg:",notnull" json:"file_url"`
	PageNo   int       `json:"page_no"`
}
