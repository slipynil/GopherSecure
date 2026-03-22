package promocode

import "fmt"

// DuplicatePromoCodeError — ошибка при попытке создать промокод, который уже существует.
type DuplicatePromoCodeError struct {
	Code string
}

// Error реализует интерфейс error.
func (e *DuplicatePromoCodeError) Error() string {
	return fmt.Sprintf("промокод '%s' уже был создан", e.Code)
}

// IsDuplicatePromoCodeError проверяет, является ли ошибка ошибкой дубликата.
func IsDuplicatePromoCodeError(err error) bool {
	_, ok := err.(*DuplicatePromoCodeError)
	return ok
}
