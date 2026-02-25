package inbound

import (
	"encoding/json"
	"fmt"
	"github.com/DerBlum/filmkritiken-backend/domain/filmkritiken"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Have images cached for 30 days
const imageCacheDuration = 60 * 60 * 24 * 30

type (
	FilmRequest struct {
		Film           *filmkritiken.Film `json:"film"`
		Von            string             `json:"von"`
		BesprochenAm   *time.Time         `json:"besprochenam"`
		BewertungOffen bool               `json:"bewertungoffen"`
	}

	SetBewertungRequest struct {
		FilmkritikenId string `json:"filmkritikenId"`
		Wertung        int    `json:"wertung"`
	}

	SetBewertungBulkRequest struct {
		FilmkritikenId string               `json:"filmkritikenId"`
		Bewertungen    []*BenutzerBewertung `json:"benutzerBewertungen"`
	}

	SetBesprochenAmRequest struct {
		BesprochenAm time.Time `json:"besprochenam"`
	}

	BenutzerBewertung struct {
		Wertung  int    `json:"wertung"`
		Benutzer string `json:"benutzer"`
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

func (h *filmkritikenHandler) handleGetFilmkritiken(ginCtx *gin.Context) {
	limit := 10
	offset := 0

	queryParams := ginCtx.Request.URL.Query()
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
	result, err := h.filmkritikenService.GetFilmkritiken(ginCtx.Request.Context(), filter)
	if err != nil {
		log.Errorf("Could not get Filmkritiken from DB: %v", err)
		ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		_, _ = ginCtx.Writer.WriteString("Could not get Filmkritiken from DB")
		return
	}

	ginCtx.JSON(http.StatusOK, result)
}

func (h *filmkritikenHandler) handleCreateFilm(ginCtx *gin.Context) {
	// read json
	fileHeader, err := ginCtx.FormFile("json")
	if err != nil {
		log.Errorf("could not get json payload: %v", err)
		ginCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		log.Errorf("could not open json payload: %v", err)
		ginCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	jsonBites, err := io.ReadAll(file)
	if err != nil {
		log.Errorf("could not read json payload: %v", err)
		ginCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	req := &FilmRequest{}
	err = json.Unmarshal(jsonBites, &req)
	if err != nil {
		log.Errorf("could not map json to FilmRequest: %v", err)
		ginCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// read image
	fileHeader, err = ginCtx.FormFile("image")
	if err != nil {
		log.Errorf("could not get uploaded image: %v", err)
		ginCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	file, err = fileHeader.Open()
	if err != nil {
		log.Errorf("could not open uploaded image: %v", err)
		ginCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	imageBites, err := io.ReadAll(file)
	if err != nil {
		log.Errorf("could not read uploaded image: %v", err)
		ginCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// create film
	filmkritikenDetails := &filmkritiken.FilmkritikenDetails{
		BeitragVon:     req.Von,
		BesprochenAm:   req.BesprochenAm,
		BewertungOffen: req.BewertungOffen,
	}
	result, err := h.filmkritikenService.CreateFilm(ginCtx.Request.Context(), req.Film, filmkritikenDetails, &imageBites)
	if err != nil {
		log.Errorf("could not create film: %v", err)
		ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		_, _ = ginCtx.Writer.WriteString(err.Error())
		return
	}

	ginCtx.JSON(http.StatusCreated, result)
}

func (h *filmkritikenHandler) handleOpenCloseBewertungen(ginCtx *gin.Context) {

	filmkritikenId := ginCtx.Param("filmkritikenId")
	if filmkritikenId == "" {
		ginCtx.Writer.WriteHeader(http.StatusBadRequest)
		_, _ = ginCtx.Writer.WriteString("Film muss angegeben werden")
		return
	}

	offenStr := ginCtx.Param("offen")
	if offenStr == "" {
		ginCtx.Writer.WriteHeader(http.StatusBadRequest)
		_, _ = ginCtx.Writer.WriteString("Neuer Zustand muss angegeben werden")
		return
	}

	if offenStr != "true" && offenStr != "false" {
		ginCtx.Writer.WriteHeader(http.StatusBadRequest)
		_, _ = ginCtx.Writer.WriteString("Neuer Zustand muss angegeben werden")
		return
	}

	offen, err := strconv.ParseBool(offenStr)
	if err != nil {
		log.Errorf("could not open / close bewertungen: %v", err)
		ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		_, _ = ginCtx.Writer.WriteString(err.Error())
	}

	err = h.filmkritikenService.OpenCloseBewertungen(ginCtx.Request.Context(), filmkritikenId, offen)

	if err != nil {
		if _, ok := err.(*filmkritiken.NotFoundError); ok {
			log.Warnf("could not find filmkritiken (%s): %v", filmkritikenId, err)
			ginCtx.Writer.WriteHeader(http.StatusNotFound)
			_, _ = ginCtx.Writer.WriteString(err.Error())
			return
		}
		log.Errorf("could not open / close bewertungen: %v", err)
		ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		_, _ = ginCtx.Writer.WriteString(err.Error())
		return
	}

	ginCtx.Writer.WriteHeader(http.StatusNoContent)
}

func (h *filmkritikenHandler) handleSetBewertung(ginCtx *gin.Context) {
	req := &SetBewertungRequest{}
	err := ginCtx.ShouldBindJSON(req)
	if err != nil {
		log.Error("could not map json to SetBewertungRequest")
		ginCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	filmkritikenId := ginCtx.Param("filmkritikenId")
	if filmkritikenId == "" {
		ginCtx.Writer.WriteHeader(http.StatusBadRequest)
		_, _ = ginCtx.Writer.WriteString("Film muss angegeben werden")
		return
	}

	usernameFromUrl := ginCtx.Param("username")
	requestContext := ginCtx.Request.Context()

	username := requestContext.Value(filmkritiken.Context_Username).(string)

	if usernameFromUrl != username {
		log.Warnf("users in URL (%s) and token (%s) do not match", usernameFromUrl, username)
		ginCtx.Writer.WriteHeader(http.StatusBadRequest)
		_, _ = ginCtx.Writer.WriteString("Benutzer muss mit eingeloggtem Benutzer übereinstimmen")
		return
	}

	err = h.filmkritikenService.SetKritik(requestContext, req.FilmkritikenId, username, req.Wertung)

	if err != nil {
		if _, ok := err.(*filmkritiken.InvalidInputError); ok {
			ginCtx.Writer.WriteHeader(http.StatusBadRequest)
			_, _ = ginCtx.Writer.WriteString(err.Error())
			return
		}
		if _, ok := err.(*filmkritiken.NotFoundError); ok {
			log.Warnf("could not find filmkritiken (%s): %v", req.FilmkritikenId, err)
			ginCtx.Writer.WriteHeader(http.StatusNotFound)
			_, _ = ginCtx.Writer.WriteString(err.Error())
			return
		}
		log.Errorf("could not set bewertung: %v", err)
		ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		_, _ = ginCtx.Writer.WriteString(err.Error())
		return
	}

	ginCtx.Writer.WriteHeader(http.StatusNoContent)
}

func (h *filmkritikenHandler) loadImage(ginCtx *gin.Context) {
	imageId := ginCtx.Param("imageId")
	if imageId == "" {
		ginCtx.Writer.WriteHeader(http.StatusBadRequest)
		_, _ = ginCtx.Writer.WriteString("Bild muss angegeben werden")
		return
	}

	imageBites, err := h.filmkritikenService.LoadImage(ginCtx.Request.Context(), imageId)
	if err != nil {
		if _, ok := err.(*filmkritiken.NotFoundError); ok {
			log.Warnf("could not find image (%s): %v", imageId, err)
			ginCtx.Writer.WriteHeader(http.StatusNotFound)
			_, _ = ginCtx.Writer.WriteString("Bild konnte nicht gefunden werden")
			return
		}
		log.Errorf("could not get image: %v", err)
		ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		_, _ = ginCtx.Writer.WriteString(err.Error())
		return
	}

	ginCtx.Writer.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%v, immutable", imageCacheDuration))
	ginCtx.Writer.WriteHeader(http.StatusOK)
	_, _ = ginCtx.Writer.Write(*imageBites)

}

func (h *filmkritikenHandler) handleSetBesprochenAm(ginCtx *gin.Context) {
	filmkritikenId := ginCtx.Param("filmkritikenId")
	if filmkritikenId == "" {
		ginCtx.Writer.WriteHeader(http.StatusBadRequest)
		ginCtx.Writer.WriteString("Film muss angegeben werden")
		return
	}

	req := &SetBesprochenAmRequest{}
	err := ginCtx.ShouldBindJSON(req)
	if err != nil {
		log.Errorf("could not map json to SetBesprochenAmRequest: %v", err)
		ginCtx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err = h.filmkritikenService.UpdateBesprochenAm(ginCtx.Request.Context(), filmkritikenId, req.BesprochenAm)
	if err != nil {
		if _, ok := err.(*filmkritiken.NotFoundError); ok {
			log.Warnf("could not find filmkritiken (%s): %v", filmkritikenId, err)
			ginCtx.Writer.WriteHeader(http.StatusNotFound)
			ginCtx.Writer.WriteString(err.Error())
			return
		}
		log.Errorf("could not update besprochenAm: %v", err)
		ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		ginCtx.Writer.WriteString(err.Error())
		return
	}

	ginCtx.Writer.WriteHeader(http.StatusNoContent)
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
