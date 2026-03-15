package handlers

import (
	"awg-service/internal/transport/dto"
	"encoding/json"
	"net/http"
)

// DeletePeer handles the DELETE request to delete a peer.
// use endpoint with publicKey VAR parameter
func (h *handlers) DeletePeer(w http.ResponseWriter, r *http.Request) {

	var req dto.DelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpResponse(w, http.StatusBadRequest, nil, err)
		return
	}

	// awg delete peer and get process status
	if err := h.awg.DeletePeer(req.PublicKey); err != nil {
		httpResponse(w, http.StatusInternalServerError, nil, err)
		return
	}
	if err := h.repository.DeleteUser(req.PublicKey); err != nil {
		httpResponse(w, http.StatusInternalServerError, nil, err)
		return
	}
	httpResponse(w, http.StatusOK, nil, nil)
}
