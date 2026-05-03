package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/bkoshelev/go-project-278/internal/db"
)

func (s *ShortLinksService) GetLinkById(id int32) (db.ShortLink, error) {

	shortLink, err := s.q.GetShortLinkById(context.Background(), id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.ShortLink{}, ErrNoRows
		}
		return db.ShortLink{}, ErrDB
	}
	return shortLink, nil
}
