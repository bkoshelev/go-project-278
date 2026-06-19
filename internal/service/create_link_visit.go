package service

import (
	"errors"
	"net/netip"

	"github.com/bkoshelev/go-project-278/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *ShortLinksService) CreateLinkVisit(c *gin.Context, ip string, linkID int32, userAgent, referer string, status int32) (db.CreateLinkVisitRow, error) {
	ctx := c.Request.Context()

	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return db.CreateLinkVisitRow{}, ServiceError{"ip", err}
	}

	linkVisit, err := s.q.CreateLinkVisit(ctx, db.CreateLinkVisitParams{
		IP:        addr,
		LinkID:    pgtype.Int4{Int32: linkID, Valid: true},
		UserAgent: userAgent,
		Referer:   referer,
		Status:    status,
	})
	if err != nil {
		return db.CreateLinkVisitRow{}, errors.Join(ErrDB, err)
	}

	return linkVisit, nil
}
