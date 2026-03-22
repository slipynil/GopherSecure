// Package service предоставляет основную бизнес-логику для управления VPN подписками и взаимодействием с клиентами через Telegram.
package service

import (
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegram-service/internal/dto"
)

// postgres определяет интерфейс для доступа к базе данных PostgreSQL.
// Методы интерфейса покрывают операции с клиентами, платежами и подключениями WireGuard.
type postgres interface {
	// postgres methods
	Close() error

	// clients methods
	AddClient(username string, chatID int64) error

	Tested(chatID int64) error
	IsTested(chatID int64) (bool, error)

	StatusTrue(chatID int64) error
	StatusFalse(chatID int64) error
	CheckStatus(chatID int64) (bool, error)

	// payments methods
	NewPayment(chatID int64, payload string) error
	SuccessfulPaymentStatus(payload string) error

	// peers methods
	NewConnection(chatID int64) (int, error)
	SaveConnection(hostID int, publicKey, presharedKey string, expiresAt time.Time) error
	DeleteConnection(hostID int) error
	MarkExpired(hostID int) error
	GetConnection(chatID int64) (*dto.DelEntity, error)
	ExpiredConnection() ([]dto.DelEntity, error)
	GetHostID(chatID int64) (int, error)
}

// telegramClient определяет интерфейс для взаимодействия с Telegram Bot API.
// Методы интерфейса покрывают отправку сообщений, управление меню и обработку платежей.
type telegramClient interface {
	// Chan возвращает канал обновлений (от tgbotapi.UpdatesChannel)
	Chan() tgbotapi.UpdatesChannel
	// Menu отправляет сообщение с главным меню
	Menu(chatID int64) error
	// UpdateMainMenu меняет сообщение на главном меню
	UpdateMainMenu(update tgbotapi.Update) error
	// UpdateSendTextWithBackAction меняет текст сообщения и ставит меню "назад"
	UpdateSendTextWithBackAction(update tgbotapi.Update, text string) error
	// UpdateSendText меняет текст сообщения
	UpdateSendText(update tgbotapi.Update, text string) error
	// SendText отправляет текстовое сообщение пользователю
	SendText(chatID int64, text string) error
	// SendFile отправляет файл (конфиг) пользователю
	SendFile(chatID int64, buffer []byte) error
	// CreateAndSendInvoice создает кнопку оплаты.
	CreateAndSendInvoice(chatID int64, payload string) error
	// PreCheckoutQuery обрабатывает запрос перед оплатой.
	// На него нужно ответить в течение 10 секунд.
	PreCheckoutQuery(update tgbotapi.Update) error
	// HandleSuccessfulPayment обрабатывает успешный платеж и отправляет результат пользователю.
	HandleSuccessfulPayment(update tgbotapi.Update) (*dto.PaymentHandler, error)
}

// httpClient определяет интерфейс для взаимодействия с HTTP API сервиса AWG.
// Методы интерфейса покрывают управление WireGuard пирами и получение конфигураций.
type httpClient interface {
	AddPeer(hostID int, DNS bool, telegramID int64) (*dto.Response, error)
	DeletePeer(publicKey string) error
	RestorePeer(publicKey, presharedKey, socket string, telegramID int64) error
	DownloadConfFile(telegramID int64) ([]byte, error)
}

// service представляет основную сервисную структуру, которая координирует взаимодействие
// между Telegram клиентом, HTTP клиентом и базой данных.
type service struct {
	telegram   telegramClient
	httpClient httpClient
	postgres   postgres
}

// New создает новый экземпляр [service] с предоставленными телеграм клиентом, HTTP клиентом и базой данных.
func New(telegram telegramClient, httpClient httpClient, postgres postgres) service {

	return service{
		telegram:   telegram,
		httpClient: httpClient,
		postgres:   postgres,
	}
}
