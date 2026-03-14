package handlers

import (
	"net/http"
)

type Response struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func httpResponse(
	w http.ResponseWriter,
	status int,
	data any,
	err error,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := Response{Data: data}

	if err != nil {
		resp.Error = err.Error()
	}

}
