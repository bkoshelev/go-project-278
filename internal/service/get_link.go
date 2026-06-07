package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/bkoshelev/go-project-278/db"
)

func (s *ShortLinksService) GetLinkById(id int32) (db.ShortLink, ServiceError) {

	shortLink, err := s.q.GetShortLinkById(context.Background(), id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.ShortLink{}, ServiceError{"id", ErrNoRows}
		}
		return db.ShortLink{}, ServiceError{"db", err}
	}
	return shortLink, ServiceError{}
}

func (s *ShortLinksService) GetLinkByShortName(shortName string) (db.ShortLink, ServiceError) {
	shortLink, err := s.q.GetShortLinkByShortName(context.Background(), shortName)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.ShortLink{}, ServiceError{"short_name", ErrNoRows}
		}
		return db.ShortLink{}, ServiceError{"db", err}

	}
	return shortLink, ServiceError{}
}
