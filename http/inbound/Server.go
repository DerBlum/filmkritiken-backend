package inbound

import (
	"github.com/DerBlum/filmkritiken-backend/domain/filmkritiken"
	gin "github.com/gin-gonic/gin"
)

func StartServer() error {

	filmkritikenService := filmkritiken.NewFilmkritikenService()
	filmkritikenHandler := NewFilmkritikenHandler(filmkritikenService)

	r := gin.Default()
	r.GET("/filmkritiken", filmkritikenHandler.handleGetFilmkritiken)
	err := r.Run()

	if err != nil {
		return err
	}
	return nil
}
