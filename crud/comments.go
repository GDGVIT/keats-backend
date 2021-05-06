package crud

import (
	"github.com/Krishap-s/keats-backend/models"
	"github.com/Krishap-s/keats-backend/pgdb"
	"github.com/Krishap-s/keats-backend/schemas"
	"github.com/google/uuid"
)

// CreateComment creates a chatmessage in the database or returns an error
func CreateComment(objIn *schemas.CommentCreate) (*models.Comment, error) {
	db := pgdb.GetDB()
	cid, err := uuid.Parse(objIn.ClubID)
	if err != nil {
		return nil, err
	}
	uid, err := uuid.Parse(objIn.UserID)
	if err != nil {
		return nil, err
	}
	pid, err := uuid.Parse(objIn.ParentID)
	comment := &models.Comment{
		PageNo:   objIn.PageNo,
		Message:  objIn.Message,
		ClubID:   cid,
		UserID:   uid,
		ParentID: pid,
	}
	_, err = db.Model(comment).Returning("*").Insert()
	if err != nil {
		return nil, err
	}

	return comment, nil
}

// GetComment gets chatmessages from a room or returns an error
func GetComment(cid string) ([]*schemas.Comment, error) {
	db := pgdb.GetDB()
	var comments []*schemas.Comment
	err := db.Model((*models.Comment)(nil)).
		Where("club_id = ?", cid).
		Select(&comments)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

// AddCommentLike increments likes field of chatmessage
func AddCommentLike(id string) error {
	db := pgdb.GetDB()
	comment := models.Comment{}
	_, err := db.Model(&comment).
		Set("likes = likes + 1").
		Where("id = ?", id).
		Returning("id").
		UpdateNotZero()
	if err != nil {
		return err
	}
	return nil
}
