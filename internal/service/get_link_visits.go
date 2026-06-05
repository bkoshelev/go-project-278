package service

import (
	"context"
	"fmt"

	"github.com/bkoshelev/go-project-278/internal/db"
)

func (s *ShortLinksService) GetLinkVisits(params db.GetLinkVisitsParams) ([]db.GetLinkVisitsRow, error) {

	linkVisits, err := s.q.GetLinkVisits(context.Background(), params)

	if err != nil {
		fmt.Println("ошибка получения короткой ссылки", err)
		return nil, ErrDB
	}

	return linkVisits, nil
}
