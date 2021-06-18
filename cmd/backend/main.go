package main

import (
	"context"

	"github.com/DerBlum/filmkritiken-backend/domain/filmkritiken"
	httpInbound "github.com/DerBlum/filmkritiken-backend/http/inbound"
	"github.com/DerBlum/filmkritiken-backend/infrastructure/db/mongo"
	"github.com/caarlos0/env/v6"
	log "github.com/sirupsen/logrus"
)

type LogConfig struct {
	LogLevel string `env:"LOG_LEVEL" envDefault:"INFO"`
}

func main() {
	log.SetLevel(getLogLevel())
	log.Info("starting filmkritiken-backend")

	mongoConfig := mongo.MongoDbConfig{}
	if err := env.Parse(&mongoConfig); err != nil {
		panic(err)
	}

	serverConfig := httpInbound.ServerConfig{}
	if err := env.Parse(&serverConfig); err != nil {
		panic(err)
	}

	mongoDbRepository, err := mongo.NewMongoDbRepository(context.Background(), &mongoConfig)
	if err != nil {
		panic(err)
	}
	filmkritikenService := filmkritiken.NewFilmkritikenService(mongoDbRepository, mongoDbRepository)

	httpInbound.StartServer(&serverConfig, filmkritikenService)

}

func getLogLevel() log.Level {
	logConfig := LogConfig{}
	if err := env.Parse(&logConfig); err != nil {
		log.Warnf("could not parse LogLevel: %s, using INFO", logConfig.LogLevel)
		return log.InfoLevel
	} else {
		log.Infof("using log level %v", logConfig.LogLevel)
	}

	switch logConfig.LogLevel {
	case "TRACE":
		return log.TraceLevel
	case "DEBUG":
		return log.DebugLevel
	case "INFO":
		return log.InfoLevel
	case "WARNING":
		return log.WarnLevel
	case "FATAL":
		return log.FatalLevel
	case "PANIC":
		return log.PanicLevel
	default:
		return log.InfoLevel
	}

}
