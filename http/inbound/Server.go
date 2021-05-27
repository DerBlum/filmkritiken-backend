package inbound

import (
	"github.com/DerBlum/filmkritiken-backend/domain/filmkritiken"
	"github.com/gin-contrib/cors"
	gin "github.com/gin-gonic/gin"
)

type ServerConfig struct {
	CorsAllowOrigins []string `env:"CORS_ALLOW_ORIGINS" envDefault:"https://filmkritiken-frontend.marsrover.418-teapot.de"`
}

func StartServer(serverConfig *ServerConfig, filmkritikenService filmkritiken.FilmkritikenService) error {
	filmkritikenHandler := NewFilmkritikenHandler(filmkritikenService)

	handlers := []gin.HandlerFunc{
		TraceIdMiddleware,
	}

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     serverConfig.CorsAllowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT"},
		AllowHeaders:     []string{"content-type", "Content-Length", "Accept-Encoding", "Authorization", "origin", "Cache-Control"},
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
