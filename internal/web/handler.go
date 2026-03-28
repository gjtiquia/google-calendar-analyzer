package web

import (
	"net/http"

	"github.com/gjtiquia/google-calendar-analyzer/views/pages"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	if err := pages.Home().Render(r.Context(), w); err != nil {
		http.Error(w, "failed to render home page", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
