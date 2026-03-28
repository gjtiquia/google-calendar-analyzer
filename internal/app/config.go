package app

import "os"

type Config struct {
	Env  string
	Addr string
}

func LoadConfigFromEnv() Config {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	addr := os.Getenv("APP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	return Config{Env: env, Addr: addr}
}
