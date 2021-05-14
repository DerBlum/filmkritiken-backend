package inbound

import (
	"context"

	"github.com/DerBlum/filmkritiken-backend/domain/filmkritiken"
	"github.com/DerBlum/filmkritiken-backend/infrastructure/db/mongo"
	gin "github.com/gin-gonic/gin"
)

func StartServer() error {

	mongoDbRepository, err := mongo.NewMongoDbRepository(context.Background())
	if err != nil {
		panic(err)
	}
	filmkritikenService := filmkritiken.NewFilmkritikenService(mongoDbRepository)
	filmkritikenHandler := NewFilmkritikenHandler(filmkritikenService)

	handlers := []gin.HandlerFunc{
		TraceIdMiddleware,
	}

	r := gin.Default()
	api := r.Group("/api", handlers...)

	api.GET("/filmkritiken", filmkritikenHandler.handleGetFilmkritiken)
	api.POST("/film", NewAuthHandler([]string{"film.add"}), filmkritikenHandler.handleCreateFilm)
	err = r.Run()

	if err != nil {
		return err
	}
	return nil
}
