package crud

import (
//	"github.com/google/uuid"

	"github.com/Krishap-s/keats-backend/db"
	"github.com/Krishap-s/keats-backend/models"
	"github.com/Krishap-s/keats-backend/schemas"
)


// CreateUser creates a user in the database or returns an error
func CreateUser(objIn *schemas.UserCreate) (*models.User, error) {
	db := db.GetDB()
	if objIn.Username == "" {
		objIn.Username = "Blake"
	}
	user := &models.User{
		Username: objIn.Username,
		PhoneNo: objIn.PhoneNo,
	}

	_, err := db.Model(user).
		OnConflict("(phone_no) DO NOTHING").
		Returning("*").
		SelectOrInsert()
	return user, err
}

// UpdateUser updates an existing user in the database or returns an error
func UpdateUser(objIn *schemas.UserUpdate) (*models.User, error) {
	db := db.GetDB()

	user := &models.User{
		PhoneNo: objIn.PhoneNo,
		Username: objIn.Username,
		Email: objIn.Email,
		Bio: objIn.Bio,
	}

	_, err := db.Model(user).Returning("*").Where("phone_no = ?phone_no").UpdateNotZero()

	return user, err
}

// GetUser fetches an existing user or returns an error
func GetUser(phone_no string) (*models.User, error) {
	db := db.GetDB()

	user := &models.User{
		PhoneNo: phone_no,
	}

	err := db.Model(user).Where("phone_no = ?phone_no").Select()
	if err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser deletes an existing user or returns an error
func DeleteUser(phone_no string) (*models.User, error) {
	db := db.GetDB()

	user := &models.User{
		PhoneNo: phone_no,
	}

	_, err := db.Model(user).Where("phone_no = ?phone_no").Delete()
	if err != nil {
		return nil, err
	}

	return user, nil
}
