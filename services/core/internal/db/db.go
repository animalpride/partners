package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/animalpride/animalpride-core/services/core/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
	)

	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := database.DB()
	if err != nil {
		return nil, err
	}
	configurePool(sqlDB, cfg.Database.Pool)

	return database, nil
}

func configurePool(sqlDB *sql.DB, pool config.DatabasePool) {
	maxOpen := pool.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 10
	}
	maxIdle := pool.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 5
	}
	if maxIdle > maxOpen {
		maxIdle = maxOpen
	}
	maxLifetimeMinutes := pool.ConnMaxLifetimeMinutes
	if maxLifetimeMinutes <= 0 {
		maxLifetimeMinutes = 30
	}
	maxIdleMinutes := pool.ConnMaxIdleMinutes
	if maxIdleMinutes <= 0 {
		maxIdleMinutes = 5
	}

	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(time.Duration(maxLifetimeMinutes) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(maxIdleMinutes) * time.Minute)
}
