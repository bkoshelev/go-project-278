package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bkoshelev/go-project-278/db"
	"github.com/bkoshelev/go-project-278/internal/service"
	"github.com/gin-gonic/gin"
)

func Ping(router *gin.Engine) *gin.Engine {
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	return router
}

func CreateLink(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.POST("/api/links", func(c *gin.Context) {
		var req CreateLinkPayload
		if err := bindPayload(c, &req); err != nil {

			if ve, ok := errors.AsType[ValidationError](err); ok {
				c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": ve.toJSON()})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": InvalidErr})
			}

			return
		}

		shortLink, err := services.CreateShortLink(c, req.OriginalUrl, req.ShortName)

		if err != nil {
			se, ok := errors.AsType[service.ServiceError](err)
			switch {
			case ok && errors.Is(se.Err, service.ErrShortName):
			case ok && errors.Is(se.Err, service.ErrDuplicate):
				verr := ValidationError{FieldName: se.FieldName, Err: se.Err}
				c.JSON(http.StatusBadRequest, gin.H{"errors": verr.toJSON()})
			case errors.Is(se.Err, service.ErrDB):
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusCreated, shortLink)
	})

	return router
}

func GetShortLinks(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.GET("/api/links", func(c *gin.Context) {
		query := GetMultipleEntityQueryParams{Range: Range{Begin: 0, End: 10}}
		if err := c.BindQuery(&query); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": InvalidErr})
			return
		}

		begin := query.Range.Begin
		end := query.Range.End

		shortLinks, err := services.GetLinks(c,
			db.GetShortLinksParams{
				Limit:  end - begin + 1,
				Offset: int32(begin),
			},
		)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrDB):
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
			return
		}

		countLinks, err := services.CountLinks(c)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrDB):
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
			return
		}

		c.Header("Content-Range", fmt.Sprintf(
			"links %v-%v/%v", begin, end, countLinks))
		c.JSON(http.StatusOK, shortLinks)
	})

	return router
}

func GetShortLinkById(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.GET("/api/links/:id", func(c *gin.Context) {
		params := GetEntityUriParams{}
		if err := c.ShouldBindUri(&params); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return

		}

		fmt.Println("PARAMS: ", params)

		shortLink, err := services.GetLinkById(c, int32(params.ID))
		if err != nil {
			se, ok := errors.AsType[service.ServiceError](err)

			switch {
			case ok && errors.Is(se.Err, service.ErrNoRows):
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			case errors.Is(se.Err, service.ErrDB):
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, shortLink)
	})

	return router
}

func UpdateLink(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.PUT("/api/links/:id", func(c *gin.Context) {
		params := GetEntityUriParams{}
		if err := c.ShouldBindUri(&params); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		}

		var req CreateLinkPayload
		if err := bindPayload(c, &req); err != nil {
			if ve, ok := errors.AsType[ValidationError](err); ok {
				c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": ve.toJSON()})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": InvalidErr})
			}
			return
		}

		updatedShortLink, err := services.UpdateShortLink(
			c,
			int32(params.ID),
			req.OriginalUrl,
			req.ShortName,
		)

		if err != nil {
			se, ok := errors.AsType[service.ServiceError](err)
			switch {
			case ok && errors.Is(se.Err, service.ErrNoRows):
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			case ok && errors.Is(se.Err, service.ErrDuplicate):
				verr := ValidationError{FieldName: se.FieldName, Err: se.Err}
				c.JSON(http.StatusBadRequest, gin.H{"errors": verr.toJSON()})
			case errors.Is(err, service.ErrDB):
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, updatedShortLink)
	})

	return router
}

func DeleteLink(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.DELETE("/api/links/:id", func(c *gin.Context) {
		params := GetEntityUriParams{}
		if err := c.ShouldBindUri(&params); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		}

		err := services.DeleteShortLink(c, int32(params.ID))
		if err != nil {
			se, ok := errors.AsType[service.ServiceError](err)
			switch {
			case ok && errors.Is(se.Err, service.ErrNoRows):
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			case errors.Is(se.Err, service.ErrDB):
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
			return
		}

		c.Status(http.StatusNoContent)
	})

	return router
}

func GetLinkVisits(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.GET("/api/link_visits", func(c *gin.Context) {
		query := GetMultipleEntityQueryParams{Range: Range{Begin: 0, End: 10}}

		if err := c.BindQuery(&query); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": InvalidErr})
			return
		}

		begin := query.Range.Begin
		end := query.Range.End

		shortLinks, err := services.GetLinkVisits(
			c,
			db.GetLinkVisitsParams{
				Limit:  end - begin + 1,
				Offset: int32(begin),
			},
		)

		if err != nil {
			switch {
			case errors.Is(err, service.ErrDB):
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
			return
		}

		countLinks, err := services.CountLinkVisits(c)

		if err != nil {
			switch {
			case errors.Is(err, service.ErrDB):
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
			return
		}

		c.Header("Content-Range", fmt.Sprintf(
			"link_visits %v-%v/%v", begin, end, countLinks))
		c.JSON(http.StatusOK, shortLinks)
	})

	return router
}

func RedirectShortLink(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.GET("/r/:code", func(c *gin.Context) {
		params := RedirectUriParams{}
		if err := c.ShouldBindUri(&params); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		}

		shortLink, err := services.GetLinkByShortName(c, params.ShortName)

		if err != nil {
			se, ok := errors.AsType[service.ServiceError](err)

			switch {
			case ok && errors.Is(se.Err, service.ErrNoRows):
				verr := ValidationError{FieldName: se.FieldName, Err: se.Err}
				c.JSON(http.StatusNotFound, gin.H{"errors": verr.toJSON()})
			case errors.Is(se.Err, service.ErrDB):
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
			return
		}

		_, err = services.CreateLinkVisit(
			c,
			c.ClientIP(),
			shortLink.ID,
			c.Request.UserAgent(),
			c.Request.Referer(),
			http.StatusFound,
		)

		if err != nil {
			se, ok := errors.AsType[service.ServiceError](err)

			switch {
			case ok && errors.Is(se.Err, service.ErrIp):
				verr := ValidationError{FieldName: se.FieldName, Err: se.Err}
				c.JSON(http.StatusBadRequest, gin.H{"errors": verr.toJSON()})
			case errors.Is(se.Err, service.ErrDB):
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
			return
		}

		c.Redirect(http.StatusFound, shortLink.OriginalUrl)
	})

	return router
}

func UnknownRoute(router *gin.Engine) *gin.Engine {

	router.NoRoute(func(c *gin.Context) {

		c.Status(http.StatusNotFound)

	})

	return router

}
