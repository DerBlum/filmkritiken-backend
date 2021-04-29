package inbound

import (
	"github.com/DerBlum/filmkritiken-backend/domain/filmkritiken"
	gin "github.com/gin-gonic/gin"
)

func StartServer() {

	filmkritikenService := filmkritiken.NewFilmkritikenService()
	filmkritikenHandler := NewFilmkritikenHandler(filmkritikenService)

	r := gin.Default()
	r.GET("/filmkritiken", filmkritikenHandler.handleGetFilmkritiken)
	r.Run()
}
