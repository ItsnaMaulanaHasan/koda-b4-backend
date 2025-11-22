package utils

import (
	"backend-daily-greens/lib"
	"fmt"
	"net/url"

	"github.com/gin-gonic/gin"
)

func BuildHateoasPagination(ctx *gin.Context, page int, limit int, search string, total int) lib.HateoasLink {
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	rawQuery := ctx.Request.URL.Query()

	makeURL := func(p int) string {
		q := url.Values{}

		for key, vals := range rawQuery {
			for _, v := range vals {
				q.Add(key, v)
			}
		}

		q.Set("page", fmt.Sprintf("%d", p))
		q.Set("limit", fmt.Sprintf("%d", limit))
		if search != "" {
			q.Set("search", search)
		}

		return fmt.Sprintf("%s://%s%s?%s",
			getScheme(ctx),
			ctx.Request.Host,
			ctx.FullPath(),
			q.Encode(),
		)
	}

	return lib.HateoasLink{
		Self: makeURL(page),
		Next: func() any {
			if page < totalPages {
				return makeURL(page + 1)
			}
			return nil
		}(),
		Prev: func() any {
			if page > 1 {
				return makeURL(page - 1)
			}
			return nil
		}(),
		Last: makeURL(totalPages),
	}
}

func BuildHateoasPaginationHistories(ctx *gin.Context, page int, limit int, date string, statusId int, total int) lib.HateoasLink {
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	rawQuery := ctx.Request.URL.Query()

	makeURL := func(p int) string {
		q := url.Values{}

		for key, vals := range rawQuery {
			for _, v := range vals {
				q.Add(key, v)
			}
		}

		q.Set("page", fmt.Sprintf("%d", p))
		q.Set("limit", fmt.Sprintf("%d", limit))
		if date != "" {
			q.Set("date", date)
		}
		if statusId != 0 {
			q.Set("statusid", fmt.Sprintf("%d", statusId))
		}

		return fmt.Sprintf("%s://%s%s?%s",
			getScheme(ctx),
			ctx.Request.Host,
			ctx.FullPath(),
			q.Encode(),
		)
	}

	return lib.HateoasLink{
		Self: makeURL(page),
		Next: func() any {
			if page < totalPages {
				return makeURL(page + 1)
			}
			return nil
		}(),
		Prev: func() any {
			if page > 1 {
				return makeURL(page - 1)
			}
			return nil
		}(),
		Last: makeURL(totalPages),
	}
}

func getScheme(ctx *gin.Context) string {
	if ctx.Request.TLS != nil {
		return "https"
	}
	return "http"
}
