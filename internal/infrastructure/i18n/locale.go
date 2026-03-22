package i18n

import (
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Locale struct {
	printer *message.Printer
	tag     language.Tag
}

func New(locale string) *Locale {
	tag := language.Make(locale)
	return &Locale{
		printer: message.NewPrinter(tag),
		tag:     tag,
	}
}

// FormatCurrency formats a currency amount with the appropriate symbol and separators.
// For pt-BR: R$ 1.234,56
// For en-US: $1,234.56
func (l *Locale) FormatCurrency(amount float64, currencyCode string) string {
	symbol := currencySymbol(currencyCode)
	return l.printer.Sprintf("%s %.2f", symbol, amount)
}

// FormatNumber formats a number with the appropriate decimal and thousands separators.
// For pt-BR: 1.234,56
// For en-US: 1,234.56
func (l *Locale) FormatNumber(n float64) string {
	return l.printer.Sprintf("%.2f", n)
}

// FormatDate formats a date in dd/mm/YYYY format.
func (l *Locale) FormatDate(t time.Time) string {
	return t.Format("02/01/2006")
}

// FormatDateShort formats a date in dd-mm-YYYY format.
func (l *Locale) FormatDateShort(t time.Time) string {
	return t.Format("02-01-2006")
}

func currencySymbol(code string) string {
	switch code {
	case "BRL":
		return "R$"
	case "USD":
		return "$"
	case "EUR":
		return "€"
	case "GBP":
		return "£"
	default:
		return code
	}
}
