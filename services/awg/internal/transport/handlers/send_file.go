package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
)

// handler for sending configuration file by VAR=id
func (h *handlers) SendConfFile(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	filepath, err := h.repository.GetFile(id)
	if err != nil {
		httpResponse(w, http.StatusNotFound, nil, err)
		return
	}
	http.ServeFile(w, r, filepath)
}
