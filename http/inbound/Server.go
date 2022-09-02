package inbound

import (
	"github.com/DerBlum/filmkritiken-backend/domain/filmkritiken"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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
	r.MaxMultipartMemory = 8 << 20 // max 8 MB file size
	r.Use(
		cors.New(
			cors.Config{
				AllowOrigins:     serverConfig.CorsAllowOrigins,
				AllowMethods:     []string{"GET", "POST", "PUT", "PATCH"},
				AllowHeaders:     []string{"content-type", "Content-Length", "Accept-Encoding", "Authorization", "origin", "Cache-Control"},
				AllowCredentials: true,
			},
		),
	)
	api := r.Group("/api", handlers...)

	api.GET("/filmkritiken", filmkritikenHandler.handleGetFilmkritiken)
	api.GET("/images/:imageId", filmkritikenHandler.loadImage)
	api.POST(
		"/filme",
		NewAuthHandler([]string{"film.add"}),
		filmkritikenHandler.handleCreateFilm,
	)
	api.PUT(
		"/filmkritiken/:filmkritikenId/bewertungen/:username",
		NewAuthHandler([]string{"bewertung.add"}),
		filmkritikenHandler.handleSetBewertung,
	)
	api.PATCH(
		"/filmkritiken/:filmkritikenId/bewertungenoffen/:offen",
		NewAuthHandler([]string{"bewertung.openclose"}),
		filmkritikenHandler.handleOpenCloseBewertungen,
	)
	err := r.Run()

	if err != nil {
		return err
	}
	return nil
}
