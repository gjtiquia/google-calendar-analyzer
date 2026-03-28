package auth

import (
	"net/http"
	"time"

	"github.com/gjtiquia/google-calendar-analyzer/internal/session"
	"golang.org/x/oauth2"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

type Handler struct {
	clientID     string
	clientSecret string
	redirectURL  string
	oauth        *oauth2.Config
	sess         *session.Manager
}

func NewHandler(clientID, clientSecret, redirectURL string, sess *session.Manager) *Handler {
	return &Handler{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		oauth: NewOAuth2Config(
			clientID,
			clientSecret,
			redirectURL,
		),
		sess: sess,
	}
}

func (h *Handler) oauthConfigured() bool {
	return h.clientID != "" && h.clientSecret != "" && h.redirectURL != ""
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if !h.oauthConfigured() {
		http.Redirect(w, r, "/?error=oauth_not_configured", http.StatusSeeOther)
		return
	}
	state, err := RandomState()
	if err != nil {
		http.Redirect(w, r, "/?error=state_gen", http.StatusSeeOther)
		return
	}
	if err := h.sess.WriteOAuthState(w, state); err != nil {
		http.Redirect(w, r, "/?error=state_cookie", http.StatusSeeOther)
		return
	}
	url := h.oauth.AuthCodeURL(state, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	if !h.oauthConfigured() {
		http.Error(w, "oauth not configured", http.StatusServiceUnavailable)
		return
	}
	if errParam := r.FormValue("error"); errParam != "" {
		http.Redirect(w, r, "/?error="+errParam, http.StatusSeeOther)
		return
	}
	stateCookie, err := h.sess.ReadOAuthState(r)
	if err != nil {
		http.Redirect(w, r, "/?error=invalid_state", http.StatusSeeOther)
		return
	}
	h.sess.ClearOAuthState(w)
	if r.FormValue("state") != stateCookie {
		http.Redirect(w, r, "/?error=state_mismatch", http.StatusSeeOther)
		return
	}
	code := r.FormValue("code")
	if code == "" {
		http.Redirect(w, r, "/?error=missing_code", http.StatusSeeOther)
		return
	}
	ctx := r.Context()
	tok, err := h.oauth.Exchange(ctx, code)
	if err != nil {
		http.Redirect(w, r, "/?error=token_exchange", http.StatusSeeOther)
		return
	}
	cli := h.oauth.Client(ctx, tok)
	svc, err := googleoauth2.NewService(ctx, option.WithHTTPClient(cli))
	if err != nil {
		http.Redirect(w, r, "/?error=userinfo_client", http.StatusSeeOther)
		return
	}
	ui, err := svc.Userinfo.Get().Do()
	if err != nil {
		http.Redirect(w, r, "/?error=userinfo", http.StatusSeeOther)
		return
	}
	expUnix := accessTokenExpiryUnix(tok)

	payload := &session.Payload{
		Sub:                  ui.Id,
		Email:                ui.Email,
		AccessToken:          tok.AccessToken,
		AccessTokenExpiryUTC: expUnix,
	}
	if err := h.sess.WriteSession(w, payload); err != nil {
		http.Redirect(w, r, "/?error=session_write", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func accessTokenExpiryUnix(tok *oauth2.Token) int64 {
	if tok == nil {
		return 0
	}
	if !tok.Expiry.IsZero() {
		return tok.Expiry.UTC().Unix()
	}
	if tok.ExpiresIn > 0 {
		return time.Now().UTC().Add(time.Duration(tok.ExpiresIn) * time.Second).Unix()
	}
	return 0
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	h.sess.ClearSession(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
