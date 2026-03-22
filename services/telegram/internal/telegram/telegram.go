// Package telegram содержит реализацию Telegram бота для сервиса GopherSecure.
// Бот управляет взаимодействием с пользователями через Telegram Bot API.
package telegram

import (
	"fmt"
	"telegram-service/internal/dto"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const MENU = "   🌐 GopherSecure VPN "

var actionLabels = map[string]string{
	dto.ActionBack:    "↩️ Назад",
	dto.ActionTrial:   "⚡ Попробовать бесплатно",
	dto.ActionPricing: "💰 Стоимость",
	dto.ActionPay:     "💳 Купить подписку",
	dto.ActionConfig:  "📦 Мой конфиг",
	dto.ActionHelp:    "❓ Помощь",
}

var commands = []tgbotapi.BotCommand{
	{Command: "start", Description: "Начать работу с ботом"},
	{Command: "menu", Description: "Меню"},
	{Command: "promo", Description: "🎁 Промокод"},
}

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
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(actionLabels[dto.ActionTrial], dto.ActionTrial),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(actionLabels[dto.ActionPay], dto.ActionPay),
			tgbotapi.NewInlineKeyboardButtonData(actionLabels[dto.ActionPricing], dto.ActionPricing),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(actionLabels[dto.ActionConfig], dto.ActionConfig),
			tgbotapi.NewInlineKeyboardButtonData(actionLabels[dto.ActionHelp], dto.ActionHelp),
		),
	)
}

// keyboardBackMenu создает клавиатуру с кнопкой возврата в главное меню.
func keyboardBackMenu() tgbotapi.InlineKeyboardMarkup {
	btn := tgbotapi.NewInlineKeyboardButtonData(actionLabels[dto.ActionBack], dto.ActionBack)
	row := tgbotapi.NewInlineKeyboardRow(btn)
	return tgbotapi.NewInlineKeyboardMarkup(row)
}

// Menu отправляет новое сообщение с главным меню пользователю.
func (t *Telegram) Menu(chatID int64) error {

	msg := tgbotapi.NewMessage(chatID, MENU)
	msg.ReplyMarkup = keyboardMainMenu()

	_, err := t.bot.Send(msg)
	return err
}

// UpdateMainMenu редактирует текущее сообщение и выводит главное меню.
func (t *Telegram) UpdateMainMenu(update tgbotapi.Update) error {

	msg := tgbotapi.NewEditMessageTextAndMarkup(
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		MENU,
		keyboardMainMenu(),
	)

	_, err := t.bot.Send(msg)
	return err
}

// UpdateSendTextWithBackAction редактирует текущее сообщение на заданный текст с кнопкой возврата.
func (t *Telegram) UpdateSendTextWithBackAction(update tgbotapi.Update, text string) error {
	msg := tgbotapi.NewEditMessageTextAndMarkup(
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		text,
		keyboardBackMenu(),
	)

	_, err := t.bot.Send(msg)

	return err
}

// UpdateSendText редактирует текущее сообщение на заданный текст без кнопки возврата.
func (t *Telegram) UpdateSendText(update tgbotapi.Update, text string) error {
	msg := tgbotapi.NewEditMessageText(
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		text,
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
