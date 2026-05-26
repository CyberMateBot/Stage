package config

import (
	"net/url"
	"os"
	"strings"
	"time"
)

// getenvFirst returns the first non-empty environment variable from keys.
func getenvFirst(keys ...string) string {
	for _, key := range keys {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	return ""
}

// postgresFromDatabaseURL parses Railway/Heroku-style DATABASE_URL into ConfigPostgres.
func postgresFromDatabaseURL(raw string) (ConfigPostgres, bool) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return ConfigPostgres{}, false
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "postgres" && scheme != "postgresql" {
		return ConfigPostgres{}, false
	}

	cfg := ConfigPostgres{
		SSLMode: "disable",
	}

	if u.User != nil {
		cfg.User = u.User.Username()
		cfg.Pass, _ = u.User.Password()
	}

	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "5432"
	}
	cfg.Host = host
	cfg.Port = port

	dbName := strings.TrimPrefix(u.Path, "/")
	if dbName != "" {
		cfg.DBName = dbName
	}

	if sslMode := u.Query().Get("sslmode"); sslMode != "" {
		cfg.SSLMode = sslMode
	}

	return cfg, true
}

func mergePostgresFromEnv(cfg ConfigPostgres) ConfigPostgres {
	if cfg.Host == "" {
		cfg.Host = getenvFirst("PG_HOST", "PGHOST")
	}
	if cfg.Port == "" {
		cfg.Port = getenvFirst("PG_PORT", "PGPORT")
	}
	if cfg.User == "" {
		cfg.User = getenvFirst("PG_USER", "PGUSER")
	}
	if cfg.Pass == "" {
		cfg.Pass = getenvFirst("PG_PASS", "PG_PASSWORD", "PGPASSWORD")
	}
	if cfg.DBName == "" {
		cfg.DBName = getenvFirst("PG_DBNAME", "PGDATABASE")
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = os.Getenv("PG_SSLMODE")
	}
	if cfg.SSLRootCert == "" {
		cfg.SSLRootCert = os.Getenv("PG_SSLROOTCERT")
	}
	return cfg
}

func applyPostgresDefaults(cfg ConfigPostgres) ConfigPostgres {
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == "" {
		cfg.Port = "5432"
	}
	if cfg.User == "" {
		cfg.User = "postgres"
	}
	if cfg.Pass == "" {
		cfg.Pass = "postgres"
	}
	if cfg.DBName == "" {
		cfg.DBName = "postgres"
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}
	if cfg.DriverLogLevel == "" {
		cfg.DriverLogLevel = "info"
	}
	if cfg.PoolStatPeriod == 0 {
		cfg.PoolStatPeriod = 30 * time.Second
	}
	if cfg.PoolMaxConns == 0 {
		cfg.PoolMaxConns = 10
	}
	if cfg.PoolMinConns == 0 {
		cfg.PoolMinConns = 1
	}
	if cfg.PoolMaxConnLifeTime == 0 {
		cfg.PoolMaxConnLifeTime = time.Hour
	}
	if cfg.PoolMaxConnIdleTime == 0 {
		cfg.PoolMaxConnIdleTime = 30 * time.Minute
	}
	if cfg.PoolHealthCheckPeriod == 0 {
		cfg.PoolHealthCheckPeriod = time.Minute
	}
	return cfg
}
