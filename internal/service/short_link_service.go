package service

import (
	"errors"

	"github.com/bkoshelev/go-project-278/db"
	"github.com/bkoshelev/go-project-278/internal/gen_id"
)

type ShortLinksService struct {
	q           *db.Queries
	idGenerator gen_id.IdGenerator
	host        string
}

func NewShortLinksService(q *db.Queries, id_gen gen_id.IdGenerator, host string) *ShortLinksService {
	return &ShortLinksService{q, id_gen, host}
}

type ServiceError struct {
	FieldName string
	Err       error
}

func (e ServiceError) Error() string {
	return e.Err.Error()
}

var (
	ErrDuplicate = errors.New("short name already in use")
	ErrShortName = errors.New("ошибка создания короткого имени. Попробуйте еще раз")
	ErrDB        = errors.New("неизвестная ошибка взаимодействия с базой данных")
	ErrNoRows    = errors.New("короткая ссылка не найдена")
	ErrIp        = errors.New("ошибка обработки ip адреса")
)
