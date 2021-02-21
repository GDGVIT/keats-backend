package schemas

// CommentCreate represents a comment to be created
type CommentCreate struct {
	ID       string `json:"id"`
	RoomID   string `json:"room_id"`
	ParentID string `json:"parent_id"`
	UserID   string `json:"user_id"`
	PageNo   int    `json:"page_no"`
	Message  string `json:"message"`
	Likes    int    `json:"likes"`
}

// Comment represents a comment to be returned as a response
type Comment struct {
	ID       string `json:"id"`
	RoomID   string `json:"room_id"`
	ParentID string `json:"parent_id"`
	UserID   string `json:"user_id"`
	PageNo   int    `json:"page_no"`
	Message  string `json:"message"`
	Likes    int    `json:"likes"`
}
