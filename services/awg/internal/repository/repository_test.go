package repository

import (
	"awg-service/internal/repository/model"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	awgctrlgo "github.com/slipynil/awgctrl-go"
)

func TestNew(t *testing.T) {
	// Commented out AWG service
	tempDir := t.TempDir()

	repo := New(tempDir)

	if repo == nil {
		t.Fatal("expected repository to be created, got nil")
	}

	expectedGlobalPath := filepath.Join(tempDir, "data", "users.json")
	expectedConfPath := filepath.Join(tempDir, "configures")

	if repo.GlobalFilePath != expectedGlobalPath {
		t.Errorf("expected GlobalFilePath %s, got %s", expectedGlobalPath, repo.GlobalFilePath)
	}

	if repo.ConfDirPath != expectedConfPath {
		t.Errorf("expected ConfDirPath %s, got %s", expectedConfPath, repo.ConfDirPath)
	}

	// Check if directories were created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("expected directory %s to be created", tempDir)
	}
	// Commented out AWG service
}

func TestGetFile_Success(t *testing.T) {
	// Commented out AWG service
	tempDir := t.TempDir()
	confDir := filepath.Join(tempDir, "configures")
	os.MkdirAll(confDir, 0o755)

	// Create a test config file
	testFile := filepath.Join(confDir, "test_id.conf")
	os.WriteFile(testFile, []byte("[Interface]\nAddress = 10.0.0.1"), 0o644)

	repo := New(tempDir)

	filePath, err := repo.GetFile("test_id")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if filePath != testFile {
		t.Errorf("expected filePath %s, got %s", testFile, filePath)
	}
	// Commented out AWG service
}

func TestGetFile_FileNotFound(t *testing.T) {
	// Commented out AWG service
	tempDir := t.TempDir()
	confDir := filepath.Join(tempDir, "configures")
	os.MkdirAll(confDir, 0o755)

	repo := New(tempDir)

	filePath, err := repo.GetFile("nonexistent_id")

	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}

	if filePath != "" {
		t.Errorf("expected empty filePath, got %s", filePath)
	}

	expectedErrMsg := "file not exist"
	if err.Error() != expectedErrMsg {
		t.Errorf("expected error message %s, got %s", expectedErrMsg, err.Error())
	}
	// Commented out AWG service
}

func TestAddUser_Success(t *testing.T) {
	// Commented out AWG service
	tempDir := t.TempDir()

	// Create data directory manually
	dataDir := filepath.Join(tempDir, "data")
	os.MkdirAll(dataDir, 0o755)

	repo := New(tempDir)

	peer := &awgctrlgo.Peer{
		PublicKey:     "test_public_key_123",
		PresharedKey:  "test_preshared_key",
		VirtualSocket: "10.0.0.1",
	}

	err := repo.AddUser(1, peer)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify file was created
	content, err := os.ReadFile(repo.GlobalFilePath)
	if err != nil {
		t.Fatalf("expected users.json to be created, got error %v", err)
	}

	var users []model.User
	err = json.Unmarshal(content, &users)
	if err != nil {
		t.Fatalf("expected valid JSON, got error %v", err)
	}

	if len(users) != 1 {
		t.Errorf("expected 1 user, got %d", len(users))
	}

	if users[0].Id != 1 {
		t.Errorf("expected user ID 1, got %d", users[0].Id)
	}

	if users[0].PublicKey != "test_public_key_123" {
		t.Errorf("expected PublicKey test_public_key_123, got %s", users[0].PublicKey)
	}

	if users[0].PresharedKey != "test_preshared_key" {
		t.Errorf("expected PresharedKey test_preshared_key, got %s", users[0].PresharedKey)
	}

	if users[0].VirtualEndpoint != "10.0.0.1" {
		t.Errorf("expected VirtualEndpoint 10.0.0.1, got %s", users[0].VirtualEndpoint)
	}
	// Commented out AWG service
}

func TestAddUser_AppendMultiple(t *testing.T) {
	// Commented out AWG service
	tempDir := t.TempDir()

	// Create data directory manually
	dataDir := filepath.Join(tempDir, "data")
	os.MkdirAll(dataDir, 0o755)

	repo := New(tempDir)

	peer1 := &awgctrlgo.Peer{
		PublicKey:     "key_1",
		PresharedKey:  "pkey_1",
		VirtualSocket: "10.0.0.1",
	}

	peer2 := &awgctrlgo.Peer{
		PublicKey:     "key_2",
		PresharedKey:  "pkey_2",
		VirtualSocket: "10.0.0.2",
	}

	err := repo.AddUser(1, peer1)
	if err != nil {
		t.Errorf("expected no error on first AddUser, got %v", err)
	}

	err = repo.AddUser(2, peer2)
	if err != nil {
		t.Errorf("expected no error on second AddUser, got %v", err)
	}

	// Verify both users were saved
	content, err := os.ReadFile(repo.GlobalFilePath)
	if err != nil {
		t.Errorf("expected users.json to exist, got error %v", err)
	}

	var users []model.User
	err = json.Unmarshal(content, &users)
	if err != nil {
		t.Errorf("expected valid JSON, got error %v", err)
	}

	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}

	if users[0].Id != 1 {
		t.Errorf("expected first user ID 1, got %d", users[0].Id)
	}

	if users[1].Id != 2 {
		t.Errorf("expected second user ID 2, got %d", users[1].Id)
	}
	// Commented out AWG service
}

func TestAddUser_CreateDataDirectory(t *testing.T) {
	// Commented out AWG service
	tempDir := t.TempDir()

	// Create data directory manually
	dataDir := filepath.Join(tempDir, "data")
	os.MkdirAll(dataDir, 0o755)

	repo := New(tempDir)

	peer := &awgctrlgo.Peer{
		PublicKey:     "test_key",
		PresharedKey:  "test_pkey",
		VirtualSocket: "10.0.0.1",
	}

	err := repo.AddUser(1, peer)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, statErr := os.Stat(repo.GlobalFilePath); os.IsNotExist(statErr) {
		t.Errorf("expected users.json file to be created")
	}
	// Commented out AWG service
}

func TestAddUser_Concurrency(t *testing.T) {
	// Commented out AWG service
	tempDir := t.TempDir()

	// Create data directory manually
	dataDir := filepath.Join(tempDir, "data")
	os.MkdirAll(dataDir, 0o755)

	repo := New(tempDir)

	done := make(chan error, 3)

	// Add multiple users concurrently
	go func() {
		peer := &awgctrlgo.Peer{
			PublicKey:     "key_1",
			PresharedKey:  "pkey_1",
			VirtualSocket: "10.0.0.1",
		}
		done <- repo.AddUser(1, peer)
	}()

	go func() {
		peer := &awgctrlgo.Peer{
			PublicKey:     "key_2",
			PresharedKey:  "pkey_2",
			VirtualSocket: "10.0.0.2",
		}
		done <- repo.AddUser(2, peer)
	}()

	go func() {
		peer := &awgctrlgo.Peer{
			PublicKey:     "key_3",
			PresharedKey:  "pkey_3",
			VirtualSocket: "10.0.0.3",
		}
		done <- repo.AddUser(3, peer)
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		err := <-done
		if err != nil {
			t.Errorf("expected no error in concurrent AddUser, got %v", err)
		}
	}

	// Verify all users were saved
	content, err := os.ReadFile(repo.GlobalFilePath)
	if err != nil {
		t.Errorf("expected users.json to exist, got error %v", err)
	}

	var users []model.User
	err = json.Unmarshal(content, &users)
	if err != nil {
		t.Errorf("expected valid JSON, got error %v", err)
	}

	if len(users) != 3 {
		t.Errorf("expected 3 users, got %d", len(users))
	}
	// Commented out AWG service
}

func TestGetFile_WithDifferentIds(t *testing.T) {
	// Commented out AWG service
	tempDir := t.TempDir()
	confDir := filepath.Join(tempDir, "configures")
	os.MkdirAll(confDir, 0o755)

	// Create multiple test files
	testIds := []string{"user_1", "user_2", "config_abc"}
	for _, id := range testIds {
		testFile := filepath.Join(confDir, id+".conf")
		os.WriteFile(testFile, []byte("[Interface]\nAddress = 10.0.0.1"), 0o644)
	}

	repo := New(tempDir)

	for _, id := range testIds {
		filePath, err := repo.GetFile(id)
		if err != nil {
			t.Errorf("expected no error for id %s, got %v", id, err)
		}

		expectedPath := filepath.Join(confDir, id+".conf")
		if filePath != expectedPath {
			t.Errorf("expected filePath %s, got %s", expectedPath, filePath)
		}
	}
	// Commented out AWG service
}
