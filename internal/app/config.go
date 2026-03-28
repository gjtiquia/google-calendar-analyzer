package app

import "os"

type Config struct {
	Addr string
}

func LoadConfigFromEnv() Config {
	addr := os.Getenv("APP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	return Config{Addr: addr}
}
