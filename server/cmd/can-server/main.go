package main

import (
	"log"

	"github.com/user/can-server/config"
	"github.com/user/can-server/internal/api"
	"github.com/user/can-server/internal/db"
)

func main() {
	cfg := config.Load()

	if err := db.Init(cfg.Database); err != nil {
		log.Fatalf("database init failed: %v", err)
	}

	srv := api.NewServer(cfg)
	log.Printf("starting can-server on :%d", cfg.Server.Port)
	if err := srv.Start(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
