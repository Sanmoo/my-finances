package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAccount(t *testing.T) {
	tests := []struct {
		name    string
		accName string
		wantErr error
	}{
		{
			name:    "valid account",
			accName: "Checking",
			wantErr: nil,
		},
		{
			name:    "with leading/trailing spaces",
			accName: "  Checking  ",
			wantErr: nil,
		},
		{
			name:    "empty name",
			accName: "",
			wantErr: ErrEmptyAccountName,
		},
		{
			name:    "only spaces",
			accName: "   ",
			wantErr: ErrEmptyAccountName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc, err := NewAccount(tt.accName)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, acc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, acc)
				assert.Equal(t, "checking", acc.Name)
			}
		})
	}
}
