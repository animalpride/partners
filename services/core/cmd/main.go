package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/animalpride/partners/services/core/internal/config"
	"github.com/animalpride/partners/services/core/internal/db"
	"github.com/animalpride/partners/services/core/internal/repository"
	"github.com/animalpride/partners/services/core/internal/routes"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var configPath string
	var importOnStartup bool
	var importOnly bool
	var combinedURL string
	var countriesURL string
	var statesURL string
	var citiesURL string
	var importTimeout time.Duration

	flag.StringVar(&configPath, "config", envOrDefault("CORE_CONFIG_PATH", "config.yml"), "path to core config file")
	flag.BoolVar(&importOnStartup, "import-locations-on-start", envBoolOrDefault("CORE_IMPORT_ON_STARTUP", false), "refresh location lookup data before starting server")
	flag.BoolVar(&importOnly, "import-locations-only", envBoolOrDefault("CORE_IMPORT_ON_STARTUP_ONLY", false), "run location refresh and exit")
	flag.StringVar(&combinedURL, "combined-url", os.Getenv("CORE_LOCATIONS_COMBINED_URL"), "combined countries/states/cities dataset URL")
	flag.StringVar(&countriesURL, "countries-url", os.Getenv("CORE_LOCATIONS_COUNTRIES_URL"), "countries dataset URL")
	flag.StringVar(&statesURL, "states-url", os.Getenv("CORE_LOCATIONS_STATES_URL"), "states dataset URL")
	flag.StringVar(&citiesURL, "cities-url", os.Getenv("CORE_LOCATIONS_CITIES_URL"), "cities dataset URL")
	flag.DurationVar(&importTimeout, "import-timeout", envDurationOrDefault("CORE_IMPORT_TIMEOUT", 10*time.Minute), "location import timeout")
	flag.Parse()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	database, err := db.NewDB(cfg)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	if importOnStartup || importOnly {
		repo := repository.NewLocationRepository(database)
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		importCtx, cancel := context.WithTimeout(ctx, importTimeout)
		defer cancel()

		imported, err := repo.RefreshFromGitHub(importCtx, combinedURL, countriesURL, statesURL, citiesURL, 0)
		if err != nil {
			log.Fatalf("location refresh failed: %v", err)
		}
		if imported {
			log.Printf("location refresh completed")
		} else {
			log.Printf("location refresh skipped (another instance is already refreshing)")
		}

		if importOnly {
			log.Printf("import-only mode complete; exiting")
			return
		}
	}

	router := routes.SetupRouter(cfg, database)
	address := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port)
	if err := router.Run(address); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envBoolOrDefault(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDurationOrDefault(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
