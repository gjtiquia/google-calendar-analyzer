package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func testSecret() []byte {
	return []byte("01234567890123456789012345678901")
}

func TestManager_sessionRoundTrip(t *testing.T) {
	m := NewManager(Config{
		CookieName:    "t_sess",
		SessionSecret: testSecret(),
		MaxAgeSeconds: 3600,
		SecureCookies: false,
	})
	w := httptest.NewRecorder()
	want := &Payload{
		Sub:                  "sub1",
		Email:                "u@example.com",
		AccessToken:          "tok",
		AccessTokenExpiryUTC: time.Now().Add(time.Hour).Unix(),
	}
	if err := m.WriteSession(w, want); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}
	got, err := m.ReadSession(req)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("nil payload")
	}
	if got.Sub != want.Sub || got.Email != want.Email || got.AccessToken != want.AccessToken {
		t.Fatalf("got %+v", got)
	}
}

func TestPayload_Expired(t *testing.T) {
	p := &Payload{AccessTokenExpiryUTC: time.Now().Add(-time.Minute).Unix()}
	if !p.Expired(time.Now()) {
		t.Fatal("expected expired")
	}
	p2 := &Payload{AccessTokenExpiryUTC: time.Now().Add(time.Hour).Unix()}
	if p2.Expired(time.Now()) {
		t.Fatal("expected valid")
	}
}

func TestManager_ReadSession_invalidCookie(t *testing.T) {
	m := NewManager(Config{
		CookieName:    "t_sess",
		SessionSecret: testSecret(),
		MaxAgeSeconds: 3600,
		SecureCookies: false,
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "t_sess", Value: "not-valid"})
	_, err := m.ReadSession(req)
	if err == nil {
		t.Fatal("expected decode error")
	}
}
