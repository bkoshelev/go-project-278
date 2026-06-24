package service

import (
	"errors"
	"fmt"

	"github.com/bkoshelev/go-project-278/db"
	"github.com/bkoshelev/go-project-278/internal/gen_id"
)

type ShortLinksService struct {
	q           *db.Queries
	idGenerator gen_id.IDGenerator
	host        string
}

func NewShortLinksService(q *db.Queries, id_gen gen_id.IDGenerator, host string) *ShortLinksService {
	return &ShortLinksService{q, id_gen, host}
}

type ServiceError struct {
	FieldName string
	Err       error
}

func (e ServiceError) Error() string {
	return e.Err.Error()
}

type DBError struct {
	Err error
}

func (e DBError) Error() string {
	return fmt.Sprintf("неизвестная ошибка взаимодействия с базой данных: %v", e.Err.Error())
}

func (e DBError) Unwrap() error {
	return e.Err
}

func (e DBError) Is(target error) bool {
	return target == ErrDB
}

var (
	ErrDuplicate = errors.New("short name already in use")
	ErrShortName = errors.New("ошибка создания короткого имени. Попробуйте еще раз")
	ErrDB        = errors.New("неизвестная ошибка взаимодействия с базой данных")
	ErrNoRows    = errors.New("короткая ссылка не найдена")
)
