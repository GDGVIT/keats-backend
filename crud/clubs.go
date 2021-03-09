package crud

import (
	"github.com/Krishap-s/keats-backend/models"
	"github.com/Krishap-s/keats-backend/pgdb"
	"github.com/Krishap-s/keats-backend/schemas"
	"github.com/google/uuid"
)

// CreateUser creates a user in the database or returns an error
func CreateClub(objIn *schemas.ClubCreate) (*models.Club, error) {
	db := pgdb.GetDB()
	uid, err := uuid.Parse(objIn.HostID)
	if err != nil {
		return nil, err
	}
	club := &models.Club{
		ClubName: objIn.ClubName,
		PageSync: objIn.PageSync,
		FileURL:  objIn.FileURL,
		Private:  objIn.Private,
		PageNo:   objIn.PageNo,
		HostID:   uid,
	}

	_, err = db.Model(club).
		Returning("*").
		Insert()
	if err != nil {
		return nil, err
	}

	return club, nil
}
