package storage

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/pkg/errors"
)

type Storage struct {
	db *DB
}

func Init() *Storage {
	db, err := DBInit()
	if err != nil {
		log.Fatalf("Failed to initialize storage: %s", err)
	}

	return &Storage{
		db: db,
	}
}

func (s *Storage) FindIssue(issueId string) (*Issue, error) {
	issue := &Issue{}
	if err := s.db.GetOne(
		issue,
		"select * from issues where issue_id=?",
		issueId,
	); err != nil {
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
		}
		return nil, errors.Wrap(err, "Error finding issue")
	}
	return issue, nil
}

func (s *Storage) SaveNewIssue(issue *Issue) error {
	fmt.Println("\tSaving new issue")
	if err := s.db.Insert(issue); err != nil {
		return errors.Wrap(err, "Error saving new issue")
	}
	if err := s.saveComments(issue.Comments); err != nil {
		return errors.Wrap(err, "Error saving new issue")
	}

	return nil
}

func (s *Storage) SaveNewCard(card *Card) error {
	if err := s.db.Insert(card); err != nil {
		return errors.Wrap(err, "Error saving new card")
	}
	return nil
}

func (s *Storage) saveComments(comments []*Comment) error {
	for _, comment := range comments {
		if err := s.db.Insert(comment); err != nil {
			return errors.Wrap(err, "Error saving new comment")
		}
	}
	return nil
}

func (s *Storage) UpdateCard(card *Card) (int64, error) {
	count, err := s.db.Update(card)
	if err != nil {
		return 0, errors.Wrap(err, "Error updating card")
	}
	return count, nil
}

func (s *Storage) Close() {
	s.db.dbMap.Db.Close()
}
