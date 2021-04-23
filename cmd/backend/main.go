package main

import (
	httpInbound "github.com/DerBlum/filmkritiken-backend/http/inbound"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("App running")

	httpInbound.StartServer()

}
