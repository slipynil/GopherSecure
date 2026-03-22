package main

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"cli-admins/internal/client"
)

var ldflagsAddr = "0.0.0.0:8080"

func main() {
	// Получить ADDRESS из переменной окружения или использовать default
	address := ldflagsAddr

	promoClient := client.NewPromoClient(address)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "create":
		handleCreate(promoClient, os.Args[2:])

	case "update":
		handleUpdate(promoClient, os.Args[2:])

	case "list":
		handleList(promoClient)

	case "delete":
		handleDelete(promoClient, os.Args[2:])

	default:
		fmt.Printf("❌ Неизвестная команда: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleCreate(promoClient *client.PromoClient, args []string) {
	if len(args) < 4 {
		fmt.Println("❌ Использование: cli-admins create <code> <bonus_days> <max_uses> <expires_at>")
		fmt.Println("Пример: cli-admins create BONUS30 30 100 2026-03-29T23:59:59Z")
		os.Exit(1)
	}

	code := args[0]
	bonusDays, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Printf("❌ bonus_days должно быть числом: %v\n", err)
		os.Exit(1)
	}

	maxUses, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Printf("❌ max_uses должно быть числом: %v\n", err)
		os.Exit(1)
	}

	expiresAt := args[3]

	// Проверить формат даты
	if _, err := time.Parse(time.RFC3339, expiresAt); err != nil {
		fmt.Printf("❌ Неверный формат даты. Используйте RFC3339 (2026-03-29T23:59:59Z): %v\n", err)
		os.Exit(1)
	}

	req := client.CreatePromoRequest{
		Code:      code,
		BonusDays: bonusDays,
		MaxUses:   maxUses,
		ExpiresAt: expiresAt,
	}

	result, err := promoClient.CreatePromo(req)
	if err != nil {
		fmt.Printf("❌ Ошибка: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Промокод успешно создан:\n")
	printJSON(result)
}

func handleUpdate(promoClient *client.PromoClient, args []string) {
	if len(args) < 4 {
		fmt.Println("❌ Использование: cli-admins update <id> <bonus_days> <max_uses> <expires_at>")
		fmt.Println("Пример: cli-admins update 1 60 200 2026-03-29T23:59:59Z")
		os.Exit(1)
	}

	promoID, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("❌ id должно быть числом: %v\n", err)
		os.Exit(1)
	}

	bonusDays, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Printf("❌ bonus_days должно быть числом: %v\n", err)
		os.Exit(1)
	}

	maxUses, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Printf("❌ max_uses должно быть числом: %v\n", err)
		os.Exit(1)
	}

	expiresAt := args[3]

	// Проверить формат даты
	if _, err := time.Parse(time.RFC3339, expiresAt); err != nil {
		fmt.Printf("❌ Неверный формат даты. Используйте RFC3339 (2026-03-29T23:59:59Z): %v\n", err)
		os.Exit(1)
	}

	req := client.UpdatePromoRequest{
		BonusDays: bonusDays,
		MaxUses:   maxUses,
		ExpiresAt: expiresAt,
	}

	result, err := promoClient.UpdatePromo(promoID, req)
	if err != nil {
		fmt.Printf("❌ Ошибка: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Промокод #%d успешно обновлен:\n", promoID)
	printJSON(result)
}

func handleList(promoClient *client.PromoClient) {
	promos, err := promoClient.ListPromos()
	if err != nil {
		fmt.Printf("❌ Ошибка: %v\n", err)
		os.Exit(1)
	}

	if len(promos) == 0 {
		fmt.Println("📭 Промокодов не найдено")
		return
	}

	fmt.Println("\n📋 Все промокоды:")
	fmt.Println("─────────────────────────────────────────────────────────────────────────────────")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tКОД\tДНИ\tМАКС\tИСП\tАКТИВ\tИСТЕКАЕТ")

	for _, promo := range promos {
		id := fmt.Sprintf("%v", promo["id"])
		code := fmt.Sprintf("%v", promo["code"])
		bonusDays := fmt.Sprintf("%v", promo["bonus_days"])
		maxUses := fmt.Sprintf("%v", promo["max_uses"])
		usedCount := fmt.Sprintf("%v", promo["used_count"])
		isActive := fmt.Sprintf("%v", promo["is_active"])
		expiresAt := fmt.Sprintf("%v", promo["expires_at"])

		// Форматировать дату
		if expiresAtStr, ok := promo["expires_at"].(string); ok {
			if t, err := time.Parse(time.RFC3339, expiresAtStr); err == nil {
				expiresAt = t.Format("2006-01-02")
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", id, code, bonusDays, maxUses, usedCount, isActive, expiresAt)
	}

	w.Flush()
	fmt.Println("─────────────────────────────────────────────────────────────────────────────────")
}

func handleDelete(promoClient *client.PromoClient, args []string) {
	if len(args) < 1 {
		fmt.Println("❌ Использование: cli-admins delete <id>")
		fmt.Println("Пример: cli-admins delete 1")
		os.Exit(1)
	}

	promoID, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("❌ id должно быть числом: %v\n", err)
		os.Exit(1)
	}

	result, err := promoClient.DeletePromo(promoID)
	if err != nil {
		fmt.Printf("❌ Ошибка: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Промокод #%d успешно удален:\n", promoID)
	printJSON(result)
}

func printJSON(data map[string]any) {
	for key, value := range data {
		fmt.Printf("  %s: %v\n", key, value)
	}
}

func printUsage() {
	fmt.Println(`
GopherSecure Admin CLI - управление промокодами

Использование:
  cli-admins <command> [arguments]

Команды:
  create <code> <bonus_days> <max_uses> <expires_at>
    Создать новый промокод
    Пример: cli-admins create BONUS30 30 100 2026-03-29T23:59:59Z

  update <id> <bonus_days> <max_uses> <expires_at>
    Обновить параметры промокода
    Пример: cli-admins update 1 60 200 2026-03-29T23:59:59Z

  list
    Показать список всех промокодов

  delete <id>
    Удалить (деактивировать) промокод
    Пример: cli-admins delete 1

Переменные окружения:
  ADDRESS - адрес телеграм сервиса (по умолчанию: localhost:8080)
  Пример: ADDRESS=api.example.com:8080 cli-admins list
`)
}
