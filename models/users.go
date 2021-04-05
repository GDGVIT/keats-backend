package models

import "github.com/google/uuid"

// User represents a user in the database
type User struct {
	ID         uuid.UUID `pg:",pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	Username   string    `pg:",notnull" json:"username"`
	PhoneNo    string    `pg:",unique,notnull" json:"phone_number"`
	ProfilePic string    `pg:",default:\'https://i.ibb.co/drJX0MS/default-photo.jpg\'" json:"profile_pic"`
	Email      string    `json:"email"`
	Bio        string    `json:"bio"`
}
