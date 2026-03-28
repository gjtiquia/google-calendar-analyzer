package app

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Env                string
	Addr               string
	BaseURL            string
	SessionSecret      []byte
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	SessionCookieName string
	SessionMaxAge     int
}

func LoadConfigFromEnv() (Config, error) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	addr := os.Getenv("APP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	baseURL := strings.TrimRight(os.Getenv("APP_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = "http://127.0.0.1" + addr
	}

	secret, err := parseSessionSecret(os.Getenv("SESSION_SECRET"))
	if err != nil {
		return Config{}, err
	}

	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")

	cookieName := os.Getenv("SESSION_COOKIE_NAME")
	if cookieName == "" {
		cookieName = "gca_session"
	}

	maxAge := 3600
	if v := os.Getenv("SESSION_MAX_AGE_SECONDS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return Config{}, fmt.Errorf("SESSION_MAX_AGE_SECONDS: invalid value %q", v)
		}
		maxAge = n
	}

	return Config{
		Env:                env,
		Addr:               addr,
		BaseURL:            baseURL,
		SessionSecret:      secret,
		GoogleClientID:     clientID,
		GoogleClientSecret: clientSecret,
		GoogleRedirectURL:  redirectURL,
		SessionCookieName: cookieName,
		SessionMaxAge:     maxAge,
	}, nil
}

func (c Config) OAuthConfigured() bool {
	return c.GoogleClientID != "" && c.GoogleClientSecret != "" && c.GoogleRedirectURL != ""
}

func parseSessionSecret(s string) ([]byte, error) {
	if s == "" {
		return nil, errors.New("SESSION_SECRET is required")
	}
	if b, err := base64.StdEncoding.DecodeString(s); err == nil && len(b) >= 32 {
		return b, nil
	}
	if len(s) >= 32 {
		return []byte(s), nil
	}
	return nil, errors.New("SESSION_SECRET must decode to at least 32 bytes (or use a raw string of 32+ bytes)")
}
