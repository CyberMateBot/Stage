package config

import "testing"

func TestPostgresFromDatabaseURL(t *testing.T) {
	raw := "postgresql://user:secret@db.example.com:6543/mydb?sslmode=require"
	cfg, ok := postgresFromDatabaseURL(raw)
	if !ok {
		t.Fatal("expected ok")
	}
	if cfg.Host != "db.example.com" || cfg.Port != "6543" || cfg.User != "user" || cfg.Pass != "secret" || cfg.DBName != "mydb" || cfg.SSLMode != "require" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}

func TestLoadPostgresConfig_RailwayEnvNames(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("PG_HOST", "")
	t.Setenv("PGHOST", "postgres.railway.internal")
	t.Setenv("PGPORT", "5432")
	t.Setenv("PGUSER", "railway")
	t.Setenv("PGPASSWORD", "pw")
	t.Setenv("PGDATABASE", "railway")
	t.Setenv("PG_SSLMODE", "require")

	cfg := LoadPostgresConfig()
	if cfg.Host != "postgres.railway.internal" || cfg.User != "railway" || cfg.Pass != "pw" || cfg.DBName != "railway" || cfg.SSLMode != "require" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}

func TestLoadPostgresConfig_DatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@host:5432/db?sslmode=require")
	t.Setenv("PGHOST", "other-host")

	cfg := LoadPostgresConfig()
	if cfg.Host != "host" || cfg.DBName != "db" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}
