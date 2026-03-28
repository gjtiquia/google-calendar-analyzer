package session

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/hkdf"
)

const oauthStateCookie = "gca_oauth_state"

type Payload struct {
	Sub                  string `json:"sub"`
	Email                string `json:"email"`
	AccessToken          string `json:"access_token"`
	AccessTokenExpiryUTC int64  `json:"access_token_expiry"`
}

func (p *Payload) Expired(now time.Time) bool {
	if p == nil {
		return true
	}
	if p.AccessTokenExpiryUTC <= 0 {
		return true
	}
	return now.Unix() >= p.AccessTokenExpiryUTC
}

type Manager struct {
	cfg     Config
	codec   *securecookie.SecureCookie
	stCodec *securecookie.SecureCookie
}

type Config struct {
	CookieName    string
	SessionSecret []byte
	MaxAgeSeconds int
	SecureCookies bool
}

func NewManager(c Config) *Manager {
	hashKey, blockKey := deriveSessionKeys(c.SessionSecret)
	stHash, stBlock := deriveOAuthStateKeys(c.SessionSecret)
	return &Manager{
		cfg:     c,
		codec:   securecookie.New(hashKey, blockKey),
		stCodec: securecookie.New(stHash, stBlock),
	}
}

func deriveSessionKeys(secret []byte) (hashKey, blockKey []byte) {
	return derivePair(secret, []byte("gca-session-cookie-v1"))
}

func deriveOAuthStateKeys(secret []byte) (hashKey, blockKey []byte) {
	return derivePair(secret, []byte("gca-oauth-state-cookie-v1"))
}

func derivePair(secret, info []byte) (hashKey, blockKey []byte) {
	r := hkdf.New(sha256.New, secret, nil, info)
	hashKey = make([]byte, 64)
	blockKey = make([]byte, 32)
	_, _ = io.ReadFull(r, hashKey)
	_, _ = io.ReadFull(r, blockKey)
	return hashKey, blockKey
}

func (m *Manager) WriteSession(w http.ResponseWriter, p *Payload) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	encoded, err := m.codec.Encode(m.cfg.CookieName, b)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     m.cfg.CookieName,
		Value:    encoded,
		Path:     "/",
		MaxAge:   m.cfg.MaxAgeSeconds,
		HttpOnly: true,
		Secure:   m.cfg.SecureCookies,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (m *Manager) ReadSession(r *http.Request) (*Payload, error) {
	c, err := r.Cookie(m.cfg.CookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, nil
		}
		return nil, err
	}
	if c.Value == "" {
		return nil, nil
	}
	var raw []byte
	if err := m.codec.Decode(m.cfg.CookieName, c.Value, &raw); err != nil {
		return nil, err
	}
	var p Payload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (m *Manager) ClearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.cfg.CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   m.cfg.SecureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}

func (m *Manager) WriteOAuthState(w http.ResponseWriter, state string) error {
	encoded, err := m.stCodec.Encode(oauthStateCookie, state)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    encoded,
		Path:     "/auth/google/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   m.cfg.SecureCookies,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (m *Manager) ReadOAuthState(r *http.Request) (string, error) {
	c, err := r.Cookie(oauthStateCookie)
	if err != nil || c.Value == "" {
		return "", errors.New("missing oauth state cookie")
	}
	var state string
	if err := m.stCodec.Decode(oauthStateCookie, c.Value, &state); err != nil {
		return "", err
	}
	return state, nil
}

func (m *Manager) ClearOAuthState(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    "",
		Path:     "/auth/google/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   m.cfg.SecureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}
