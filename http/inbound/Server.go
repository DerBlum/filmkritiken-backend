package inbound

import (
	"fmt"
	"github.com/DerBlum/filmkritiken-backend/domain/filmkritiken"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"time"
)

type ServerConfig struct {
	CorsAllowOrigins        []string `env:"CORS_ALLOW_ORIGINS" envDefault:"https://filmkritiken.marsrover.418-teapot.de"`
	MetricsEndpointUser     string   `env:"METRICS_ENDPOINT_USER"`
	MetricsEndpointPassword string   `env:"METRICS_ENDPOINT_PASSWORD"`
}

var inFlightGauge prometheus.Gauge
var requestCounter *prometheus.CounterVec
var durationHistogram *prometheus.HistogramVec

func init() {
	initPrometheusMetrics()
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
				AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "OPTIONS"},
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
		metricsAuthHandler = gin.BasicAuth(gin.Accounts{
			serverConfig.MetricsEndpointUser: serverConfig.MetricsEndpointPassword,
		})
	}
	r.GET("/metrics", ginOmitLogMiddleware, metricsAuthHandler, gin.WrapH(promhttp.Handler()))

	api := r.Group("/api", handlers...)
	api.GET("/filmkritiken", metricsHandlerWrapper(filmkritikenHandler.handleGetFilmkritiken, "getFilmkritiken"))
	api.GET("/images/:imageId", metricsHandlerWrapper(filmkritikenHandler.loadImage, "loadImage"))
	api.POST(
		"/filme",
		NewAuthHandler([]string{"film.add"}),
		metricsHandlerWrapper(filmkritikenHandler.handleCreateFilm, "createFilm"),
	)
	api.PUT(
		"/filmkritiken/:filmkritikenId/bewertungen/:username",
		NewAuthHandler([]string{"bewertung.add"}),
		metricsHandlerWrapper(filmkritikenHandler.handleSetBewertung, "setBewertung"),
	)
	api.PATCH(
		"/filmkritiken/:filmkritikenId/bewertungenoffen/:offen",
		NewAuthHandler([]string{"bewertung.openclose"}),
		metricsHandlerWrapper(filmkritikenHandler.handleOpenCloseBewertungen, "openCloseBewertungen"),
	)
	api.PATCH(
		"/filmkritiken/:filmkritikenId/besprochenAm",
		NewAuthHandler([]string{"film.add"}),
		filmkritikenHandler.handleSetBesprochenAm,
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

func initPrometheusMetrics() {
	inFlightGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "in_flight_requests",
		Help: "A gauge of requests currently being served by the wrapped handler.",
	})

	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "A counter for requests to the wrapped handler.",
		},
		[]string{"handler", "code", "method"},
	)

	// duration is partitioned by the HTTP method and handler. It uses custom
	// buckets based on the expected request duration.
	durationHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "A histogram of latencies for requests.",
			Buckets: []float64{0.025, 0.05, 0.1, 0.25, 0.500, 1, 2.5},
		},
		[]string{"handler", "method"},
	)

	prometheus.MustRegister(inFlightGauge, requestCounter, durationHistogram)
}

func metricsHandlerWrapper(handler func(*gin.Context), handlerName string) func(*gin.Context) {
	return func(c *gin.Context) {

		// inflight request gauge
		inFlightGauge.Inc()
		defer inFlightGauge.Dec()

		start := time.Now()

		handler(c)

		// prepare values for labels
		method := c.Request.Method
		statusCode := fmt.Sprintf("%d", c.Writer.Status())

		// measure duration for non-panic requests
		durationHistogram.
			WithLabelValues(handlerName, method).
			Observe(time.Since(start).Seconds())

		// measure amount of non-panic requests
		requestCounter.
			WithLabelValues(handlerName, statusCode, c.Request.Method).
			Inc()

	}
}
