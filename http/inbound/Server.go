package inbound

import (
	"github.com/DerBlum/filmkritiken-backend/domain/filmkritiken"
	"github.com/gin-contrib/cors"
	gin "github.com/gin-gonic/gin"
)

func StartServer(filmkritikenService filmkritiken.FilmkritikenService) error {
	filmkritikenHandler := NewFilmkritikenHandler(filmkritikenService)

	handlers := []gin.HandlerFunc{
		TraceIdMiddleware,
	}

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		//AllowOrigins:     []string{"https://filmkritiken-frontend.marsrover.418-teapot.de"},
		AllowOrigins:     []string{"https://filmkritiken-frontend.marsrover.418-teapot.de", "http://localhost:4200"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Content-Length", "Accept-Encoding", "Authorization", "origin", "Cache-Control"},
		AllowCredentials: true,
	}))
	api := r.Group("/api", handlers...)

	api.GET("/filmkritiken", filmkritikenHandler.handleGetFilmkritiken)
	api.POST("/filme",
		NewAuthHandler([]string{"film.add"}),
		filmkritikenHandler.handleCreateFilm)
	api.PUT("/filmkritiken/:filmkritikenId/bewertungen/:username",
		NewAuthHandler([]string{"bewertung.add"}),
		filmkritikenHandler.handleSetBewertung)
	err := r.Run()

	if err != nil {
		return err
	}
	return nil
}
