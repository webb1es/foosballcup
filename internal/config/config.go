package config

import (
	"fmt"
	"strconv"
)

const (
	defaultPort         = 8080
	defaultDatabasePath = "foosballcup.db"
)

type Config struct {
	Port         int
	DatabasePath string
}

func Load(getenv func(string) string) (Config, error) {
	port := defaultPort
	if raw := getenv("PORT"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 65535 {
			return Config{}, fmt.Errorf("PORT must be a number from 1 to 65535, got %q", raw)
		}
		port = parsed
	}
	dbPath := getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = defaultDatabasePath
	}
	return Config{Port: port, DatabasePath: dbPath}, nil
}

func (c Config) Addr() string {
	return ":" + strconv.Itoa(c.Port)
}
