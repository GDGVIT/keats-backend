package schemas

// UserCreate represents a user to be created
type UserCreate struct {
	PhoneNo string `json:"phone_number"`
}

// UserUpdate represents a user to be updated
type UserUpdate struct {
	Username string `json:"username"`
	PhoneNo string `json:"phone_number"`
	ProfilePic string `json:"profile_pic"`
	Email string `json:"email"`
	Bio string `json:"bio"`
}

// User represents a user to be returned as a response
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	PhoneNo string `json:"phone_number"`
	ProfilePic string `json:"profile_pic"`
	Email string `json:"email"`
	Bio string `json:"bio"`
}

// UserDelete represents a user that has been deleted
type UserDelete struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	PhoneNo string `json:"phone_number"`
	ProfilePic string `json:"profile_pic"`
	Email string `json:"email"`
	Bio string `json:"bio"`
}
