package main

import (
	"log"
	"os"
	"strconv"

	"github.com/animalpride/animalpride-core/services/core/internal/config"
	"github.com/animalpride/animalpride-core/services/core/internal/db"
	"github.com/animalpride/animalpride-core/services/core/internal/routes"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg, err := config.LoadConfig("config.yml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	database, err := db.NewDB(cfg)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	router := routes.SetupRouter(cfg, database)
	address := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port)
	if err := router.Run(address); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
