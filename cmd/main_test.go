package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeNegativeAmountArgs(t *testing.T) {
	args := []string{"add", "expense", "-d", "18", "-a", "sam", "-c", "rest", "-D", "Restituição livup para o Edu", "-20"}

	got := normalizeNegativeAmountArgs(args)

	assert.Equal(t, []string{"add", "expense", "-d", "18", "-a", "sam", "-c", "rest", "-D", "Restituição livup para o Edu", "--", "-20"}, got)
}

func TestParseMonth(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantYear  int
		wantMonth time.Month
		wantErr   bool
	}{
		{
			name:      "full year and month",
			input:     "2026-06",
			wantYear:  2026,
			wantMonth: time.June,
		},
		{
			name:      "two-digit year and month",
			input:     "26-06",
			wantYear:  2026,
			wantMonth: time.June,
		},
		{
			name:      "month only uses current year",
			input:     "06",
			wantYear:  time.Now().Year(),
			wantMonth: time.June,
		},
		{
			name:    "invalid month",
			input:   "abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMonth(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantYear, got.Year())
			assert.Equal(t, tt.wantMonth, got.Month())
		})
	}
}
