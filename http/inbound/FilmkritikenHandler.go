package inbound

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/DerBlum/filmkritiken-backend/domain/filmkritiken"
	gin "github.com/gin-gonic/gin"
)

type filmkritikenHandler struct {
	filmkritikenService filmkritiken.FilmkritikenService
}

func NewFilmkritikenHandler(filmkritikenService filmkritiken.FilmkritikenService) *filmkritikenHandler {
	return &filmkritikenHandler{
		filmkritikenService: filmkritikenService,
	}
}

func (h *filmkritikenHandler) handleGetFilmkritiken(ctx *gin.Context) {

	limit := 10
	offset := 0

	queryParams := ctx.Request.URL.Query()
	parsedValue, err := parseIntFromQueryParam(queryParams, "limit")
	if err != nil {
		limit = parsedValue
	}
	parsedValue, err = parseIntFromQueryParam(queryParams, "offset")
	if err != nil {
		offset = parsedValue
	}

	result, err := h.filmkritikenService.GetFilmkritiken(limit, offset)
	if err != nil {
		// TODO better error handling
		ctx.Writer.WriteHeader(http.StatusInternalServerError)
		ctx.Writer.WriteString(err.Error())
		return
	}

	ctx.JSON(200, result)

}

func parseIntFromQueryParam(queryParams url.Values, paramName string) (int, error) {
	values := queryParams[paramName]
	if len(values) == 1 {
		parsedLimit, err := strconv.Atoi(values[0])
		if err != nil {
			return 0, err
		}
		return parsedLimit, nil
	}
	return 0, fmt.Errorf("could not parse query parameter %s to int", paramName)
}
