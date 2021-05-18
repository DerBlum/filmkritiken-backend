package inbound

import (
	"github.com/DerBlum/filmkritiken-backend/domain/filmkritiken"
	gin "github.com/gin-gonic/gin"
)

func StartServer(filmkritikenService filmkritiken.FilmkritikenService) error {
	filmkritikenHandler := NewFilmkritikenHandler(filmkritikenService)

	handlers := []gin.HandlerFunc{
		TraceIdMiddleware,
	}

	r := gin.Default()
	api := r.Group("/api", handlers...)

	api.GET("/filmkritiken", filmkritikenHandler.handleGetFilmkritiken)
	api.POST("/film", NewAuthHandler([]string{"film.add"}), filmkritikenHandler.handleCreateFilm)
	err := r.Run()

	if err != nil {
		return err
	}
	return nil
}
