package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	// ErrInvalidDateFormat indicates the input matched none of the supported formats.
	ErrInvalidDateFormat = errors.New("invalid date format")
	// ErrImpossibleRollback indicates a bare day later than today whose previous
	// month has no such day (e.g. "31" right after a 30-day month).
	ErrImpossibleRollback = errors.New("day does not exist in the previous month")
)

// ParseDate parses a flexible date format: YYYY-MM-DD, YY-MM-DD, MM-DD or DD.
// Missing parts fall back to now's year and month, even when the result is in the
// future. It returns the zero time when the input matches no format.
func ParseDate(dateStr string, now time.Time) time.Time {
	if strings.TrimSpace(dateStr) == "" {
		return time.Time{}
	}

	dateStr = strings.TrimSpace(dateStr)

	// Format: YYYY-MM-DD
	if t, err := time.Parse("2006-01-02", dateStr); err == nil {
		return t.UTC()
	}

	// Format: YY-MM-DD
	if t, err := time.Parse("06-01-02", dateStr); err == nil {
		if t.Year() < 100 {
			t = t.AddDate(2000, 0, 0)
		}
		return t.UTC()
	}

	// Format: MM-DD (use current year)
	if t, err := time.Parse("01-02", dateStr); err == nil {
		return time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	}

	// Format: DD (use current month and year)
	if t, err := time.Parse("2", dateStr); err == nil {
		return time.Date(now.Year(), now.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	}

	return time.Time{}
}

// ParseRealizationDate parses the realization date a user typed. An implicit date
// never lands in the future: a bare day (DD) later than today belongs to the
// previous month, so on 04/10 the input "27" means 27/09. Formats that name a
// month or a year are taken literally.
func ParseRealizationDate(dateStr string, now time.Time) (time.Time, error) {
	trimmed := strings.TrimSpace(dateStr)

	date := ParseDate(trimmed, now)
	if date.IsZero() {
		return time.Time{}, fmt.Errorf("%w: %q (expected DD, MM-DD, YY-MM-DD or YYYY-MM-DD)", ErrInvalidDateFormat, trimmed)
	}

	if !isBareDay(trimmed) || date.Day() <= now.Day() {
		return date, nil
	}

	// Step through the first of the month: AddDate(0, -1, 0) applied to a day 31
	// would overflow into the following month.
	firstOfCurrentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	previousMonth := firstOfCurrentMonth.AddDate(0, -1, 0)

	rolled := time.Date(previousMonth.Year(), previousMonth.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	if rolled.Month() != previousMonth.Month() {
		return time.Time{}, fmt.Errorf("%w: day %d does not exist in %s %d", ErrImpossibleRollback, date.Day(), strings.ToLower(previousMonth.Month().String()), previousMonth.Year())
	}

	return rolled, nil
}

// isBareDay reports whether the input is a day number without a month or year.
func isBareDay(dateStr string) bool {
	if strings.Contains(dateStr, "-") {
		return false
	}
	_, err := time.Parse("2", dateStr)
	return err == nil
}
