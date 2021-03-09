package crud

import (
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
	if err != nil {
		return nil, err
	}
	user := &models.User{
		ID:       uid,
		PhoneNo:  objIn.PhoneNo,
		Username: objIn.Username,
		Email:    objIn.Email,
		Bio:      objIn.Bio,
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

// DeleteUser deletes an existing user or returns an error
func DeleteUser(id string) (*models.User, error) {
	db := pgdb.GetDB()
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	user := &models.User{
		ID: uid,
	}

	_, err = db.Model(user).WherePK().Delete()
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserClub gets clubuser records from the database
func GetUserClub(id string) ([]*models.Club, error) {
	db := pgdb.GetDB()
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	var clubs []*models.Club
	err = db.Model(&clubs).
		Join("INNER JOIN club_users as cu").
		JoinOn("cu.club_id = club.id").
		Where("cu.user_id = ?", uid).
		Select()
	if err != nil {
		return nil, err
	}
	return clubs, nil
}
