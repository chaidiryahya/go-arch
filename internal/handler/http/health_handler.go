package http

import (
	"net/http"
)

type HealthHandler struct {
	clk clock
}

func NewHealthHandler(clk clock) *HealthHandler {
	return &HealthHandler{clk: clk}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	WriteSuccess(w, r, h.clk, map[string]string{"status": "ok"})
}
