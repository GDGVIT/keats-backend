package crud

import (
	"github.com/google/uuid"

	"github.com/Krishap-s/keats-backend/db"
	"github.com/Krishap-s/keats-backend/models"
	"github.com/Krishap-s/keats-backend/schemas"
)

// CreateUser creates a user in the database or returns an error
func CreateUser(objIn *schemas.UserCreate) (*models.User, error) {
	db := db.GetDB()
	user := &models.User{
		Username: objIn.Username,
		IsActive: objIn.IsActive,
	}

	_, err := db.Model(user).
		OnConflict("DO NOTHING").
		Returning("*").
		Insert()
	return user, err
}

// UpdateUser updates an existing user in the database or returns an error
func UpdateUser(objIn *schemas.UserUpdate) (*models.User, error) {
	db := db.GetDB()

	// convert string in json body to UUID
	id, err := uuid.Parse(objIn.ID)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:       id,
		Username: objIn.Username,
		IsActive: objIn.IsActive,
	}

	_, err = db.Model(user).Returning("*").WherePK().UpdateNotZero()

	return user, err
}

// Parse parses UUID to user and returns the user or returns an error
func Parse(userID string) (*models.User, error) {
	// convert string in body to UUID
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	user := &models.User{ID: id}
	return user, nil

}

// GetUser fetches an existing user or returns an error
func GetUser(userID string) (*models.User, error) {
	db := db.GetDB()

	user, err := Parse(userID)
	if err != nil {
		return nil, err
	}

	err = db.Model(user).WherePK().Select()
	if err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser deletes an existing user or returns an error
func DeleteUser(userID string) (*models.User, error) {
	db := db.GetDB()

	user, err := Parse(userID)
	if err != nil {
		return nil, err
	}

	_, err = db.Model(user).WherePK().Delete()
	if err != nil {
		return nil, err
	}

	return user, nil
}
