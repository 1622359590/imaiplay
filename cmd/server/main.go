package main

import (
	"log"

	"github.com/1622359590/imaiplay/internal/config"
	"github.com/1622359590/imaiplay/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}
	if err := server.Run(cfg); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
