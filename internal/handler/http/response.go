package http

import (
	"encoding/json"
	"net/http"
	"time"
)

type Response struct {
	Status            string      `json:"status"`
	ServerProcessTime string      `json:"server_process_time"`
	Data              interface{} `json:"data,omitempty"`
	ErrorMessage      string      `json:"error_message,omitempty"`
}

func writeJSON(w http.ResponseWriter, statusCode int, startTime time.Time, data interface{}, errMsg string) {
	status := "ok"
	if statusCode >= 400 {
		status = "error"
	}

	resp := Response{
		Status:            status,
		ServerProcessTime: time.Since(startTime).String(),
		Data:              data,
		ErrorMessage:      errMsg,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

func WriteSuccess(w http.ResponseWriter, startTime time.Time, data interface{}) {
	writeJSON(w, http.StatusOK, startTime, data, "")
}

func WriteError(w http.ResponseWriter, startTime time.Time, statusCode int, errMsg string) {
	writeJSON(w, statusCode, startTime, nil, errMsg)
}
