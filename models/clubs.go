package models

import "github.com/google/uuid"

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
