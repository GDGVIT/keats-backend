package schemas

// RoomCreate represents a room to be created
type RoomCreate struct {
	ID       string `json:"id"`
	FileURL  string `json:"file_url"`
	PageNo   int    `json:"page_no"`
	Private  bool   `json:"private"`
	PageSync bool   `json:"page_sync"`
}

// RoomUpdate represents a room to be updated
type RoomUpdate struct {
	ID       string `json:"id"`
	FileURL  string `json:"file_url"`
	PageNo   int    `json:"page_no"`
	Private  bool   `json:"private"`
	PageSync bool   `json:"page_sync"`
}

// Room represents a room to be returned as a response
type Room struct {
	ID       string `json:"id"`
	FileURL  string `json:"file_url"`
	PageNo   int    `json:"page_no"`
	Private  bool   `json:"private"`
	PageSync bool   `json:"page_sync"`
}
