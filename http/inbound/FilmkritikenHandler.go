package inbound

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/DerBlum/filmkritiken-backend/domain/filmkritiken"
	gin "github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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
		log.Errorf("Could not get Filmkritiken from DB: %v", err)
		ctx.Writer.WriteHeader(http.StatusInternalServerError)
		ctx.Writer.WriteString("Could not get Filmkritiken from DB")
		return
	}

	ctx.JSON(200, result)

}

func (h *filmkritikenHandler) handleCreateFilm(ginCtx *gin.Context) {

	req := &FilmRequest{}
	err := ginCtx.ShouldBindJSON(req)
	if err != nil {
		log.Error("Could not map json to FilmRequest")
		ginCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	result, err := h.filmkritikenService.CreateFilm(ginCtx.Request.Context(), req.Film, req.Von, req.BesprochenAm)
	if err != nil {
		ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		ginCtx.Writer.WriteString(err.Error())
		return
	}

	ginCtx.JSON(200, result)
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
