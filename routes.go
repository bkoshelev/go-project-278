package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bkoshelev/go-project-278/db"
	"github.com/bkoshelev/go-project-278/internal/service"
	"github.com/gin-gonic/gin"
)

func ping(router *gin.Engine) *gin.Engine {
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	return router
}

func createLink(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.POST("/api/links", func(c *gin.Context) {
		var req CreateLinkPayload
		if err := bindPayload(c, &req); err != nil {
			return
		}

		shortLink, err := services.CreateShortLink(req.OriginalUrl, req.ShortName)
		if err.Err != nil {
			handleServiceError(c, err)
			return
		}

		c.JSON(http.StatusCreated, shortLink)
	})

	return router
}

func getShortLinks(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.GET("/api/links", func(c *gin.Context) {
		query := GetMulitpleEntityQueryParams{Range: Range{Begin: 0, End: 10}}

		if err := bindQuery(c, &query); err != nil {
			return
		}
		begin := query.Range.Begin
		end := query.Range.End

		shortLinks, err2 := services.GetLinks(
			db.GetShortLinksParams{
				Limit:  end - begin + 1,
				Offset: int32(begin),
			},
		)
		if err2.Err != nil {
			handleServiceError(c, err2)
			return
		}

		countLinks, err2 := services.CountLinks()
		if err2.Err != nil {
			handleServiceError(c, err2)
			return
		}

		c.Header("Content-Range", fmt.Sprintf(
			"links %v-%v/%v", begin, end, countLinks))
		c.JSON(http.StatusOK, shortLinks)
	})

	return router
}

func getShortLinkById(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.GET("/api/links/:id", func(c *gin.Context) {
		params := GetEntityUriParams{}
		if err := bindUri(c, &params); err != nil {
			return
		}

		shortLink, err := services.GetLinkById(int32(params.ID))

		if err.Err != nil {
			handleServiceError(c, err)
			return
		}

		c.JSON(http.StatusOK, shortLink)
	})

	return router
}

func updateLink(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.PUT("/api/links/:id", func(c *gin.Context) {
		params := GetEntityUriParams{}
		if err := bindUri(c, &params); err != nil {
			return
		}

		var req CreateLinkPayload
		if err := bindPayload(c, &req); err != nil {
			return
		}

		fmt.Println("UPDATE LINK", params.ID, req.OriginalUrl, req.ShortName)
		log.Println("UPDATE LINK", params.ID, req.OriginalUrl, req.ShortName)

		updatedShortLink, err := services.UpdateShortLink(
			int32(params.ID),
			req.OriginalUrl,
			req.ShortName,
		)

		if err.Err != nil {
			handleServiceError(c, err)
			return
		}

		c.JSON(http.StatusOK, updatedShortLink)
	})

	return router
}

func deleteLink(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.DELETE("/api/links/:id", func(c *gin.Context) {
		params := GetEntityUriParams{}
		if err := bindUri(c, &params); err != nil {
			return
		}

		err := services.DeleteShortLink(int32(params.ID))

		if err.Err != nil {
			handleServiceError(c, err)
			return
		}

		c.Status(http.StatusNoContent)
	})

	return router
}

func getLinkVisits(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.GET("/api/link_visits", func(c *gin.Context) {
		query := GetMulitpleEntityQueryParams{Range: Range{Begin: 0, End: 10}}

		if err := bindQuery(c, &query); err != nil {
			return
		}
		begin := query.Range.Begin
		end := query.Range.End

		shortLinks, err2 := services.GetLinkVisits(
			db.GetLinkVisitsParams{
				Limit:  end - begin + 1,
				Offset: int32(begin),
			},
		)

		if err2.Err != nil {
			handleServiceError(c, err2)
			return
		}

		countLinks, err2 := services.CountLinkVisits()

		if err2.Err != nil {
			handleServiceError(c, err2)
			return
		}

		c.Header("Content-Range", fmt.Sprintf(
			"link_visits %v-%v/%v", begin, end, countLinks))
		c.JSON(http.StatusOK, shortLinks)
	})

	return router
}

func redirectShortLink(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.GET("/r/:code", func(c *gin.Context) {
		params := RedirectUriParams{}
		if err := bindUri(c, &params); err != nil {
			return
		}

		shortLink, err := services.GetLinkByShortName(params.ShortName)

		if err.Err != nil {
			handleServiceError(c, err)
			return
		}

		_, err = services.CreateLinkVisit(
			c.ClientIP(),
			shortLink.ID,
			c.Request.UserAgent(),
			c.Request.Referer(),
			http.StatusFound,
		)

		if err.Err != nil {
			handleServiceError(c, err)
			return
		}

		c.Redirect(http.StatusFound, shortLink.OriginalUrl)
	})

	return router
}

func unknownRoute(router *gin.Engine) *gin.Engine {
	router.NoRoute(func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})
	return router
}
