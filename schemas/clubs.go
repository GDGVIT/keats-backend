package schemas

// ClubCreate represents a room to be created
type ClubCreate struct {
	ID       string `json:"id"`
	ClubName string `json:"clubname"`
	ClubPic  string `json:"club_pic"`
	FileURL  string `json:"file_url"`
	PageNo   int    `json:"page_no"`
	Private  bool   `json:"private"`
	PageSync bool   `json:"page_sync"`
	HostID   string `json:"host_id"`
}

// ClubUpdate represents a room to be updated
type ClubUpdate struct {
	ID       string `json:"id"`
	ClubName string `json:"clubname"`
	ClubPic  string `json:"club_pic"`
	FileURL  string `json:"file_url"`
	PageNo   int    `json:"page_no"`
	Private  bool   `json:"private"`
	PageSync bool   `json:"page_sync"`
	HostID   string `json:"host_id"`
}

// Club represents a room to be returned as a response
type Club struct {
	ID             string `json:"id"`
	ClubName       string `json:"clubname"`
	ClubPic        string `json:"club_pic"`
	FileURL        string `json:"file_url"`
	PageNo         int    `json:"page_no"`
	Private        bool   `json:"private"`
	PageSync       bool   `json:"page_sync"`
	HostID         string `json:"host_id"`
	HostName       string `json:"host_name"`
	HostProfilePic string `json:"host_profile_pic"`
}
