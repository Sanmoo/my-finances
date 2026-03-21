package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNamespace(t *testing.T) {
	tests := []struct {
		name    string
		nsName  string
		opts    []NamespaceOption
		wantErr error
	}{
		{
			name:    "valid namespace",
			nsName:  "main",
			opts:    []NamespaceOption{},
			wantErr: nil,
		},
		{
			name:    "with leading/trailing spaces",
			nsName:  "  main  ",
			opts:    []NamespaceOption{},
			wantErr: nil,
		},
		{
			name:    "empty name",
			nsName:  "",
			opts:    []NamespaceOption{},
			wantErr: ErrEmptyNamespaceName,
		},
		{
			name:    "only spaces",
			nsName:  "   ",
			opts:    []NamespaceOption{},
			wantErr: ErrEmptyNamespaceName,
		},
		{
			name:    "with default credit card",
			nsName:  "main",
			opts:    []NamespaceOption{WithDefaultCreditCard(5)},
			wantErr: nil,
		},
		{
			name:    "with custom currency",
			nsName:  "main",
			opts:    []NamespaceOption{WithDefaultCurrency("USD")},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns, err := NewNamespace(tt.nsName, tt.opts...)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, ns)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ns)
				assert.Equal(t, "main", ns.Name)
			}
		})
	}
}

func TestNewNamespace_DefaultCurrencyValidation(t *testing.T) {
	_, err := NewNamespace("test", WithDefaultCurrency("US"))
	assert.ErrorIs(t, err, ErrInvalidCurrency)
}
