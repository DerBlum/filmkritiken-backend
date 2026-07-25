package main

import (
	"context"

	"github.com/DerBlum/filmkritiken-backend/infrastructure/db/mongo"
	"github.com/caarlos0/env/v11"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("Starting database seeder...")

	mongoConfig := mongo.Config{}
	if err := env.Parse(&mongoConfig); err != nil {
		panic(err)
	}

	mongoDbRepository, err := mongo.NewMongoDbRepository(context.Background(), &mongoConfig)
	if err != nil {
		panic(err)
	}

	if err := seedIfEmpty(context.Background(), mongoDbRepository); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	log.Info("Database seeder finished.")
}
