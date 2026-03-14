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
	AddClientFunc              func(username string, chatID int64) error
	TestedFunc                 func(chatID int64) error
	IsTestedFunc               func(chatID int64) bool
	StatusTrueFunc             func(chatID int64) error
	StatusFalseFunc            func(chatID int64) error
	CheckStatusFunc            func(chatID int64) bool
	NewPaymentFunc             func(chatID int64, payload string) error
	SuccessfulPaymentStatusFunc func(payload string) error
	NewConnectionFunc          func(chatID int64, expires_at time.Time) error
	SaveKeyFunc                func(chatID int64, publicKey string) error
	ExpiredConnectionFunc      func() ([]dto.DelEntity, error)
	GetHostIDFunc              func(chatID int64) (int, error)
	CloseFunc                  func() error
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

func (m *mockPostgres) NewConnection(chatID int64, expires_at time.Time) error {
	if m.NewConnectionFunc != nil {
		return m.NewConnectionFunc(chatID, expires_at)
	}
	return nil
}

func (m *mockPostgres) SaveKey(chatID int64, publicKey string) error {
	if m.SaveKeyFunc != nil {
		return m.SaveKeyFunc(chatID, publicKey)
	}
	return nil
}

func (m *mockPostgres) ExpiredConnection() ([]dto.DelEntity, error) {
	if m.ExpiredConnectionFunc != nil {
		return m.ExpiredConnectionFunc()
	}
	return nil, nil
}

func (m *mockPostgres) GetHostID(chatID int64) (int, error) {
	if m.GetHostIDFunc != nil {
		return m.GetHostIDFunc(chatID)
	}
	return 0, nil
}

func (m *mockPostgres) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

type mockTelegramClient struct {
	ChanFunc                   func() tgbotapi.UpdatesChannel
	MenuFunc                   func(chatID int64) error
	UpdateMainMenuFunc         func(update tgbotapi.Update) error
	UpdateSendTextFunc         func(update tgbotapi.Update, text string) error
	SendTextFunc               func(chatID int64, text string) error
	SendFileFunc               func(chatID int64, buffer []byte) error
	CreateAndSendInvoiceFunc   func(chatID int64, payload string) error
	PreCheckoutQueryFunc       func(update tgbotapi.Update) error
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

// Tests

func TestNewService(t *testing.T) {
	mockTg := &mockTelegramClient{}
	mockHttp := &mockHTTPClient{}
	mockDb := &mockPostgres{}

	svc := New(mockTg, mockHttp, mockDb)

	if svc.telegram != mockTg {
		t.Error("telegram client not set")
	}
	if svc.httpClient != mockHttp {
		t.Error("http client not set")
	}
	if svc.postgres != mockDb {
		t.Error("postgres not set")
	}
}

func TestAdd_WithPaid30Days(t *testing.T) {
	mockTg := &mockTelegramClient{}
	mockHttp := &mockHTTPClient{}
	mockDb := &mockPostgres{}

	dateCalled := false
	hostIDCalled := false
	keyWasSaved := false

	mockDb.NewConnectionFunc = func(chatID int64, expiresAt time.Time) error {
		dateCalled = true
		now := time.Now()
		expectedDate := now.Add(30 * 24 * time.Hour)
		diff := expiresAt.Sub(expectedDate).Abs()
		if diff > time.Minute {
			t.Errorf("expected ~30 days, got %v", expiresAt.Sub(now))
		}
		return nil
	}

	mockDb.GetHostIDFunc = func(chatID int64) (int, error) {
		hostIDCalled = true
		if chatID != 12345 {
			t.Errorf("expected chatID 12345, got %d", chatID)
		}
		return 5, nil
	}

	mockHttp.AddPeerFunc = func(hostID int, DNS bool, telegramID int64) (*dto.Response, error) {
		if hostID != 5 {
			t.Errorf("expected hostID 5, got %d", hostID)
		}
		if !DNS {
			t.Error("expected DNS to be true")
		}
		return &dto.Response{
			Data: map[string]any{"public_key": "test_key"},
		}, nil
	}

	mockDb.SaveKeyFunc = func(chatID int64, publicKey string) error {
		keyWasSaved = true
		if publicKey != "test_key" {
			t.Errorf("expected test_key, got %s", publicKey)
		}
		return nil
	}

	mockHttp.DownloadConfFileFunc = func(telegramID int64) ([]byte, error) {
		return []byte("config data"), nil
	}

	mockTg.SendFileFunc = func(chatID int64, buffer []byte) error {
		if string(buffer) != "config data" {
			t.Errorf("expected config data, got %s", string(buffer))
		}
		return nil
	}

	svc := New(mockTg, mockHttp, mockDb)
	err := svc.add(12345, 20000)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !dateCalled {
		t.Error("NewConnection not called")
	}
	if !hostIDCalled {
		t.Error("GetHostID not called")
	}
	if !keyWasSaved {
		t.Error("SaveKey not called")
	}
}

func TestAdd_WithTestAccess24h(t *testing.T) {
	mockTg := &mockTelegramClient{}
	mockHttp := &mockHTTPClient{}
	mockDb := &mockPostgres{}

	mockDb.NewConnectionFunc = func(chatID int64, expiresAt time.Time) error {
		now := time.Now()
		expectedDate := now.Add(24 * time.Hour)
		diff := expiresAt.Sub(expectedDate).Abs()
		if diff > time.Minute {
			t.Errorf("expected ~24 hours, got %v", expiresAt.Sub(now))
		}
		return nil
	}

	mockDb.GetHostIDFunc = func(chatID int64) (int, error) {
		return 1, nil
	}

	mockHttp.AddPeerFunc = func(hostID int, DNS bool, telegramID int64) (*dto.Response, error) {
		return &dto.Response{
			Data: map[string]any{"public_key": "test_key"},
		}, nil
	}

	mockDb.SaveKeyFunc = func(chatID int64, publicKey string) error {
		return nil
	}

	mockHttp.DownloadConfFileFunc = func(telegramID int64) ([]byte, error) {
		return []byte("config"), nil
	}

	mockTg.SendFileFunc = func(chatID int64, buffer []byte) error {
		return nil
	}

	svc := New(mockTg, mockHttp, mockDb)
	err := svc.add(12345, 0)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAdd_DBError(t *testing.T) {
	mockTg := &mockTelegramClient{}
	mockHttp := &mockHTTPClient{}
	mockDb := &mockPostgres{}

	mockDb.NewConnectionFunc = func(chatID int64, expiresAt time.Time) error {
		return errors.New("db connection error")
	}

	svc := New(mockTg, mockHttp, mockDb)
	err := svc.add(12345, 20000)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAdd_GetHostIDError(t *testing.T) {
	mockTg := &mockTelegramClient{}
	mockHttp := &mockHTTPClient{}
	mockDb := &mockPostgres{}

	mockDb.NewConnectionFunc = func(chatID int64, expiresAt time.Time) error {
		return nil
	}

	mockDb.GetHostIDFunc = func(chatID int64) (int, error) {
		return 0, errors.New("host not found")
	}

	svc := New(mockTg, mockHttp, mockDb)
	err := svc.add(12345, 20000)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAdd_HTTPError(t *testing.T) {
	mockTg := &mockTelegramClient{}
	mockHttp := &mockHTTPClient{}
	mockDb := &mockPostgres{}

	mockDb.NewConnectionFunc = func(chatID int64, expiresAt time.Time) error {
		return nil
	}

	mockDb.GetHostIDFunc = func(chatID int64) (int, error) {
		return 5, nil
	}

	mockHttp.AddPeerFunc = func(hostID int, DNS bool, telegramID int64) (*dto.Response, error) {
		return nil, errors.New("http error")
	}

	svc := New(mockTg, mockHttp, mockDb)
	err := svc.add(12345, 20000)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetConfFile_NoSubscription(t *testing.T) {
	mockTg := &mockTelegramClient{}
	mockHttp := &mockHTTPClient{}
	mockDb := &mockPostgres{}

	mockDb.CheckStatusFunc = func(chatID int64) bool {
		return false
	}

	updateSendTextCalled := false
	mockTg.UpdateSendTextFunc = func(update tgbotapi.Update, text string) error {
		updateSendTextCalled = true
		if text != "у вас нет подписки" {
			t.Errorf("expected no subscription message, got %s", text)
		}
		return nil
	}

	svc := New(mockTg, mockHttp, mockDb)

	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 12345},
			},
		},
	}

	err := svc.getConfFile(update)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updateSendTextCalled {
		t.Error("UpdateSendText not called")
	}
}

func TestGetConfFile_WithSubscription(t *testing.T) {
	mockTg := &mockTelegramClient{}
	mockHttp := &mockHTTPClient{}
	mockDb := &mockPostgres{}

	mockDb.CheckStatusFunc = func(chatID int64) bool {
		if chatID != 12345 {
			t.Errorf("expected chatID 12345, got %d", chatID)
		}
		return true
	}

	sendFileCalled := false
	mockHttp.DownloadConfFileFunc = func(telegramID int64) ([]byte, error) {
		if telegramID != 12345 {
			t.Errorf("expected telegramID 12345, got %d", telegramID)
		}
		return []byte("config data"), nil
	}

	mockTg.SendFileFunc = func(chatID int64, buffer []byte) error {
		sendFileCalled = true
		if chatID != 12345 {
			t.Errorf("expected chatID 12345, got %d", chatID)
		}
		if string(buffer) != "config data" {
			t.Errorf("expected config data, got %s", string(buffer))
		}
		return nil
	}

	svc := New(mockTg, mockHttp, mockDb)

	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 12345},
			},
		},
	}

	err := svc.getConfFile(update)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sendFileCalled {
		t.Error("SendFile not called")
	}
}

func TestGetConfFile_DownloadError(t *testing.T) {
	mockTg := &mockTelegramClient{}
	mockHttp := &mockHTTPClient{}
	mockDb := &mockPostgres{}

	mockDb.CheckStatusFunc = func(chatID int64) bool {
		return true
	}

	mockHttp.DownloadConfFileFunc = func(telegramID int64) ([]byte, error) {
		return nil, errors.New("download failed")
	}

	svc := New(mockTg, mockHttp, mockDb)

	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 12345},
			},
		},
	}

	err := svc.getConfFile(update)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestInvoice(t *testing.T) {
	mockTg := &mockTelegramClient{}
	mockHttp := &mockHTTPClient{}
	mockDb := &mockPostgres{}

	payloadCaptured := ""
	mockDb.NewPaymentFunc = func(chatID int64, payload string) error {
		if chatID != 12345 {
			t.Errorf("expected chatID 12345, got %d", chatID)
		}
		payloadCaptured = payload
		return nil
	}

	mockTg.CreateAndSendInvoiceFunc = func(chatID int64, payload string) error {
		if chatID != 12345 {
			t.Errorf("expected chatID 12345, got %d", chatID)
		}
		if payload != payloadCaptured {
			t.Errorf("expected payload %s, got %s", payloadCaptured, payload)
		}
		return nil
	}

	svc := New(mockTg, mockHttp, mockDb)

	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 12345},
			},
		},
	}

	err := svc.Invoice(update)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that a UUID was generated (non-empty payload)
	if payloadCaptured == "" {
		t.Error("expected non-empty payload")
	}
}

func TestInvoice_DBError(t *testing.T) {
	mockTg := &mockTelegramClient{}
	mockHttp := &mockHTTPClient{}
	mockDb := &mockPostgres{}

	mockDb.NewPaymentFunc = func(chatID int64, payload string) error {
		return errors.New("db error")
	}

	svc := New(mockTg, mockHttp, mockDb)

	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 12345},
			},
		},
	}

	err := svc.Invoice(update)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestInvoice_TelegramError(t *testing.T) {
	mockTg := &mockTelegramClient{}
	mockHttp := &mockHTTPClient{}
	mockDb := &mockPostgres{}

	mockDb.NewPaymentFunc = func(chatID int64, payload string) error {
		return nil
	}

	mockTg.CreateAndSendInvoiceFunc = func(chatID int64, payload string) error {
		return errors.New("telegram error")
	}

	svc := New(mockTg, mockHttp, mockDb)

	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 12345},
			},
		},
	}

	err := svc.Invoice(update)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCheckSubscription(t *testing.T) {
	mockTg := &mockTelegramClient{}
	mockHttp := &mockHTTPClient{}
	mockDb := &mockPostgres{}

	deletePeerCalled := false
	statusFalseCalled := false
	sendTextCalled := false

	mockDb.ExpiredConnectionFunc = func() ([]dto.DelEntity, error) {
		return []dto.DelEntity{
			{ChatID: 100, PublicKey: "key1"},
			{ChatID: 200, PublicKey: "key2"},
		}, nil
	}

	mockHttp.DeletePeerFunc = func(publicKey string) error {
		deletePeerCalled = true
		if publicKey != "key1" && publicKey != "key2" {
			t.Errorf("unexpected public key: %s", publicKey)
		}
		return nil
	}

	mockDb.StatusFalseFunc = func(chatID int64) error {
		statusFalseCalled = true
		if chatID != 100 && chatID != 200 {
			t.Errorf("unexpected chatID: %d", chatID)
		}
		return nil
	}

	mockTg.SendTextFunc = func(chatID int64, text string) error {
		sendTextCalled = true
		if chatID != 100 && chatID != 200 {
			t.Errorf("unexpected chatID: %d", chatID)
		}
		return nil
	}

	svc := New(mockTg, mockHttp, mockDb)

	// Create a test logger
	testLogger, closeFunc, err := logger.NewLogger()
	if err != nil {
		t.Fatalf("failed to create test logger: %v", err)
	}
	defer closeFunc()

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run with very short duration for testing
	go svc.CheckSubcription(ctx, testLogger, 50*time.Millisecond)

	// Let it run once
	time.Sleep(150 * time.Millisecond)
	cancel()

	// Give goroutine time to exit
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
