// Package telegram содержит реализацию Telegram бота для сервиса GopherSecure.
// Бот управляет взаимодействием с пользователями через Telegram Bot API.
package telegram

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegram-service/internal/dto"
)

// Telegram представляет Telegram бота с функциями управления меню и процессом оплаты.
type Telegram struct {
	bot           *tgbotapi.BotAPI
	updates       tgbotapi.UpdatesChannel
	providerToken string
}

// New создает новый экземпляр Telegram бота с указанными токеном и токеном провайдера платежей.
// Возвращает инициализированного бота готового к обработке обновлений.
func New(telegramToken, providerToken string) (*Telegram, error) {
	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		return nil, err
	}
	telegram := &Telegram{
		bot:           bot,
		providerToken: providerToken,
	}

	u := tgbotapi.NewUpdate(0)
	telegram.updates = bot.GetUpdatesChan(u)

	// меню слева внизу с командами
	commands := []tgbotapi.BotCommand{
		{Command: "menu", Description: "Меню"},
	}
	_, err = telegram.bot.Request(tgbotapi.NewSetMyCommands(commands...))
	if err != nil {
		return nil, err
	}
	return telegram, nil
}

// Chan возвращает канал обновлений от Telegram Bot API.
func (t *Telegram) Chan() tgbotapi.UpdatesChannel {
	return t.updates
}

// keyboardMainMenu создает клавиатуру с кнопками главного меню.
// Включает опции для получения конфига, помощи, тестирования, стоимости и оплаты.
func keyboardMainMenu() tgbotapi.InlineKeyboardMarkup {
	options := []string{"получить конфиг", "помощь", "протестировать", "стоимость", "оплатить"}

	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(options))
	for _, opt := range options {
		btn := tgbotapi.NewInlineKeyboardButtonData(opt, dto.EncodeCallbackData(opt))
		row := tgbotapi.NewInlineKeyboardRow(btn)
		rows = append(rows, row)
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// keyboardBackMenu создает клавиатуру с кнопкой возврата в главное меню.
func keyboardBackMenu() tgbotapi.InlineKeyboardMarkup {
	opt := "<- назад"
	btn := tgbotapi.NewInlineKeyboardButtonData(opt, dto.EncodeCallbackData(opt))
	row := tgbotapi.NewInlineKeyboardRow(btn)
	return tgbotapi.NewInlineKeyboardMarkup(row)
}

// Menu отправляет новое сообщение с главным меню пользователю.
func (t *Telegram) Menu(chatID int64) error {

	msg := tgbotapi.NewMessage(chatID, "📱 Главное меню")
	msg.ReplyMarkup = keyboardMainMenu()

	_, err := t.bot.Send(msg)
	return err
}

// UpdateMainMenu редактирует текущее сообщение и выводит главное меню.
func (t *Telegram) UpdateMainMenu(update tgbotapi.Update) error {

	msg := tgbotapi.NewEditMessageTextAndMarkup(
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		"📱 главное меню",
		keyboardMainMenu(),
	)

	_, err := t.bot.Send(msg)
	return err
}

// UpdateSendText редактирует текущее сообщение на заданный текст с кнопкой возврата.
func (t *Telegram) UpdateSendText(update tgbotapi.Update, text string) error {
	msg := tgbotapi.NewEditMessageTextAndMarkup(
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		text,
		keyboardBackMenu(),
	)

	_, err := t.bot.Send(msg)

	return err
}

// SendFile отправляет конфигурационный файл пользователю.
// Файл назначается с именем формата awg{timestamp}.conf.
func (t *Telegram) SendFile(chatID int64, bufer []byte) error {
	// create document struct
	unix := time.Now().Unix()
	file := tgbotapi.FileBytes{
		Name:  fmt.Sprintf("awg%d.conf", unix),
		Bytes: bufer,
	}
	msg := tgbotapi.NewDocument(chatID, file)
	_, err := t.bot.Send(msg)
	return err
}

// SendText отправляет текстовое сообщение пользователю.
func (t *Telegram) SendText(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := t.bot.Send(msg)
	return err
}
