package config

// DatabasePool holds connection pool settings for a GORM database connection.
// Embed this struct inside each service's Database config struct:
//
//	type Database struct {
//	    ...
//	    Pool sharedconfig.DatabasePool `yaml:"pool"`
//	}
type DatabasePool struct {
	MaxOpenConns           int `yaml:"max_open_conns"`
	MaxIdleConns           int `yaml:"max_idle_conns"`
	ConnMaxLifetimeMinutes int `yaml:"conn_max_lifetime_minutes"`
	ConnMaxIdleMinutes     int `yaml:"conn_max_idle_minutes"`
}
