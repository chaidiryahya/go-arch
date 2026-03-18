package db

import (
	"database/sql"
	"fmt"

	"go-arch/internal/infra/configuration"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

type Registry struct {
	connections map[string]*sql.DB
}

func NewRegistry(configs map[string]configuration.DatabaseConfig) (*Registry, error) {
	r := &Registry{
		connections: make(map[string]*sql.DB),
	}

	for name, cfg := range configs {
		dsn, err := buildDSN(cfg)
		if err != nil {
			r.Close()
			return nil, fmt.Errorf("building DSN for %q: %w", name, err)
		}

		conn, err := sql.Open(cfg.Driver, dsn)
		if err != nil {
			r.Close()
			return nil, fmt.Errorf("opening connection %q: %w", name, err)
		}

		conn.SetMaxOpenConns(cfg.MaxOpenConns)
		conn.SetMaxIdleConns(cfg.MaxIdleConns)
		conn.SetConnMaxLifetime(cfg.ConnMaxLifetime)

		if err := conn.Ping(); err != nil {
			r.Close()
			return nil, fmt.Errorf("pinging connection %q: %w", name, err)
		}

		r.connections[name] = conn
	}

	return r, nil
}

func (r *Registry) Get(name string) (*sql.DB, error) {
	conn, ok := r.connections[name]
	if !ok {
		return nil, fmt.Errorf("database connection %q not found", name)
	}
	return conn, nil
}

func (r *Registry) Close() error {
	var firstErr error
	for name, conn := range r.connections {
		if err := conn.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("closing connection %q: %w", name, err)
		}
	}
	return firstErr
}

func buildDSN(cfg configuration.DatabaseConfig) (string, error) {
	switch cfg.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName), nil
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName), nil
	default:
		return "", fmt.Errorf("unsupported driver: %s", cfg.Driver)
	}
}
