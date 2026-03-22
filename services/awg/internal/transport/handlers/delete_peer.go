package handlers

import (
	"awg-service/internal/transport/dto"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// DeletePeer handles the DELETE request to delete a peer using transactional approach.
// Flow: Delete from JSON first, then from WireGuard (safer for rollback)
func (h *handlers) DeletePeer(w http.ResponseWriter, r *http.Request) {

	var req dto.DelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpResponse(w, http.StatusBadRequest, nil, err)
		return
	}

	// Step 1: Delete from JSON first (safe point for rollback)
	deleteResult, err := h.repository.DeleteUserEx(req.PublicKey)
	if err != nil {
		log.Printf("ERROR: Failed to delete user from JSON (publicKey: %s): %v", req.PublicKey, err)
		httpResponse(w, http.StatusInternalServerError, nil, fmt.Errorf("failed to delete from storage"))
		return
	}

	// If user not found, return 404
	if !deleteResult.Found {
		httpResponse(w, http.StatusNotFound, nil, fmt.Errorf("peer not found"))
		return
	}

	// Step 2: Delete from WireGuard
	if err := h.awg.DeletePeer(req.PublicKey); err != nil {
		// ROLLBACK: Restore user in JSON if WireGuard deletion fails
		log.Printf("ERROR: Failed to delete peer from WireGuard (publicKey: %s): %v. Attempting rollback...", req.PublicKey, err)

		if restoreErr := h.repository.RestoreUser(deleteResult.User); restoreErr != nil {
			// CRITICAL: Rollback failed - data inconsistency
			log.Printf("CRITICAL ERROR: Failed to restore user during rollback (publicKey: %s): %v", req.PublicKey, restoreErr)
			httpResponse(w, http.StatusInternalServerError, nil, fmt.Errorf("critical error: failed to rollback deletion"))
			return
		}

		// Rollback succeeded
		log.Printf("INFO: Successfully rolled back deletion for publicKey: %s", req.PublicKey)
		httpResponse(w, http.StatusInternalServerError, nil, fmt.Errorf("failed to delete peer from WireGuard"))
		return
	}

	// Success
	httpResponse(w, http.StatusOK, nil, nil)
}
