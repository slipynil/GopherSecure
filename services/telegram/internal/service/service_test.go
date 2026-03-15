package service

import (
	"context"
	"errors"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegram-service/internal/dto"
	"telegram-service/logger"
)

// Mock implementations

type mockPostgres struct {
	AddClientFunc               func(username string, chatID int64) error
	TestedFunc                  func(chatID int64) error
	IsTestedFunc                func(chatID int64) bool
	StatusTrueFunc              func(chatID int64) error
	StatusFalseFunc             func(chatID int64) error
	CheckStatusFunc             func(chatID int64) bool
	NewPaymentFunc              func(chatID int64, payload string) error
	SuccessfulPaymentStatusFunc func(payload string) error
	NewConnectionFunc           func(chatID int64, expiresAt time.Time) (int, error)
	DeleteConnectionFunc        func(chatID int64) error
	RenewConnectionFunc         func(chatID int64, expiresAt time.Time) error
	GetPeerFunc                 func(chatID int64) (string, string, error)
	SaveKeysFunc                func(chatID int64, pubKey, psk string) error
	ExpiredConnectionFunc       func() ([]dto.DelEntity, error)
	CloseFunc                   func() error
}

func (m *mockPostgres) AddClient(username string, chatID int64) error {
	if m.AddClientFunc != nil {
		return m.AddClientFunc(username, chatID)
	}
	return nil
}

func (m *mockPostgres) Tested(chatID int64) error {
	if m.TestedFunc != nil {
		return m.TestedFunc(chatID)
	}
	return nil
}

func (m *mockPostgres) IsTested(chatID int64) bool {
	if m.IsTestedFunc != nil {
		return m.IsTestedFunc(chatID)
	}
	return false
}

func (m *mockPostgres) StatusTrue(chatID int64) error {
	if m.StatusTrueFunc != nil {
		return m.StatusTrueFunc(chatID)
	}
	return nil
}

func (m *mockPostgres) StatusFalse(chatID int64) error {
	if m.StatusFalseFunc != nil {
		return m.StatusFalseFunc(chatID)
	}
	return nil
}

func (m *mockPostgres) CheckStatus(chatID int64) bool {
	if m.CheckStatusFunc != nil {
		return m.CheckStatusFunc(chatID)
	}
	return false
}

func (m *mockPostgres) NewPayment(chatID int64, payload string) error {
	if m.NewPaymentFunc != nil {
		return m.NewPaymentFunc(chatID, payload)
	}
	return nil
}

func (m *mockPostgres) SuccessfulPaymentStatus(payload string) error {
	if m.SuccessfulPaymentStatusFunc != nil {
		return m.SuccessfulPaymentStatusFunc(payload)
	}
	return nil
}

func (m *mockPostgres) NewConnection(chatID int64, expiresAt time.Time) (int, error) {
	if m.NewConnectionFunc != nil {
		return m.NewConnectionFunc(chatID, expiresAt)
	}
	return 1, nil
}

func (m *mockPostgres) DeleteConnection(chatID int64) error {
	if m.DeleteConnectionFunc != nil {
		return m.DeleteConnectionFunc(chatID)
	}
	return nil
}

func (m *mockPostgres) RenewConnection(chatID int64, expiresAt time.Time) error {
	if m.RenewConnectionFunc != nil {
		return m.RenewConnectionFunc(chatID, expiresAt)
	}
	return nil
}

func (m *mockPostgres) GetPeer(chatID int64) (string, string, error) {
	if m.GetPeerFunc != nil {
		return m.GetPeerFunc(chatID)
	}
	return "", "", dto.ErrNotFound
}

func (m *mockPostgres) SaveKeys(chatID int64, pubKey, psk string) error {
	if m.SaveKeysFunc != nil {
		return m.SaveKeysFunc(chatID, pubKey, psk)
	}
	return nil
}

func (m *mockPostgres) ExpiredConnection() ([]dto.DelEntity, error) {
	if m.ExpiredConnectionFunc != nil {
		return m.ExpiredConnectionFunc()
	}
	return nil, nil
}

func (m *mockPostgres) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

type mockTelegramClient struct {
	ChanFunc                    func() tgbotapi.UpdatesChannel
	MenuFunc                    func(chatID int64) error
	UpdateMainMenuFunc          func(update tgbotapi.Update) error
	UpdateSendTextFunc          func(update tgbotapi.Update, text string) error
	SendTextFunc                func(chatID int64, text string) error
	SendFileFunc                func(chatID int64, buffer []byte) error
	CreateAndSendInvoiceFunc    func(chatID int64, payload string) error
	PreCheckoutQueryFunc        func(update tgbotapi.Update) error
	HandleSuccessfulPaymentFunc func(update tgbotapi.Update) (*dto.PaymentHandler, error)
}

func (m *mockTelegramClient) Chan() tgbotapi.UpdatesChannel {
	if m.ChanFunc != nil {
		return m.ChanFunc()
	}
	return nil
}

func (m *mockTelegramClient) Menu(chatID int64) error {
	if m.MenuFunc != nil {
		return m.MenuFunc(chatID)
	}
	return nil
}

func (m *mockTelegramClient) UpdateMainMenu(update tgbotapi.Update) error {
	if m.UpdateMainMenuFunc != nil {
		return m.UpdateMainMenuFunc(update)
	}
	return nil
}

func (m *mockTelegramClient) UpdateSendText(update tgbotapi.Update, text string) error {
	if m.UpdateSendTextFunc != nil {
		return m.UpdateSendTextFunc(update, text)
	}
	return nil
}

func (m *mockTelegramClient) SendText(chatID int64, text string) error {
	if m.SendTextFunc != nil {
		return m.SendTextFunc(chatID, text)
	}
	return nil
}

func (m *mockTelegramClient) SendFile(chatID int64, buffer []byte) error {
	if m.SendFileFunc != nil {
		return m.SendFileFunc(chatID, buffer)
	}
	return nil
}

func (m *mockTelegramClient) CreateAndSendInvoice(chatID int64, payload string) error {
	if m.CreateAndSendInvoiceFunc != nil {
		return m.CreateAndSendInvoiceFunc(chatID, payload)
	}
	return nil
}

func (m *mockTelegramClient) PreCheckoutQuery(update tgbotapi.Update) error {
	if m.PreCheckoutQueryFunc != nil {
		return m.PreCheckoutQueryFunc(update)
	}
	return nil
}

func (m *mockTelegramClient) HandleSuccessfulPayment(update tgbotapi.Update) (*dto.PaymentHandler, error) {
	if m.HandleSuccessfulPaymentFunc != nil {
		return m.HandleSuccessfulPaymentFunc(update)
	}
	return nil, nil
}

type mockHTTPClient struct {
	AddPeerFunc          func(hostID int, DNS bool, telegramID int64) (*dto.Response, error)
	DeletePeerFunc       func(publicKey string) error
	DownloadConfFileFunc func(telegramID int64) ([]byte, error)
}

func (m *mockHTTPClient) AddPeer(hostID int, DNS bool, telegramID int64) (*dto.Response, error) {
	if m.AddPeerFunc != nil {
		return m.AddPeerFunc(hostID, DNS, telegramID)
	}
	return nil, nil
}

func (m *mockHTTPClient) DeletePeer(publicKey string) error {
	if m.DeletePeerFunc != nil {
		return m.DeletePeerFunc(publicKey)
	}
	return nil
}

func (m *mockHTTPClient) DownloadConfFile(telegramID int64) ([]byte, error) {
	if m.DownloadConfFileFunc != nil {
		return m.DownloadConfFileFunc(telegramID)
	}
	return nil, nil
}

// helpers

func newTestService() (service, *mockTelegramClient, *mockHTTPClient, *mockPostgres) {
	tg := &mockTelegramClient{}
	http := &mockHTTPClient{}
	db := &mockPostgres{}
	return New(tg, http, db), tg, http, db
}

func callbackUpdate(chatID int64) tgbotapi.Update {
	return tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: chatID},
			},
		},
	}
}

// Tests

func TestNewService(t *testing.T) {
	svc, tg, http, db := newTestService()

	if svc.telegram != tg {
		t.Error("telegram client not set")
	}
	if svc.httpClient != http {
		t.Error("http client not set")
	}
	if svc.postgres != db {
		t.Error("postgres not set")
	}
}

// add() — new peer

func TestAdd_NewPeer_30Days(t *testing.T) {
	svc, tg, http, db := newTestService()

	newConnectionCalled := false
	saveKeysCalled := false

	db.GetPeerFunc = func(chatID int64) (string, string, error) {
		return "", "", dto.ErrNotFound
	}
	http.AddPeerFunc = func(hostID int, DNS bool, telegramID int64) (*dto.Response, error) {
		return &dto.Response{
			Data: map[string]any{
				"public_key":    "pub_key_abc",
				"preshared_key": "psk_xyz",
			},
		}, nil
	}
	db.NewConnectionFunc = func(chatID int64, expiresAt time.Time) (int, error) {
		newConnectionCalled = true
		diff := expiresAt.Sub(time.Now().Add(30 * 24 * time.Hour)).Abs()
		if diff > time.Minute {
			t.Errorf("expected ~30 days expiration, got diff: %v", diff)
		}
		return 5, nil
	}
	db.SaveKeysFunc = func(chatID int64, pubKey, psk string) error {
		saveKeysCalled = true
		if pubKey != "pub_key_abc" {
			t.Errorf("expected pub_key_abc, got %s", pubKey)
		}
		if psk != "psk_xyz" {
			t.Errorf("expected psk_xyz, got %s", psk)
		}
		return nil
	}
	http.DownloadConfFileFunc = func(telegramID int64) ([]byte, error) {
		return []byte("config"), nil
	}
	tg.SendFileFunc = func(chatID int64, buffer []byte) error { return nil }

	if err := svc.add(12345, 20000); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !newConnectionCalled {
		t.Error("NewConnection not called")
	}
	if !saveKeysCalled {
		t.Error("SaveKeys not called")
	}
}

func TestAdd_NewPeer_24Hours(t *testing.T) {
	svc, tg, http, db := newTestService()

	db.GetPeerFunc = func(chatID int64) (string, string, error) {
		return "", "", dto.ErrNotFound
	}
	http.AddPeerFunc = func(hostID int, DNS bool, telegramID int64) (*dto.Response, error) {
		return &dto.Response{
			Data: map[string]any{
				"public_key":    "pub_key",
				"preshared_key": "psk",
			},
		}, nil
	}
	db.NewConnectionFunc = func(chatID int64, expiresAt time.Time) (int, error) {
		diff := expiresAt.Sub(time.Now().Add(24 * time.Hour)).Abs()
		if diff > time.Minute {
			t.Errorf("expected ~24h expiration, got diff: %v", diff)
		}
		return 2, nil
	}
	db.SaveKeysFunc = func(chatID int64, pubKey, psk string) error { return nil }
	http.DownloadConfFileFunc = func(telegramID int64) ([]byte, error) {
		return []byte("config"), nil
	}
	tg.SendFileFunc = func(chatID int64, buffer []byte) error { return nil }

	if err := svc.add(12345, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// add() — renew existing peer

func TestAdd_RenewExistingPeer(t *testing.T) {
	svc, tg, http, db := newTestService()

	renewCalled := false

	db.GetPeerFunc = func(chatID int64) (string, string, error) {
		return "existing_pub_key", "existing_psk", nil
	}
	db.RenewConnectionFunc = func(chatID int64, expiresAt time.Time) error {
		renewCalled = true
		return nil
	}
	http.DownloadConfFileFunc = func(telegramID int64) ([]byte, error) {
		return []byte("config"), nil
	}
	tg.SendFileFunc = func(chatID int64, buffer []byte) error { return nil }

	if err := svc.add(12345, 20000); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !renewCalled {
		t.Error("RenewConnection not called for existing peer")
	}
}

func TestAdd_RenewPeer_DBError(t *testing.T) {
	svc, _, http, db := newTestService()

	db.GetPeerFunc = func(chatID int64) (string, string, error) {
		return "pub_key", "psk", nil
	}
	db.RenewConnectionFunc = func(chatID int64, expiresAt time.Time) error {
		return errors.New("db error")
	}
	http.DownloadConfFileFunc = func(telegramID int64) ([]byte, error) {
		return []byte("config"), nil
	}

	if err := svc.add(12345, 20000); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// add() — error paths for new peer

func TestAdd_NewConnection_DBError(t *testing.T) {
	svc, _, _, db := newTestService()

	db.GetPeerFunc = func(chatID int64) (string, string, error) {
		return "", "", dto.ErrNotFound
	}
	db.NewConnectionFunc = func(chatID int64, expiresAt time.Time) (int, error) {
		return 0, errors.New("db error")
	}

	if err := svc.add(12345, 20000); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAdd_AddPeer_Rollback(t *testing.T) {
	svc, _, http, db := newTestService()

	deleteConnectionCalled := false

	db.GetPeerFunc = func(chatID int64) (string, string, error) {
		return "", "", dto.ErrNotFound
	}
	db.NewConnectionFunc = func(chatID int64, expiresAt time.Time) (int, error) {
		return 3, nil
	}
	http.AddPeerFunc = func(hostID int, DNS bool, telegramID int64) (*dto.Response, error) {
		return nil, errors.New("awg unavailable")
	}
	db.DeleteConnectionFunc = func(chatID int64) error {
		deleteConnectionCalled = true
		return nil
	}

	if err := svc.add(12345, 20000); err == nil {
		t.Fatal("expected error, got nil")
	}
	if !deleteConnectionCalled {
		t.Error("DeleteConnection not called for rollback")
	}
}

func TestAdd_AddPeer_HTTPError(t *testing.T) {
	svc, _, http, db := newTestService()

	db.GetPeerFunc = func(chatID int64) (string, string, error) {
		return "", "", dto.ErrNotFound
	}
	db.NewConnectionFunc = func(chatID int64, expiresAt time.Time) (int, error) {
		return 3, nil
	}
	http.AddPeerFunc = func(hostID int, DNS bool, telegramID int64) (*dto.Response, error) {
		return nil, errors.New("awg unavailable")
	}

	if err := svc.add(12345, 20000); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAdd_SaveKeys_DBError(t *testing.T) {
	svc, _, http, db := newTestService()

	db.GetPeerFunc = func(chatID int64) (string, string, error) {
		return "", "", dto.ErrNotFound
	}
	http.AddPeerFunc = func(hostID int, DNS bool, telegramID int64) (*dto.Response, error) {
		return &dto.Response{Data: map[string]any{"public_key": "k", "preshared_key": "p"}}, nil
	}
	db.NewConnectionFunc = func(chatID int64, expiresAt time.Time) (int, error) { return 3, nil }
	db.SaveKeysFunc = func(chatID int64, pubKey, psk string) error {
		return errors.New("db error")
	}

	if err := svc.add(12345, 20000); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAdd_SaveKeys_Rollback(t *testing.T) {
	svc, _, http, db := newTestService()

	deletePeerCalled := false
	deleteConnectionCalled := false

	db.GetPeerFunc = func(chatID int64) (string, string, error) {
		return "", "", dto.ErrNotFound
	}
	db.NewConnectionFunc = func(chatID int64, expiresAt time.Time) (int, error) { return 3, nil }
	http.AddPeerFunc = func(hostID int, DNS bool, telegramID int64) (*dto.Response, error) {
		return &dto.Response{Data: map[string]any{"public_key": "pub_key", "preshared_key": "psk"}}, nil
	}
	db.SaveKeysFunc = func(chatID int64, pubKey, psk string) error {
		return errors.New("db error")
	}
	http.DeletePeerFunc = func(publicKey string) error {
		deletePeerCalled = true
		if publicKey != "pub_key" {
			t.Errorf("expected pub_key, got %s", publicKey)
		}
		return nil
	}
	db.DeleteConnectionFunc = func(chatID int64) error {
		deleteConnectionCalled = true
		return nil
	}

	if err := svc.add(12345, 20000); err == nil {
		t.Fatal("expected error, got nil")
	}
	if !deletePeerCalled {
		t.Error("DeletePeer not called for rollback")
	}
	if !deleteConnectionCalled {
		t.Error("DeleteConnection not called for rollback")
	}
}

func TestAdd_DownloadConf_Error(t *testing.T) {
	svc, _, http, db := newTestService()

	db.GetPeerFunc = func(chatID int64) (string, string, error) {
		return "", "", dto.ErrNotFound
	}
	http.AddPeerFunc = func(hostID int, DNS bool, telegramID int64) (*dto.Response, error) {
		return &dto.Response{Data: map[string]any{"public_key": "k", "preshared_key": "p"}}, nil
	}
	db.NewConnectionFunc = func(chatID int64, expiresAt time.Time) (int, error) { return 3, nil }
	db.SaveKeysFunc = func(chatID int64, pubKey, psk string) error { return nil }
	http.DownloadConfFileFunc = func(telegramID int64) ([]byte, error) {
		return nil, errors.New("download failed")
	}

	if err := svc.add(12345, 20000); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// getConfFile()

func TestGetConfFile_NoSubscription(t *testing.T) {
	svc, tg, _, db := newTestService()

	db.CheckStatusFunc = func(chatID int64) bool { return false }

	updateSendTextCalled := false
	tg.UpdateSendTextFunc = func(update tgbotapi.Update, text string) error {
		updateSendTextCalled = true
		if text != "у вас нет подписки" {
			t.Errorf("unexpected message: %s", text)
		}
		return nil
	}

	if err := svc.getConfFile(callbackUpdate(12345)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updateSendTextCalled {
		t.Error("UpdateSendText not called")
	}
}

func TestGetConfFile_WithSubscription(t *testing.T) {
	svc, tg, http, db := newTestService()

	db.CheckStatusFunc = func(chatID int64) bool { return true }
	http.DownloadConfFileFunc = func(telegramID int64) ([]byte, error) {
		return []byte("config data"), nil
	}

	sendFileCalled := false
	tg.SendFileFunc = func(chatID int64, buffer []byte) error {
		sendFileCalled = true
		if string(buffer) != "config data" {
			t.Errorf("expected config data, got %s", string(buffer))
		}
		return nil
	}

	if err := svc.getConfFile(callbackUpdate(12345)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sendFileCalled {
		t.Error("SendFile not called")
	}
}

func TestGetConfFile_DownloadError(t *testing.T) {
	svc, _, http, db := newTestService()

	db.CheckStatusFunc = func(chatID int64) bool { return true }
	http.DownloadConfFileFunc = func(telegramID int64) ([]byte, error) {
		return nil, errors.New("download failed")
	}

	if err := svc.getConfFile(callbackUpdate(12345)); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// Invoice()

func TestInvoice(t *testing.T) {
	svc, tg, _, db := newTestService()

	payloadCaptured := ""
	db.NewPaymentFunc = func(chatID int64, payload string) error {
		payloadCaptured = payload
		return nil
	}
	tg.CreateAndSendInvoiceFunc = func(chatID int64, payload string) error {
		if payload != payloadCaptured {
			t.Errorf("payload mismatch: %s != %s", payload, payloadCaptured)
		}
		return nil
	}

	if err := svc.Invoice(callbackUpdate(12345)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payloadCaptured == "" {
		t.Error("expected non-empty payload")
	}
}

func TestInvoice_DBError(t *testing.T) {
	svc, _, _, db := newTestService()

	db.NewPaymentFunc = func(chatID int64, payload string) error {
		return errors.New("db error")
	}

	if err := svc.Invoice(callbackUpdate(12345)); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestInvoice_TelegramError(t *testing.T) {
	svc, tg, _, db := newTestService()

	db.NewPaymentFunc = func(chatID int64, payload string) error { return nil }
	tg.CreateAndSendInvoiceFunc = func(chatID int64, payload string) error {
		return errors.New("telegram error")
	}

	if err := svc.Invoice(callbackUpdate(12345)); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// CheckSubscription()

func TestCheckSubscription(t *testing.T) {
	svc, tg, http, db := newTestService()

	deletePeerCalled := false
	statusFalseCalled := false
	sendTextCalled := false

	db.ExpiredConnectionFunc = func() ([]dto.DelEntity, error) {
		return []dto.DelEntity{
			{ChatID: 100, PublicKey: "key1", PresharedKey: "psk1"},
			{ChatID: 200, PublicKey: "key2", PresharedKey: "psk2"},
		}, nil
	}
	http.DeletePeerFunc = func(publicKey string) error {
		deletePeerCalled = true
		if publicKey != "key1" && publicKey != "key2" {
			t.Errorf("unexpected public key: %s", publicKey)
		}
		return nil
	}
	db.StatusFalseFunc = func(chatID int64) error {
		statusFalseCalled = true
		return nil
	}
	tg.SendTextFunc = func(chatID int64, text string) error {
		sendTextCalled = true
		return nil
	}

	testLogger, closeFunc, err := logger.NewLogger()
	if err != nil {
		t.Fatalf("failed to create test logger: %v", err)
	}
	defer closeFunc()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.CheckSubcription(ctx, testLogger, 50*time.Millisecond)
	time.Sleep(150 * time.Millisecond)
	cancel()
	time.Sleep(50 * time.Millisecond)

	if !deletePeerCalled {
		t.Error("DeletePeer not called")
	}
	if !statusFalseCalled {
		t.Error("StatusFalse not called")
	}
	if !sendTextCalled {
		t.Error("SendText not called")
	}
}
