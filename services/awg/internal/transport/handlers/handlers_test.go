package handlers

import (
	"awg-service/internal/transport/dto"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
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
	if m.addPeerFunc != nil {
		return m.addPeerFunc(fileName, virtualEndpoint, DNS)
	}
	return "", nil, errors.New("addPeerFunc not configured")
}

func (m *mockAWG) DeletePeer(peerPublicKeyStr string) error {
	if m.deletePeerFunc != nil {
		return m.deletePeerFunc(peerPublicKeyStr)
	}
	return errors.New("deletePeerFunc not configured")
}

// mockRepository is a mock implementation of the repository interface
type mockRepository struct {
	addUserFunc    func(id int64, peer *awgctrlgo.Peer) error
	deleteUserFunc func(publicKey string) error
	getFileFunc    func(id string) (string, error)
}

func (m *mockRepository) AddUser(id int64, peer *awgctrlgo.Peer) error {
	if m.addUserFunc != nil {
		return m.addUserFunc(id, peer)
	}
	return errors.New("addUserFunc not configured")
}

func (m *mockRepository) DeleteUser(publicKey string) error {
	if m.deleteUserFunc != nil {
		return m.deleteUserFunc(publicKey)
	}
	return nil
}

func (m *mockRepository) GetFile(id string) (string, error) {
	if m.getFileFunc != nil {
		return m.getFileFunc(id)
	}
	return "", errors.New("getFileFunc not configured")
}

func TestAddPeer_Success(t *testing.T) {
	mockAwg := &mockAWG{
		addPeerFunc: func(fileName, virtualEndpoint, DNS string) (string, *awgctrlgo.Peer, error) {
			return "", &awgctrlgo.Peer{PublicKey: "test_public_key_123"}, nil
		},
	}

	mockRepo := &mockRepository{
		addUserFunc: func(id int64, peer *awgctrlgo.Peer) error {
			return nil
		},
	}

	h := New(mockAwg, mockRepo)

	body := dto.Request{
		ID:              1,
		VirtualEndpoint: "10.0.0.1",
		DNS:             "8.8.8.8",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/peers", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	h.AddPeer(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var respBody dto.Response
	json.NewDecoder(w.Body).Decode(&respBody)

	if respBody.Error != "" {
		t.Errorf("expected no error, got %s", respBody.Error)
	}
}

func TestAddPeer_MissingID(t *testing.T) {
	mockAwg := &mockAWG{}
	mockRepo := &mockRepository{}
	h := New(mockAwg, mockRepo)

	body := dto.Request{
		ID:              0, // Missing ID
		VirtualEndpoint: "10.0.0.1",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/peers", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	h.AddPeer(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAddPeer_MissingVirtualEndpoint(t *testing.T) {
	mockAwg := &mockAWG{}
	mockRepo := &mockRepository{}
	h := New(mockAwg, mockRepo)

	body := dto.Request{
		ID:              1,
		VirtualEndpoint: "", // Missing VirtualEndpoint
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/peers", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	h.AddPeer(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAddPeer_AWGError(t *testing.T) {
	mockAwg := &mockAWG{
		addPeerFunc: func(fileName, virtualEndpoint, DNS string) (string, *awgctrlgo.Peer, error) {
			return "", nil, errors.New("awg service error")
		},
	}

	mockRepo := &mockRepository{}
	h := New(mockAwg, mockRepo)

	body := dto.Request{
		ID:              1,
		VirtualEndpoint: "10.0.0.1",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/peers", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	h.AddPeer(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestAddPeer_RepositoryError(t *testing.T) {
	mockAwg := &mockAWG{
		addPeerFunc: func(fileName, virtualEndpoint, DNS string) (string, *awgctrlgo.Peer, error) {
			return "", &awgctrlgo.Peer{PublicKey: "test_key"}, nil
		},
	}

	mockRepo := &mockRepository{
		addUserFunc: func(id int64, peer *awgctrlgo.Peer) error {
			return errors.New("database error")
		},
	}

	h := New(mockAwg, mockRepo)

	body := dto.Request{
		ID:              1,
		VirtualEndpoint: "10.0.0.1",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/peers", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	h.AddPeer(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestAddPeer_InvalidJSON(t *testing.T) {
	mockAwg := &mockAWG{}
	mockRepo := &mockRepository{}
	h := New(mockAwg, mockRepo)

	req := httptest.NewRequest("POST", "/peers", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	h.AddPeer(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAddPeer_EmptyBody(t *testing.T) {
	mockAwg := &mockAWG{}
	mockRepo := &mockRepository{}
	h := New(mockAwg, mockRepo)

	req := httptest.NewRequest("POST", "/peers", bytes.NewReader([]byte("")))
	w := httptest.NewRecorder()

	h.AddPeer(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAddPeer_VerifyPublicKeyInResponse(t *testing.T) {
	const expectedPublicKey = "test_public_key_12345"
	mockAwg := &mockAWG{
		addPeerFunc: func(fileName, virtualEndpoint, DNS string) (string, *awgctrlgo.Peer, error) {
			return "", &awgctrlgo.Peer{PublicKey: expectedPublicKey}, nil
		},
	}

	mockRepo := &mockRepository{
		addUserFunc: func(id int64, peer *awgctrlgo.Peer) error {
			return nil
		},
	}

	h := New(mockAwg, mockRepo)

	body := dto.Request{
		ID:              1,
		VirtualEndpoint: "10.0.0.1",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/peers", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	h.AddPeer(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Unmarshal as generic JSON to inspect the structure
	var rawResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&rawResp)

	// The response wraps dto.Response in the data field
	// Structure: {"data": {"Data": {"public_key": "..."}, "Error": ""}}
	outerData, hasData := rawResp["data"].(map[string]interface{})
	if !hasData {
		t.Fatalf("expected data field in response, got: %+v", rawResp)
	}

	innerData, hasInnerData := outerData["Data"].(map[string]interface{})
	if !hasInnerData {
		t.Fatalf("expected Data field in outer data, got: %+v", outerData)
	}

	publicKey, hasKey := innerData["public_key"].(string)
	if !hasKey {
		t.Errorf("expected public_key field in inner data")
	}

	if publicKey != expectedPublicKey {
		t.Errorf("expected public_key %s, got %s", expectedPublicKey, publicKey)
	}
}

func TestDeletePeer_Success(t *testing.T) {
	mockAwg := &mockAWG{
		deletePeerFunc: func(peerPublicKeyStr string) error {
			return nil
		},
	}

	mockRepo := &mockRepository{}
	h := New(mockAwg, mockRepo)

	body := dto.DelRequest{
		PublicKey: "test_public_key_123",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("DELETE", "/peers", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	h.DeletePeer(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDeletePeer_AWGError(t *testing.T) {
	mockAwg := &mockAWG{
		deletePeerFunc: func(peerPublicKeyStr string) error {
			return errors.New("failed to delete peer")
		},
	}

	mockRepo := &mockRepository{}
	h := New(mockAwg, mockRepo)

	body := dto.DelRequest{
		PublicKey: "test_public_key_123",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("DELETE", "/peers", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	h.DeletePeer(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestDeletePeer_InvalidJSON(t *testing.T) {
	mockAwg := &mockAWG{}
	mockRepo := &mockRepository{}
	h := New(mockAwg, mockRepo)

	req := httptest.NewRequest("DELETE", "/peers", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	h.DeletePeer(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDeletePeer_RepositoryError(t *testing.T) {
	mockAwg := &mockAWG{
		deletePeerFunc: func(peerPublicKeyStr string) error {
			return nil
		},
	}

	mockRepo := &mockRepository{
		deleteUserFunc: func(publicKey string) error {
			return errors.New("database error")
		},
	}

	h := New(mockAwg, mockRepo)

	body := dto.DelRequest{
		PublicKey: "test_public_key_123",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("DELETE", "/peers", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	h.DeletePeer(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestDeletePeer_EmptyPublicKey(t *testing.T) {
	deletePeerReceivedKey := ""
	deleteUserReceivedKey := ""

	mockAwg := &mockAWG{
		deletePeerFunc: func(peerPublicKeyStr string) error {
			deletePeerReceivedKey = peerPublicKeyStr
			return nil
		},
	}

	mockRepo := &mockRepository{
		deleteUserFunc: func(publicKey string) error {
			deleteUserReceivedKey = publicKey
			return nil
		},
	}

	h := New(mockAwg, mockRepo)

	body := dto.DelRequest{
		PublicKey: "",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("DELETE", "/peers", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	h.DeletePeer(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if deletePeerReceivedKey != "" {
		t.Errorf("expected empty string passed to awg.DeletePeer, got %s", deletePeerReceivedKey)
	}

	if deleteUserReceivedKey != "" {
		t.Errorf("expected empty string passed to repository.DeleteUser, got %s", deleteUserReceivedKey)
	}
}

func TestSendConfFile_Success(t *testing.T) {
	// Create a temporary test file
	tempFile := t.TempDir() + "/test.conf"
	os.WriteFile(tempFile, []byte("[Interface]\nAddress = 10.0.0.1"), 0o644)

	mockRepo := &mockRepository{
		getFileFunc: func(id string) (string, error) {
			return tempFile, nil
		},
	}

	h := New(&mockAWG{}, mockRepo)

	req := httptest.NewRequest("GET", "/peers/test_id/config", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "test_id"})
	w := httptest.NewRecorder()

	h.SendConfFile(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestSendConfFile_FileNotFound(t *testing.T) {
	mockRepo := &mockRepository{
		getFileFunc: func(id string) (string, error) {
			return "", errors.New("file not found")
		},
	}

	h := New(&mockAWG{}, mockRepo)

	req := httptest.NewRequest("GET", "/peers/nonexistent/config", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "nonexistent"})
	w := httptest.NewRecorder()

	h.SendConfFile(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestSendConfFile_VerifyContent(t *testing.T) {
	expectedContent := "[Interface]\nAddress = 10.0.0.1\nPrivateKey = example_key"
	tempFile := t.TempDir() + "/test.conf"
	os.WriteFile(tempFile, []byte(expectedContent), 0o644)

	mockRepo := &mockRepository{
		getFileFunc: func(id string) (string, error) {
			return tempFile, nil
		},
	}

	h := New(&mockAWG{}, mockRepo)

	req := httptest.NewRequest("GET", "/peers/test_id/config", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "test_id"})
	w := httptest.NewRecorder()

	h.SendConfFile(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if body != expectedContent {
		t.Errorf("expected response body %q, got %q", expectedContent, body)
	}
}

func TestHTTPResponse_WithData(t *testing.T) {
	w := httptest.NewRecorder()

	respData := map[string]string{"public_key": "test_key"}
	httpResponse(w, http.StatusOK, respData, nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("expected content-type application/json, got %s", contentType)
	}
}

func TestHTTPResponse_WithError(t *testing.T) {
	w := httptest.NewRecorder()

	err := errors.New("test error")
	httpResponse(w, http.StatusBadRequest, nil, err)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("expected content-type application/json, got %s", contentType)
	}
}

func TestHTTPResponse_BothNil(t *testing.T) {
	w := httptest.NewRecorder()

	httpResponse(w, http.StatusOK, nil, nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("expected content-type application/json, got %s", contentType)
	}

	var respBody dto.Response
	json.NewDecoder(w.Body).Decode(&respBody)

	if respBody.Data != nil {
		t.Errorf("expected data field to be nil or omitted, got %v", respBody.Data)
	}

	if respBody.Error != "" {
		t.Errorf("expected error field to be empty, got %s", respBody.Error)
	}
}
