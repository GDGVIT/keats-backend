package schemas

// ClubCreate represents a room to be created
type ClubCreate struct {
	ID       string `json:"id" form:"id"`
	ClubName string `json:"clubname" form:"clubname"`
	ClubPic  string `json:"club_pic" form:"club_pic"`
	FileURL  string `json:"file_url" form:"file_url"`
	PageNo   int    `json:"page_no" form:"page_no"`
	Private  bool   `json:"private" form:"private"`
	PageSync bool   `json:"page_sync" form:"page_sync"`
	HostID   string `json:"host_id" form:"host_id"`
}

// ClubUpdate represents a room to be updated
type ClubUpdate struct {
	ID       string `json:"id" form:"id"`
	ClubName string `json:"clubname" form:"clubname"`
	ClubPic  string `json:"club_pic"`
	FileURL  string `json:"file_url"`
	PageNo   int    `json:"page_no" form:"page_no"`
	HostID   string `json:"host_id" form:"host_id"`
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
