package main

import (
	"log"
	"os"
	"strconv"

	"github.com/animalpride/animalpride-core/services/denops-auth/internal/config"
	"github.com/animalpride/animalpride-core/services/denops-auth/internal/db"
	"github.com/animalpride/animalpride-core/services/denops-auth/internal/routes"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load config
	cfg, err := config.LoadConfig("config.yml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize database
	db, err := db.NewDB(cfg)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	log.Printf("Connected to database: %s", cfg.Database.Host)
	// Initialize and start the server
	router := routes.SetupRouter(cfg, db)
	address := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port)
	if err := router.Run(address); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	log.Printf("Server started on %s", address)
}
