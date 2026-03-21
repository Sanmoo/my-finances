package expr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		want    float64
		wantErr error
	}{
		{
			name: "simple number",
			expr: "42",
			want: 42.0,
		},
		{
			name: "decimal number",
			expr: "3.14",
			want: 3.14,
		},
		{
			name: "simple addition",
			expr: "2+3",
			want: 5.0,
		},
		{
			name: "simple subtraction",
			expr: "10-3",
			want: 7.0,
		},
		{
			name: "simple multiplication",
			expr: "4*5",
			want: 20.0,
		},
		{
			name: "simple division",
			expr: "20/4",
			want: 5.0,
		},
		{
			name: "mixed operations with precedence",
			expr: "2+3*4",
			want: 14.0,
		},
		{
			name: "mixed operations with parentheses",
			expr: "(2+3)*4",
			want: 20.0,
		},
		{
			name: "complex expression",
			expr: "10+2*5-20/4",
			want: 15.0,
		},
		{
			name:    "division by zero",
			expr:    "10/0",
			want:    0,
			wantErr: ErrDivisionByZero,
		},
		{
			name: "with spaces",
			expr: "2 + 3 * 4",
			want: 14.0,
		},
		{
			name: "negative numbers",
			expr: "-5+3",
			want: -2.0,
		},
		{
			name: "double negative",
			expr: "--5",
			want: 5.0,
		},
		{
			name: "nested parentheses",
			expr: "((2+3)*4)",
			want: 20.0,
		},
		{
			name: "expression in parentheses",
			expr: "(2+3)*(4+1)",
			want: 25.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.expr)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestParse_Errors(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr error
	}{
		{
			name:    "empty expression",
			expr:    "",
			wantErr: ErrUnexpectedEnd,
		},
		{
			name:    "invalid character",
			expr:    "2+a",
			wantErr: ErrUnexpectedChar,
		},
		{
			name:    "unclosed parenthesis",
			expr:    "(2+3",
			wantErr: ErrInvalidExpression,
		},
		{
			name:    "unexpected end in expression",
			expr:    "2+",
			wantErr: ErrUnexpectedEnd,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.expr)
			assert.Error(t, err)
		})
	}
}

func TestNewParser(t *testing.T) {
	p := NewParser(" 2 + 3 * 4 ")
	assert.Equal(t, "2+3*4", p.expr)
	assert.Equal(t, 0, p.pos)
}
