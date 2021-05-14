package main

import (
	httpInbound "github.com/DerBlum/filmkritiken-backend/http/inbound"
	log "github.com/sirupsen/logrus"
)

func main() {
	//log.SetLevel(log.DebugLevel)
	log.Info("Starting App")

	httpInbound.StartServer()

}
