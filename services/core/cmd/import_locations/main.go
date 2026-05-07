package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/animalpride/partners/services/core/internal/config"
	"github.com/animalpride/partners/services/core/internal/db"
	"github.com/animalpride/partners/services/core/internal/repository"
)

func main() {
	var combinedURL string
	var countriesURL string
	var statesURL string
	var citiesURL string
	var configPath string

	flag.StringVar(&configPath, "config", "config.yml", "path to core config file")
	flag.StringVar(&combinedURL, "combined-url", "", "combined countries/states/cities dataset URL")
	flag.StringVar(&countriesURL, "countries-url", "", "countries dataset URL")
	flag.StringVar(&statesURL, "states-url", "", "states dataset URL")
	flag.StringVar(&citiesURL, "cities-url", "", "cities dataset URL")
	flag.Parse()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	database, err := db.NewDB(cfg)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	repo := repository.NewLocationRepository(database)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	imported, err := repo.RefreshFromGitHub(ctx, combinedURL, countriesURL, statesURL, citiesURL, 120)
	if err != nil {
		log.Fatalf("location import failed: %v", err)
	}
	if !imported {
		log.Fatalf("location import skipped because another refresh is in progress")
	}

	log.Println("location import completed")
}
