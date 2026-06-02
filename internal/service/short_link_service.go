package service

import (
	"errors"

	"github.com/bkoshelev/go-project-278/internal/db"
	"github.com/bkoshelev/go-project-278/internal/gen_id"
)

type ShortLinksService struct {
	q           *db.Queries
	idGenerator gen_id.IdGenerator
}

func NewShortLinksService(q *db.Queries, id_gen gen_id.IdGenerator) *ShortLinksService {
	return &ShortLinksService{q, id_gen}
}

var (
	ErrDublicate = errors.New("короткая ссылка уже существует")
	ErrShortName = errors.New("ошибка создания короткого имени. Попробуйте еще раз")
	ErrDB        = errors.New("ошибка взаимодействия с базой данных")
	ErrNoRows    = errors.New("короткая ссылка не найдена")
)
