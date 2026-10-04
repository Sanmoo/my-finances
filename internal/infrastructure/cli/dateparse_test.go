package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

var (
	october4th2026 = time.Date(2026, time.October, 4, 10, 21, 0, 0, time.UTC)
	january4th2027 = time.Date(2027, time.January, 4, 9, 0, 0, 0, time.UTC)
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
	}{
		{"YYYY-MM-DD", "2026-03-15", date(2026, time.March, 15)},
		{"YY-MM-DD", "26-03-15", date(2026, time.March, 15)},
		{"YY-MM-DD maps 99 to 1999", "99-03-15", date(1999, time.March, 15)},
		{"MM-DD uses the current year", "03-15", date(2026, time.March, 15)},
		{"DD uses the current month and year", "15", date(2026, time.October, 15)},
		{"DD may land in the future", "27", date(2026, time.October, 27)},
		{"empty", "", time.Time{}},
		{"spaces only", "   ", time.Time{}},
		{"invalid", "abc", time.Time{}},
		{"out of range day", "32", time.Time{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ParseDate(tt.input, october4th2026))
		})
	}
}

func TestParseRealizationDate(t *testing.T) {
	tests := []struct {
		name     string
		now      time.Time
		input    string
		expected time.Time
		wantErr  error
	}{
		{"bare day already past", october4th2026, "02", date(2026, time.October, 2), nil},
		{"bare day is today", october4th2026, "04", date(2026, time.October, 4), nil},
		{"bare day future falls back to the previous month", october4th2026, "27", date(2026, time.September, 27), nil},
		{"bare day one day ahead falls back too", october4th2026, "5", date(2026, time.September, 5), nil},
		{"bare day with leading zero", october4th2026, "05", date(2026, time.September, 5), nil},
		{"bare day padded with spaces", october4th2026, " 27 ", date(2026, time.September, 27), nil},
		{"bare day falls back across the year", january4th2027, "27", date(2026, time.December, 27), nil},
		{"bare day missing from the previous month", time.Date(2026, time.March, 30, 12, 0, 0, 0, time.UTC), "31", time.Time{}, ErrImpossibleRollback},
		{"explicit month may be in the future", october4th2026, "12-27", date(2026, time.December, 27), nil},
		{"explicit year may be in the future", october4th2026, "2026-10-27", date(2026, time.October, 27), nil},
		{"short explicit year", october4th2026, "26-09-27", date(2026, time.September, 27), nil},
		{"invalid format", october4th2026, "abc", time.Time{}, ErrInvalidDateFormat},
		{"out of range day", october4th2026, "32", time.Time{}, ErrInvalidDateFormat},
		{"empty", october4th2026, "", time.Time{}, ErrInvalidDateFormat},
		{"spaces only", october4th2026, "   ", time.Time{}, ErrInvalidDateFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseRealizationDate(tt.input, tt.now)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.True(t, result.IsZero())
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
