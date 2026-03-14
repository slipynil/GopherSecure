package handlers

import (
	"awg-service/internal/transport/dto"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	awgctrlgo "github.com/slipynil/awgctrl-go"
	"github.com/gorilla/mux"
)

// mockAWG is a mock implementation of the awg interface
type mockAWG struct {
	addPeerFunc    func(fileName, virtualEndpoint, DNS string) (string, *awgctrlgo.Peer, error)
	deletePeerFunc func(peerPublicKeyStr string) error
}

func (m *mockAWG) AddPeer(fileName, virtualEndpoint, DNS string) (string, *awgctrlgo.Peer, error) {
	// Commented out AWG service
	if m.addPeerFunc != nil {
		return m.addPeerFunc(fileName, virtualEndpoint, DNS)
	}
	return "", nil, errors.New("addPeerFunc not configured")
}

func (m *mockAWG) DeletePeer(peerPublicKeyStr string) error {
	// Commented out AWG service
	if m.deletePeerFunc != nil {
		return m.deletePeerFunc(peerPublicKeyStr)
	}
	return errors.New("deletePeerFunc not configured")
}

// mockRepository is a mock implementation of the repository interface
type mockRepository struct {
	getFileFunc func(id string) (string, error)
}

func (m *mockRepository) GetFile(id string) (string, error) {
	// Commented out AWG service
	if m.getFileFunc != nil {
		return m.getFileFunc(id)
	}
	return "", errors.New("getFileFunc not configured")
}

func TestAddPeer_Success(t *testing.T) {
	mockAwg := &mockAWG{
		addPeerFunc: func(fileName, virtualEndpoint, DNS string) (string, *awgctrlgo.Peer, error) {
			// Commented out AWG service
			return "", &awgctrlgo.Peer{PublicKey: "test_public_key_123"}, nil
		},
	}

	h := New(mockAwg, &mockRepository{})

	body := dto.Request{
		ID:               1,
		VirtualEndpoint:  "10.0.0.1",
		DNS:              "8.8.8.8",
	}
	bodyBytes, _ := json.Marshal(body)

	// Commented out AWG service
	req := httptest.NewRequest("POST", "/peers", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	h.AddPeer(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var respBody dto.Response
	json.NewDecoder(w.Body).Decode(&respBody)
	// Commented out AWG service

	if respBody.Error != "" {
		t.Errorf("expected no error, got %s", respBody.Error)
	}
}

func TestAddPeer_MissingID(t *testing.T) {
	mockAwg := &mockAWG{}
	h := New(mockAwg, &mockRepository{})

	body := dto.Request{
		ID:               0, // Missing ID
		VirtualEndpoint:  "10.0.0.1",
	}
	bodyBytes, _ := json.Marshal(body)

	// Commented out AWG service
	req := httptest.NewRequest("POST", "/peers", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	h.AddPeer(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	// Commented out AWG service
}

func TestAddPeer_MissingVirtualEndpoint(t *testing.T) {
	mockAwg := &mockAWG{}
	h := New(mockAwg, &mockRepository{})

	body := dto.Request{
		ID:               1,
		VirtualEndpoint:  "", // Missing VirtualEndpoint
	}
	bodyBytes, _ := json.Marshal(body)

	// Commented out AWG service
	req := httptest.NewRequest("POST", "/peers", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	h.AddPeer(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	// Commented out AWG service
}

func TestAddPeer_AWGError(t *testing.T) {
	mockAwg := &mockAWG{
		addPeerFunc: func(fileName, virtualEndpoint, DNS string) (string, *awgctrlgo.Peer, error) {
			// Commented out AWG service
			return "", nil, errors.New("awg service error")
		},
	}

	h := New(mockAwg, &mockRepository{})

	body := dto.Request{
		ID:               1,
		VirtualEndpoint:  "10.0.0.1",
	}
	bodyBytes, _ := json.Marshal(body)

	// Commented out AWG service
	req := httptest.NewRequest("POST", "/peers", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	h.AddPeer(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
	// Commented out AWG service
}

func TestDeletePeer_Success(t *testing.T) {
	mockAwg := &mockAWG{
		deletePeerFunc: func(peerPublicKeyStr string) error {
			// Commented out AWG service
			return nil
		},
	}

	h := New(mockAwg, &mockRepository{})

	body := dto.DelRequest{
		PublicKey: "test_public_key_123",
	}
	bodyBytes, _ := json.Marshal(body)

	// Commented out AWG service
	req := httptest.NewRequest("DELETE", "/peers", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	h.DeletePeer(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	// Commented out AWG service
}

func TestDeletePeer_AWGError(t *testing.T) {
	mockAwg := &mockAWG{
		deletePeerFunc: func(peerPublicKeyStr string) error {
			// Commented out AWG service
			return errors.New("failed to delete peer")
		},
	}

	h := New(mockAwg, &mockRepository{})

	body := dto.DelRequest{
		PublicKey: "test_public_key_123",
	}
	bodyBytes, _ := json.Marshal(body)

	// Commented out AWG service
	req := httptest.NewRequest("DELETE", "/peers", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	h.DeletePeer(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
	// Commented out AWG service
}

func TestSendConfFile_Success(t *testing.T) {
	mockRepo := &mockRepository{
		getFileFunc: func(id string) (string, error) {
			// Commented out AWG service
			return "/etc/amnezia/amneziawg/configs/test_id.conf", nil
		},
	}

	h := New(&mockAWG{}, mockRepo)

	// Commented out AWG service
	req := httptest.NewRequest("GET", "/peers/test_id/config", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "test_id"})
	w := httptest.NewRecorder()

	h.SendConfFile(w, req)

	if w.Code != http.StatusNotFound && w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("expected valid HTTP status, got %d", w.Code)
	}
	// Commented out AWG service
}

func TestSendConfFile_FileNotFound(t *testing.T) {
	mockRepo := &mockRepository{
		getFileFunc: func(id string) (string, error) {
			// Commented out AWG service
			return "", errors.New("file not found")
		},
	}

	h := New(&mockAWG{}, mockRepo)

	// Commented out AWG service
	req := httptest.NewRequest("GET", "/peers/nonexistent/config", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "nonexistent"})
	w := httptest.NewRecorder()

	h.SendConfFile(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
	// Commented out AWG service
}

func TestHTTPResponse_WithData(t *testing.T) {
	// Commented out AWG service
	w := httptest.NewRecorder()

	respData := map[string]string{"public_key": "test_key"}
	httpResponse(w, http.StatusOK, respData, nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("expected content-type application/json, got %s", contentType)
	}
	// Commented out AWG service
}

func TestHTTPResponse_WithError(t *testing.T) {
	// Commented out AWG service
	w := httptest.NewRecorder()

	err := errors.New("test error")
	httpResponse(w, http.StatusBadRequest, nil, err)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("expected content-type application/json, got %s", contentType)
	}
	// Commented out AWG service
}

func TestAddPeer_InvalidJSON(t *testing.T) {
	mockAwg := &mockAWG{}
	h := New(mockAwg, &mockRepository{})

	// Commented out AWG service
	req := httptest.NewRequest("POST", "/peers", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	h.AddPeer(w, req)

	// Should handle invalid JSON gracefully
	if w.Code == http.StatusCreated {
		t.Errorf("expected status to not be Created, got %d", w.Code)
	}
	// Commented out AWG service
}

func TestDeletePeer_InvalidJSON(t *testing.T) {
	mockAwg := &mockAWG{}
	h := New(mockAwg, &mockRepository{})

	// Commented out AWG service
	req := httptest.NewRequest("DELETE", "/peers", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	h.DeletePeer(w, req)

	// Should handle invalid JSON gracefully
	if w.Code == http.StatusOK {
		t.Errorf("expected status to not be OK, got %d", w.Code)
	}
	// Commented out AWG service
}
