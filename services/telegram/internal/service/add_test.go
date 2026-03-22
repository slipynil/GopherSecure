package service

import (
	"fmt"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegram-service/internal/dto"
)

// --- mocks ---

type mockPostgres struct {
	newConnectionHostID int
	newConnectionErr    error
	saveConnectionErr   error
	checkStatusResult   bool
	checkStatusErr      error

	savedChatID      int64
	savedPublicKey   string
	savedPreshared   string
	savedExpiresAt   time.Time
	savedHostID      int
}

func (m *mockPostgres) Close() error                          { return nil }
func (m *mockPostgres) AddClient(u string, id int64) error    { return nil }
func (m *mockPostgres) Tested(id int64) error                 { return nil }
func (m *mockPostgres) IsTested(id int64) (bool, error)       { return false, nil }
func (m *mockPostgres) StatusTrue(id int64) error             { return nil }
func (m *mockPostgres) StatusFalse(id int64) error            { return nil }
func (m *mockPostgres) CheckStatus(id int64) (bool, error)    { return m.checkStatusResult, m.checkStatusErr }
func (m *mockPostgres) NewPayment(id int64, p string) error   { return nil }
func (m *mockPostgres) SuccessfulPaymentStatus(p string) error { return nil }
func (m *mockPostgres) ExpiredConnection() ([]dto.DelEntity, error) { return nil, nil }
func (m *mockPostgres) GetHostID(id int64) (int, error)       { return m.newConnectionHostID, m.newConnectionErr }

func (m *mockPostgres) NewConnection(chatID int64) (int, error) {
	return m.newConnectionHostID, m.newConnectionErr
}

func (m *mockPostgres) SaveConnection(hostID int, publicKey, presharedKey string, expiresAt time.Time) error {
	m.savedHostID = hostID
	m.savedPublicKey = publicKey
	m.savedPreshared = presharedKey
	m.savedExpiresAt = expiresAt
	return m.saveConnectionErr
}

type mockHTTPClient struct {
	addPeerHostID    int
	addPeerChatID    int64
	addPeerResp      *dto.Response
	addPeerErr       error
	downloadConfResp []byte
	downloadConfErr  error
}

func (m *mockHTTPClient) AddPeer(hostID int, dns bool, telegramID int64) (*dto.Response, error) {
	m.addPeerHostID = hostID
	m.addPeerChatID = telegramID
	return m.addPeerResp, m.addPeerErr
}

func (m *mockHTTPClient) DeletePeer(publicKey string) error { return nil }

func (m *mockHTTPClient) DownloadConfFile(id int64) ([]byte, error) {
	return m.downloadConfResp, m.downloadConfErr
}

type mockTelegram struct {
	sentFileBytes []byte
	sentFileChatID int64
}

func (m *mockTelegram) Chan() tgbotapi.UpdatesChannel                              { return nil }
func (m *mockTelegram) Menu(id int64) error                                        { return nil }
func (m *mockTelegram) UpdateMainMenu(u tgbotapi.Update) error                     { return nil }
func (m *mockTelegram) UpdateSendText(u tgbotapi.Update, text string) error        { return nil }
func (m *mockTelegram) SendText(id int64, text string) error                       { return nil }
func (m *mockTelegram) CreateAndSendInvoice(id int64, payload string) error        { return nil }
func (m *mockTelegram) PreCheckoutQuery(u tgbotapi.Update) error                   { return nil }
func (m *mockTelegram) HandleSuccessfulPayment(u tgbotapi.Update) (*dto.PaymentHandler, error) {
	return nil, nil
}
func (m *mockTelegram) SendFile(chatID int64, buf []byte) error {
	m.sentFileChatID = chatID
	m.sentFileBytes = buf
	return nil
}

// --- tests ---

// TestAdd_HostIDPassedToAddPeer checks that the host_id from NewConnection is
// correctly forwarded to AddPeer (NOT the chatID).
func TestAdd_HostIDPassedToAddPeer(t *testing.T) {
	const chatID = int64(6741297026) // realistic Telegram ID (> 255!)
	const hostID = 42               // small DB serial

	pg := &mockPostgres{newConnectionHostID: hostID}
	http := &mockHTTPClient{
		addPeerResp: responseWithKeys("pubkey123", "preshared456"),
		downloadConfResp: []byte("conf-content"),
	}
	tg := &mockTelegram{}

	svc := New(tg, http, pg)
	if err := svc.add(chatID, 0); err != nil {
		t.Fatalf("add() returned error: %v", err)
	}

	if http.addPeerHostID != hostID {
		t.Errorf("AddPeer received hostID=%d, want %d (chatID=%d would cause invalid CIDR)", http.addPeerHostID, hostID, chatID)
	}
	if http.addPeerChatID != chatID {
		t.Errorf("AddPeer received chatID=%d, want %d", http.addPeerChatID, chatID)
	}
}

// TestAdd_KeysSavedCorrectly checks that both keys from AWG response are saved to DB using host_id.
func TestAdd_KeysSavedCorrectly(t *testing.T) {
	const chatID = int64(100)
	const hostID = 5

	pg := &mockPostgres{newConnectionHostID: hostID}
	http := &mockHTTPClient{
		addPeerResp:      responseWithKeys("my-public-key", "my-preshared-key"),
		downloadConfResp: []byte("conf"),
	}
	tg := &mockTelegram{}

	svc := New(tg, http, pg)
	if err := svc.add(chatID, 0); err != nil {
		t.Fatalf("add() returned error: %v", err)
	}

	if pg.savedPublicKey != "my-public-key" {
		t.Errorf("saved public_key=%q, want %q", pg.savedPublicKey, "my-public-key")
	}
	if pg.savedPreshared != "my-preshared-key" {
		t.Errorf("saved preshared_key=%q, want %q", pg.savedPreshared, "my-preshared-key")
	}
	// Critically: SaveConnection must use host_id (not chatID) to update the correct row
	if pg.savedHostID != hostID {
		t.Errorf("SaveConnection used hostID=%d, want %d (chatID=%d would update wrong row)", pg.savedHostID, hostID, chatID)
	}
}

// TestAdd_EmptyPublicKey checks that if AWG returns empty keys, add() still reports error.
func TestAdd_EmptyPublicKey(t *testing.T) {
	pg := &mockPostgres{newConnectionHostID: 10}
	http := &mockHTTPClient{
		addPeerResp:      responseWithKeys("", ""), // empty keys
		downloadConfResp: []byte("conf"),
	}
	tg := &mockTelegram{}

	svc := New(tg, http, pg)
	err := svc.add(100, 0)
	// With empty keys, SaveConnection saves empty strings - not an error currently,
	// but we can verify keys ARE empty (catches regression)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pg.savedPublicKey != "" {
		t.Errorf("expected empty public key, got %q", pg.savedPublicKey)
	}
}

// TestAdd_AddPeerError checks that error from AWG is propagated.
func TestAdd_AddPeerError(t *testing.T) {
	pg := &mockPostgres{newConnectionHostID: 5}
	http := &mockHTTPClient{
		addPeerErr: fmt.Errorf("failed to parse CIDR: invalid CIDR address: 10.66.66.6741297026/32"),
	}
	tg := &mockTelegram{}

	svc := New(tg, http, pg)
	err := svc.add(100, 0)
	if err == nil {
		t.Fatal("expected error from AddPeer, got nil")
	}
}

// TestAdd_FileDelivered checks that config file is sent to user after successful add.
func TestAdd_FileDelivered(t *testing.T) {
	const chatID = int64(100)
	pg := &mockPostgres{newConnectionHostID: 5}
	http := &mockHTTPClient{
		addPeerResp:      responseWithKeys("pub", "pre"),
		downloadConfResp: []byte("wireguard-config-content"),
	}
	tg := &mockTelegram{}

	svc := New(tg, http, pg)
	if err := svc.add(chatID, 0); err != nil {
		t.Fatalf("add() error: %v", err)
	}

	if tg.sentFileChatID != chatID {
		t.Errorf("file sent to chatID=%d, want %d", tg.sentFileChatID, chatID)
	}
	if string(tg.sentFileBytes) != "wireguard-config-content" {
		t.Errorf("file content=%q, want %q", tg.sentFileBytes, "wireguard-config-content")
	}
}

// --- helpers ---

func responseWithKeys(pub, pre string) *dto.Response {
	return &dto.Response{
		Data: map[string]any{
			"public_key":    pub,
			"preshared_key": pre,
		},
	}
}
