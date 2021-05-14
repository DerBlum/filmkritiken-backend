package inbound

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/DerBlum/filmkritiken-backend/domain/filmkritiken"
	gin "github.com/gin-gonic/gin"
)

type (
	FilmRequest struct {
		Film         *filmkritiken.Film `json:"film"`
		Von          string             `json:"von"`
		BesprochenAm *time.Time         `json:"besprochenam"`
	}

	filmkritikenHandler struct {
		filmkritikenService filmkritiken.FilmkritikenService
	}
)

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
	if err == nil {
		limit = parsedValue
	}
	parsedValue, err = parseIntFromQueryParam(queryParams, "offset")
	if err == nil {
		offset = parsedValue
	}

	filter := &filmkritiken.FilmkritikenFilter{
		Limit:  limit,
		Offset: offset,
	}
	result, err := h.filmkritikenService.GetFilmkritiken(ctx.Request.Context(), filter)
	if err != nil {
		// TODO better error handling
		ctx.Writer.WriteHeader(http.StatusInternalServerError)
		ctx.Writer.WriteString(err.Error())
		return
	}

	ctx.JSON(200, result)

}

func (h *filmkritikenHandler) handleCreateFilm(ctx *gin.Context) {

	req := &FilmRequest{}
	err := ctx.ShouldBindJSON(req)
	if err != nil {
		// TODO better error handling
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	result, err := h.filmkritikenService.CreateFilm(ctx.Request.Context(), req.Film, req.Von, req.BesprochenAm)
	if err != nil {
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
