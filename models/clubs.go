package models

import "github.com/google/uuid"

// Room represents a room in the database
type Club struct {
	ID       uuid.UUID `pg:",pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	ClubName string    `pg:",notnull" json:"clubname"`
	ClubPic  string    `json:"club_pic"`
	PageSync bool      `pg:",notnull" json:"page_sync"`
	FileURL  string    `pg:",notnull" json:"file_url"`
	PageNo   int       `json:"page_no"`
	Private  bool      `pg:",notnull" json:"private"`
	HostID   uuid.UUID `pg:",type:uuid" json:"host_id"`
}
