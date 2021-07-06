package crud

import (
	"fmt"

	"github.com/Krishap-s/keats-backend/models"
	"github.com/Krishap-s/keats-backend/pgdb"
	"github.com/Krishap-s/keats-backend/schemas"
	"github.com/google/uuid"
)

func parseClubUser(clubID string, userID string) (*models.ClubUser, error) {
	cid, err := uuid.Parse(clubID)
	if err != nil {
		return nil, err
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	clubuser := &models.ClubUser{
		ClubID: cid,
		UserID: uid,
	}
	return clubuser, nil
}

// CreateUser creates a club in the database or returns an error
func CreateClub(objIn *schemas.ClubCreate) (*models.Club, error) {
	db := pgdb.GetDB()
	uid, err := uuid.Parse(objIn.HostID)
	if err != nil {
		return nil, err
	}
	if len(objIn.ClubName) > 30 || len(objIn.ClubPic) > 100 {
		return nil, fmt.Errorf("max string length")
	}
	club := &models.Club{
		ClubName: objIn.ClubName,
		ClubPic:  objIn.ClubPic,
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

	clubuser := &models.ClubUser{
		ClubID: club.ID,
		UserID: uid,
	}
	_, err = db.Model(clubuser).Returning("*").Insert()
	if err != nil {
		return nil, err
	}
	return club, nil
}

// UpdateClub updates a club in the database or returns an error
func UpdateClub(objIn *schemas.ClubUpdate) (*models.Club, error) {
	db := pgdb.GetDB()
	uid, err := uuid.Parse(objIn.ID)
	if err != nil {
		return nil, err
	}
	if len(objIn.ClubName) > 30 || len(objIn.ClubPic) > 100 {
		return nil, fmt.Errorf("max string length")
	}
	club := &models.Club{
		ID:       uid,
		ClubName: objIn.ClubName,
		ClubPic:  objIn.ClubPic,
		FileURL:  objIn.FileURL,
		PageNo:   objIn.PageNo,
	}

	_, err = db.Model(club).
		Column("club_name").
		Column("file_url").
		Column("club_pic").
		Column("page_no").
		Returning("*").
		WherePK().
		UpdateNotZero()
	if err != nil {
		return nil, err
	}

	return club, nil
}

// TogglePrivate toggles the private status of a club
func TogglePrivate(clubID string) error {
	db := pgdb.GetDB()
	_, err := db.Model((*models.Club)(nil)).
		Set("private = NOT private").
		Where("id = ?", clubID).
		Update()
	if err != nil {
		return err
	}
	return nil
}

// ToggleSync toggles the page sync feature of a club
func ToggleSync(clubID string) error {
	db := pgdb.GetDB()
	_, err := db.Model((*models.Club)(nil)).
		Set("page_sync = NOT page_sync").
		Where("id = ?", clubID).
		Update()
	if err != nil {
		return err
	}
	return nil
}

// ListClub gets all non-private clubs from database or returns an error
func ListClub(userID string, n int) ([]*schemas.Club, error) {
	db := pgdb.GetDB()
	var clubs []*schemas.Club
	err := db.Model((*models.Club)(nil)).
		ColumnExpr("club.id,club.club_name,club.club_pic,club.file_url,club.page_no,club.private,club.host_id,u.id as host_id,u.username as host_name,u.profile_pic as host_profile_pic").
		Join("INNER JOIN users as u").
		JoinOn("club.host_id = u.id").
		Where("private = false").
		Where("NOT EXISTS (SELECT * FROM club_users cu WHERE cu.club_id = club.id AND cu.user_id = ?)", userID).
		Offset((n - 1) * 10).
		Limit(10).
		Select(&clubs)
	if err != nil {
		return nil, err
	}
	return clubs, nil
}

// GetClub gets a club from the database or returns an error
func GetClub(id string) (*schemas.Club, error) {
	db := pgdb.GetDB()
	cid, err := uuid.Parse(id)
	res := new(schemas.Club)
	if err != nil {
		return nil, err
	}
	club := &models.Club{
		ID: cid,
	}
	err = db.Model(club).
		ColumnExpr("club.id,club.club_name,club.club_pic,club.file_url,club.page_no,club.private,club.host_id,u.id as host_id,u.username as host_name,u.profile_pic as host_profile_pic").
		Join("INNER JOIN users as u").
		JoinOn("club.host_id = u.id").
		WherePK().
		Select(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// CreateClubUser creates a clubuser record in the database
func CreateClubUser(clubID string, userID string) (*models.ClubUser, error) {
	db := pgdb.GetDB()
	clubuser, err := parseClubUser(clubID, userID)
	if err != nil {
		return nil, err
	}
	_, err = db.Model(clubuser).Returning("*").Insert()
	if err != nil {
		return nil, err
	}
	return clubuser, err
}

// GetClubUser get clubuser records from database
func GetClubUser(clubID string) ([]*models.User, error) {
	db := pgdb.GetDB()
	cid, err := uuid.Parse(clubID)
	if err != nil {
		return nil, err
	}
	var users []*models.User
	err = db.Model(&users).
		Join("INNER JOIN club_users as cu").
		JoinOn("cu.user_id = \"user\".id").
		Where("cu.club_id = ?", cid).
		Select()
	if err != nil {
		return nil, err
	}
	return users, nil
}

// DeleteClubUser deletes clubuser record from database
func DeleteClubUser(clubID string, userID string) (*models.ClubUser, error) {
	db := pgdb.GetDB()
	clubuser, err := parseClubUser(clubID, userID)
	if err != nil {
		return nil, err
	}
	cid := clubuser.ClubID
	uid := clubuser.UserID
	_, err = db.Model(clubuser).Where("user_id = ?user_id and club_id = ?club_id").Returning("*").Delete()
	if err != nil {
		return nil, err
	}
	var users []*models.User
	err = db.Model(&users).
		ColumnExpr("\"user\".\"id\" , \"user\".\"username\", \"user\".\"profile_pic\", \"user\".\"phone_no\", \"user\".\"email\", \"user\".\"bio\"").
		Join("INNER JOIN club_users as cu").
		JoinOn("cu.user_id = \"user\".\"id\"").
		Where("cu.club_id = ?", cid).
		Select()
	if err != nil {
		return nil, err
	}
	// Reset Host ID to someone else if host themselves is leaving
	club := &models.Club{
		ID: cid,
	}
	err = db.Model(club).WherePK().Select()
	if err != nil {
		return nil, err
	}
	if club.HostID == uid {
		if len(users) != 0 {
			club.HostID = users[0].ID
		} else {
			club.HostID = uuid.Nil
			club.Private = true
		}
		_, err = db.Model(club).WherePK().Update()
		if err != nil {
			return nil, err
		}
	}
	return clubuser, err
}
