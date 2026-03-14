package httpclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"telegram-service/internal/dto"
)

func TestNew(t *testing.T) {
	endpoint := "http://example.com"
	c := New(endpoint)

	if c.url != endpoint {
		t.Errorf("expected url %s, got %s", endpoint, c.url)
	}
	if c.http == nil {
		t.Error("expected http client to be initialized")
	}
}

func TestAddPeerWithDNS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/peers" {
			t.Errorf("expected /peers, got %s", r.URL.Path)
		}

		var req dto.AddPeerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.ID != 12345 {
			t.Errorf("expected id 12345, got %d", req.ID)
		}
		if req.VirtualEndpoint != "10.66.66.5/32" {
			t.Errorf("expected endpoint 10.66.66.5/32, got %s", req.VirtualEndpoint)
		}
		if req.DNS != "1.1.1.1, 8.8.8.8" {
			t.Errorf("expected DNS, got %s", req.DNS)
		}

		resp := dto.Response{
			Data: map[string]any{
				"public_key": "test_key_123",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	resp, err := c.AddPeer(5, true, 12345)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.GetKey() != "test_key_123" {
		t.Errorf("expected key test_key_123, got %s", resp.GetKey())
	}
}

func TestAddPeerWithoutDNS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req dto.AddPeerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.DNS != "" {
			t.Errorf("expected empty DNS, got %s", req.DNS)
		}

		resp := dto.Response{
			Data: map[string]any{
				"public_key": "test_key_456",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	resp, err := c.AddPeer(10, false, 54321)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetKey() != "test_key_456" {
		t.Errorf("expected key test_key_456, got %s", resp.GetKey())
	}
}

func TestAddPeerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := dto.Response{
			Error: "invalid peer configuration",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	_, err := c.AddPeer(1, true, 12345)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeletePeer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/peers" {
			t.Errorf("expected /peers, got %s", r.URL.Path)
		}

		var req dto.DelPeerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.PublicKey != "test_public_key" {
			t.Errorf("expected key test_public_key, got %s", req.PublicKey)
		}

		resp := dto.Response{}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	err := c.DeletePeer("test_public_key")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeletePeerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := dto.Response{
			Error: "peer not found",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	err := c.DeletePeer("nonexistent_key")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDownloadConfFile(t *testing.T) {
	configData := []byte("# WireGuard Config\n[Interface]\n...")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/peers/12345/config" {
			t.Errorf("expected /peers/12345/config, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write(configData)
	}))
	defer server.Close()

	c := New(server.URL)
	data, err := c.DownloadConfFile(12345)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != string(configData) {
		t.Errorf("expected %s, got %s", string(configData), string(data))
	}
}

func TestDownloadConfFileNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	c := New(server.URL)
	_, err := c.DownloadConfFile(99999)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDownloadConfFileServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c := New(server.URL)
	_, err := c.DownloadConfFile(12345)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestResponseDecodeSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := dto.Response{
			Data: map[string]any{
				"public_key": "decoded_key",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	httpResp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}
	defer httpResp.Body.Close()

	resp, err := responseDecode(httpResp)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetKey() != "decoded_key" {
		t.Errorf("expected decoded_key, got %s", resp.GetKey())
	}
}

func TestResponseDecodeWithError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := dto.Response{
			Error: "some error message",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	httpResp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}
	defer httpResp.Body.Close()

	_, err = responseDecode(httpResp)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestResponseDecodeInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	httpResp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}
	defer httpResp.Body.Close()

	_, err = responseDecode(httpResp)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
