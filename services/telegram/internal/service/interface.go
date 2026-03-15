package service

import (
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegram-service/internal/dto"
)

type postgres interface {
	// postgres methods
	Close() error

	// clients methods
	AddClient(username string, chatID int64) error

	Tested(chatID int64) error
	IsTested(chatID int64) bool

	StatusTrue(chatID int64) error
	StatusFalse(chatID int64) error
	CheckStatus(chatID int64) bool

	// payments methods
	NewPayment(chatID int64, payload string) error
	SuccessfulPaymentStatus(payload string) error

	// peers methods
	NewConnection(chatID int64, expiresAt time.Time) (int, error)
	DeleteConnection(chatID int64) error
	RenewConnection(chatID int64, expiresAt time.Time) error
	GetPeer(chatID int64) (publicKey, presharedKey string, err error)
	SaveKeys(chatID int64, pubKey, psk string) error
	ExpiredConnection() ([]dto.DelEntity, error)
}

type telegramClient interface {
	// Chan возвращает канал обновлений (от tgbotapi.UpdatesChannel)
	Chan() tgbotapi.UpdatesChannel
	// Menu отправляет сообщение с главным меню
	Menu(chatID int64) error
	// UpdateMainMenu меняет сообщение на главном меню
	UpdateMainMenu(update tgbotapi.Update) error
	// UpdateSendText меняет текст сообщения и ставит меню "назад"
	UpdateSendText(update tgbotapi.Update, text string) error
	// SendText отправляет текстовое сообщение пользователю
	SendText(chatID int64, text string) error
	// SendFile отправляет файл (конфиг) пользователю
	SendFile(chatID int64, buffer []byte) error
	// создает кнопку оплаты
	CreateAndSendInvoice(chatID int64, payload string) error
	// запрос перед оплатой
	// на него нужно ответить в течение 10 секунд
	PreCheckoutQuery(update tgbotapi.Update) error
	// handler, успешная оплата
	// отправляет успешный результат пользователю
	HandleSuccessfulPayment(update tgbotapi.Update) (*dto.PaymentHandler, error)
}

type httpClient interface {
	AddPeer(hostID int, DNS bool, telegramID int64) (*dto.Response, error)
	DeletePeer(publicKey string) error
	DownloadConfFile(telegramID int64) ([]byte, error)
}

type service struct {
	telegram   telegramClient
	httpClient httpClient
	postgres   postgres
}

func New(telegram telegramClient, httpClient httpClient, postgres postgres) service {

	return service{
		telegram:   telegram,
		httpClient: httpClient,
		postgres:   postgres,
	}
}
