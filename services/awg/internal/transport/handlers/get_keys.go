package handlers

import (
	"awg-service/internal/transport/dto"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// GetKeys handles the GET request to retrieve both public_key and preshared_key for a peer.
func (h *handlers) GetKeys(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		httpResponse(w, http.StatusBadRequest, nil, fmt.Errorf("invalid id format"))
		return
	}

	user, err := h.repository.GetUser(idInt)
	if err != nil {
		httpResponse(w, http.StatusNotFound, nil, fmt.Errorf("user not found"))
		return
	}

	resp := dto.CreateKeysResponse(user.PublicKey, user.PresharedKey)
	httpResponse(w, http.StatusOK, resp, nil)
}
