package service

import (
	"context"
	"net/netip"

	"github.com/bkoshelev/go-project-278/db"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *ShortLinksService) CreateLinkVisit(ip string, linkId int32, userAgent string, referer string, status int32) (db.CreateLinkVisitRow, error) {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return db.CreateLinkVisitRow{}, ErrDB
	}

	linkVisit, err := s.q.CreateLinkVisit(context.Background(), db.CreateLinkVisitParams{
		Ip:        addr,
		LinkID:    pgtype.Int4{Int32: linkId, Valid: true},
		UserAgent: userAgent,
		Referer:   referer,
		Status:    status,
	})
	if err != nil {
		return db.CreateLinkVisitRow{}, err
	}

	return linkVisit, nil
}
