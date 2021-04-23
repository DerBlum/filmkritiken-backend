package inbound

import (
	gin "github.com/gin-gonic/gin"
)

func StartServer() {
	r := gin.Default()
	r.Run()
}
