package inbound

import (
	"github.com/DerBlum/filmkritiken-backend/domain/filmkritiken"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type ServerConfig struct {
	CorsAllowOrigins        []string `env:"CORS_ALLOW_ORIGINS" envDefault:"https://filmkritiken-frontend.marsrover.418-teapot.de,https://filmkritiken.marsrover.418-teapot.de"`
	MetricsEndpointUser     string   `env:"METRICS_ENDPOINT_USER"`
	MetricsEndpointPassword string   `env:"METRICS_ENDPOINT_PASSWORD"`
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

	ginOmitLogConfig := gin.LoggerConfig{SkipPaths: []string{"/"}}
	ginOmitLogMiddleware := gin.LoggerWithConfig(ginOmitLogConfig)
	r.GET("/", ginOmitLogMiddleware, healthcheckHandler)
	metricsAuthHandler := NewEmptyHandler()
	if serverConfig.MetricsEndpointUser != "" && serverConfig.MetricsEndpointPassword != "" {
		metricsAuthHandler = NewBasicAuthHandler(serverConfig.MetricsEndpointUser, serverConfig.MetricsEndpointPassword)
	}
	r.GET("/metrics", ginOmitLogMiddleware, metricsAuthHandler, gin.WrapH(promhttp.Handler()))

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

func healthcheckHandler(context *gin.Context) {
	log.Trace("healthcheck called")
	context.Status(200)
}
