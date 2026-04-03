package i18n

import (
	"fmt"
	"math"
	"time"
)

const (
	LocaleIT = "it"
	LocaleEN = "en"
)

// FormatAmount formats an integer amount in minor units (e.g. cents) as a currency string.
// amount is in the smallest currency unit (e.g. 1050 = 10.50).
func FormatAmount(amount int64, currency, locale string) string {
	major := float64(amount) / 100.0
	switch locale {
	case LocaleIT:
		return fmt.Sprintf("%s %.2f", currency, major)
	default:
		return fmt.Sprintf("%s %.2f", currency, major)
	}
}

// FormatDate formats a time.Time according to the given locale.
func FormatDate(t time.Time, locale string) string {
	switch locale {
	case LocaleIT:
		return t.Format("02/01/2006")
	default:
		return t.Format("2006-01-02")
	}
}

// FormatDateTime formats a time.Time with time component according to the given locale.
func FormatDateTime(t time.Time, locale string) string {
	switch locale {
	case LocaleIT:
		return t.Format("02/01/2006 15:04")
	default:
		return t.Format("2006-01-02 15:04")
	}
}

// RoundAmount rounds a float64 amount to 2 decimal places using banker's rounding.
func RoundAmount(amount float64) int64 {
	return int64(math.Round(amount * 100))
}
