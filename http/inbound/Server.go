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

	r := gin.Default()
	r.GET("/filmkritiken", filmkritikenHandler.handleGetFilmkritiken)
	r.POST("/film", filmkritikenHandler.handleCreateFilm)
	err = r.Run()

	if err != nil {
		return err
	}
	return nil
}
