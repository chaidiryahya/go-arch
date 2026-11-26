package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Response is the envelope returned to API clients. It deliberately carries no
// request id: the correlation id is internal to the backend and surfaces only
// in logs.
type Response struct {
	Status            string      `json:"status"`
	ServerProcessTime string      `json:"server_process_time"`
	Data              interface{} `json:"data,omitempty"`
	ErrorMessage      string      `json:"error_message,omitempty"`
}

func writeJSON(w http.ResponseWriter, clk clock, ctx context.Context, statusCode int, startTime time.Time, data interface{}, errMsg string) {
	status := "ok"
	if statusCode >= 400 {
		status = "error"
	}

	resp := Response{
		Status:            status,
		ServerProcessTime: clk.Now().Sub(startTime).String(),
		Data:              data,
		ErrorMessage:      errMsg,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

func WriteSuccess(w http.ResponseWriter, r *http.Request, clk clock, data interface{}) {
	ctx := r.Context()
	writeJSON(w, clk, ctx, http.StatusOK, GetStartTime(ctx, clk), data, "")
}

func WriteError(w http.ResponseWriter, r *http.Request, clk clock, statusCode int, errMsg string) {
	ctx := r.Context()
	writeJSON(w, clk, ctx, statusCode, GetStartTime(ctx, clk), nil, errMsg)
}
