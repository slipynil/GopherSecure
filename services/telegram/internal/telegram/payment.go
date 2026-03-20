package telegram

import (
	"telegram-service/internal/dto"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CreateAndSendInvoice создает и отправляет счет на оплату пользователю.
// Стоимость подписки составляет 20000 копеек (200 рублей) на 30 дней.
func (t *Telegram) CreateAndSendInvoice(chatID int64, payload string) error {

	title := "Оплата услуги"
	description := "Доступ на 30 дней"
	currency := "RUB"
	basePrice := tgbotapi.LabeledPrice{
		Label:  "Оплата",
		Amount: 20000,
	}

	invoice := tgbotapi.NewInvoice(
		chatID, title, description,
		payload, t.providerToken, "test",
		currency, []tgbotapi.LabeledPrice{basePrice},
	)
	invoice.SuggestedTipAmounts = []int{}
	invoice.MaxTipAmount = 0
	_, err := t.bot.Send(invoice)
	return err
}

// PreCheckoutQuery обрабатывает запрос до оплаты от Telegram.
// Требует ответа в течение 10 секунд и подтверждает готовность принять платеж.
func (t *Telegram) PreCheckoutQuery(update tgbotapi.Update) error {
	query := update.PreCheckoutQuery
	// Говорим Телеграму, что готовы принять платеж
	// Можно добавить проверки: есть ли товар в наличии, доступен ли пользователь и т.д.
	// Если всё ок — отвечаем ok=true
	answer := tgbotapi.PreCheckoutConfig{
		PreCheckoutQueryID: query.ID,
		OK:                 true,
	}
	_, err := t.bot.Request(answer)
	return err
}

// HandleSuccessfulPayment обрабатывает успешный платеж и отправляет подтверждение пользователю.
// Возвращает структуру с деталями платежа для дальнейшей обработки.
func (t *Telegram) HandleSuccessfulPayment(update tgbotapi.Update) (*dto.PaymentHandler, error) {
	payment := update.Message.SuccessfulPayment
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "✅ Оплата прошла успешно! Услуга активирована.")
	_, err := t.bot.Send(msg)
	if err != nil {
		return nil, err
	}
	paymentHandler := dto.PaymentHandler{
		InvoicePayload: payment.InvoicePayload,
		TotalAmount:    payment.TotalAmount,
		Currency:       payment.Currency,
	}
	return &paymentHandler, nil
}
