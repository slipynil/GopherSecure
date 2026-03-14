package handlers

import (
	"awg-service/internal/transport/dto"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// AddPeer handles the POST request to add a peer.
// use json body with publicKey, id parameters
func (h *handlers) AddPeer(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	var req dto.Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpResponse(w, http.StatusBadRequest, nil, err)
		return
	}

	// check if file name and virtual endpoint are empty
	if req.ID == 0 || req.VirtualEndpoint == "" {
		err := fmt.Errorf("id and virtual endpoint are required")
		httpResponse(w, http.StatusBadRequest, nil, err)
		return
	}

	fileID := strconv.FormatInt(req.ID, 10)
	_, peer, err := h.awg.AddPeer(fileID, req.VirtualEndpoint, req.DNS)

	if err != nil {
		httpResponse(w, http.StatusInternalServerError, nil, err)
		return
	}

	resp := dto.CreatePeerResponse(peer.PublicKey)
	httpResponse(w, http.StatusCreated, resp, nil)
}
