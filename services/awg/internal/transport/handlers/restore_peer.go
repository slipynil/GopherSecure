package handlers

import (
	"awg-service/internal/transport/dto"
	"encoding/json"
	"fmt"
	"net/http"
)

// RestorePeer handles POST /peers/restore — registers an existing peer back into WireGuard.
// Does not generate new keys or create a new config file.
// Used when renewing a subscription so the user's old .conf file keeps working.
func (h *handlers) RestorePeer(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req dto.RestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpResponse(w, http.StatusBadRequest, nil, err)
		return
	}

	if req.PublicKey == "" || req.PresharedKey == "" || req.Socket == "" {
		httpResponse(w, http.StatusBadRequest, nil, fmt.Errorf("public_key, preshared_key and socket are required"))
		return
	}

	if err := h.awg.RestorePeer(req.PublicKey, req.PresharedKey, req.Socket); err != nil {
		httpResponse(w, http.StatusInternalServerError, nil, err)
		return
	}

	httpResponse(w, http.StatusOK, nil, nil)
}
