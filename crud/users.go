package crud

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/Krishap-s/keats-backend/models"
	"github.com/Krishap-s/keats-backend/pgdb"
	"github.com/Krishap-s/keats-backend/schemas"
)

// CreateUser creates a user in the database or returns an error
func CreateUser(objIn *schemas.UserCreate) (*models.User, error) {
	db := pgdb.GetDB()
	if objIn.Username == "" {
		objIn.Username = "Blake"
	}
	if len(objIn.Username) > 30 || len(objIn.Email) > 50 {
		return nil, fmt.Errorf("max string length")
	}

	user := &models.User{
		Username: objIn.Username,
		PhoneNo:  objIn.PhoneNo,
	}

	_, err := db.Model(user).
		Where("phone_no = ?phone_no").
		OnConflict("(phone_no) DO NOTHING").
		Returning("*").
		SelectOrInsert()
	return user, err
}

// UpdateUser updates an existing user in the database or returns an error
func UpdateUser(objIn *schemas.UserUpdate) (*models.User, error) {
	db := pgdb.GetDB()

	uid, err := uuid.Parse(objIn.ID)
	if len(objIn.Username) > 30 || len(objIn.Email) > 50 || len(objIn.Bio) > 100 || len(objIn.ProfilePic) > 100 {
		return nil, fmt.Errorf("max string length")
	}
	if err != nil {
		return nil, err
	}
	user := &models.User{
		ID:         uid,
		PhoneNo:    objIn.PhoneNo,
		ProfilePic: objIn.ProfilePic,
		Username:   objIn.Username,
		Email:      objIn.Email,
		Bio:        objIn.Bio,
	}

	_, err = db.Model(user).Returning("*").WherePK().UpdateNotZero()

	return user, err
}

// GetUser fetches an existing user or returns an error
func GetUser(id string) (*models.User, error) {
	db := pgdb.GetDB()

	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	user := &models.User{
		ID: uid,
	}

	err = db.Model(user).WherePK().Select()
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserClub gets clubuser records from the database
func GetUserClub(id string, n int) ([]*schemas.Club, error) {
	db := pgdb.GetDB()
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	var clubs []*schemas.Club
	err = db.Model((*models.Club)(nil)).
		ColumnExpr("club.id,club.club_name,club.club_pic,club.file_url,club.page_no,club.private,club.host_id,u.id as host_id,u.username as host_name,u.profile_pic as host_profile_pic").
		Join("INNER JOIN club_users as cu").
		JoinOn("cu.club_id = club.id").
		Join("INNER JOIN users as u").
		JoinOn("club.host_id = u.id").
		Where("cu.user_id = ?", uid).
		Offset((n - 1) * 10).
		Limit(10).
		Select(&clubs)
	if err != nil {
		return nil, err
	}
	return clubs, nil
}
