package interactive

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStdioPrompter_Text(t *testing.T) {
	t.Run("returns input without default", func(t *testing.T) {
		input := "hello\n"
		p := NewStdioPrompterWithIO(strings.NewReader(input), &strings.Builder{})

		result, err := p.Text("Name", "")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("returns default on empty input", func(t *testing.T) {
		input := "\n"
		p := NewStdioPrompterWithIO(strings.NewReader(input), &strings.Builder{})

		result, err := p.Text("Name", "defaultVal")
		require.NoError(t, err)
		assert.Equal(t, "defaultVal", result)
	})

	t.Run("trims whitespace", func(t *testing.T) {
		input := "  foo bar  \n"
		p := NewStdioPrompterWithIO(strings.NewReader(input), &strings.Builder{})

		result, err := p.Text("Name", "")
		require.NoError(t, err)
		assert.Equal(t, "foo bar", result)
	})
}

func TestStdioPrompter_Confirm(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		defaultYes bool
		expected   bool
	}{
		{"empty with default yes", "\n", true, true},
		{"empty with default no", "\n", false, false},
		{"y", "y\n", false, true},
		{"yes", "yes\n", false, true},
		{"s", "s\n", false, true},
		{"sim", "sim\n", false, true},
		{"n", "n\n", true, false},
		{"no", "no\n", true, false},
		{"random", "blah\n", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewStdioPrompterWithIO(strings.NewReader(tt.input), &strings.Builder{})
			result, err := p.Confirm("Continue?", tt.defaultYes)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
