package service

import (
	"errors"

	"github.com/bkoshelev/go-project-278/internal/db"
)

type ShortLinksService struct {
	q *db.Queries
}

func NewShortLinksService(q *db.Queries) *ShortLinksService {
	return &ShortLinksService{q}
}

var (
	ErrDublicate = errors.New("короткая ссылка уже существует")
	ErrShortName = errors.New("ошибка создания короткого имени. Попробуйте еще раз")
	ErrDB        = errors.New("ошибка взаимодействия с базой данных")
	ErrNoRows    = errors.New("короткая ссылка не найдена")
)
