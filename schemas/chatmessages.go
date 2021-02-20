package schemas

// ChatMessageCreate represents a chat message to be created
type ChatMessageCreate struct {
	ID      string `json:"id"`
	RoomID  string `json:"room_id"`
	UserID  string `json:"user_id"`
	Message string `json:"message"`
	Likes   int    `json:"likes"`
}

// ChatMessage represents a chat message to be returned as a response
type ChatMessage struct {
	ID      string `json:"id"`
	RoomID  string `json:"room_id"`
	UserID  string `json:"user_id"`
	Message string `json:"message"`
	Likes   int    `json:"likes"`
}
