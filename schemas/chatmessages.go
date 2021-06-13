package schemas

import "time"

// ChatMessageCreate represents a chat message to be created
type ChatMessageCreate struct {
	ClubID  string `json:"club_id"`
	UserID  string `json:"user_id"`
	Message string `json:"message"`
	Likes   int    `json:"likes"`
}

// ChatMessage represents a chat message to be returned as a response
type ChatMessage struct {
	ID      string `json:"id"`
	ClubID  string `json:"club_id"`
	UserID  string `json:"user_id"`
	Message string `json:"message"`
	Likes   int    `json:"likes"`
	TimeCreated time.Time `json:"time_created"`
}
