package interactive

import (
	"testing"

	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestRenderCLI(t *testing.T) {
	t.Run("minimal expense", func(t *testing.T) {
		cmd := RenderCLI(entity.EntryTypeExpense, "50.00", "nubank", "2025-06-15", "almoço", "", "", nil, 0)
		expected := `myfin add expense 50.00 --account nubank --date 2025-06-15 --description almoço`
		assert.Equal(t, expected, cmd)
	})

	t.Run("expense with all fields", func(t *testing.T) {
		cmd := RenderCLI(entity.EntryTypeExpense, "1000/3", "nubank", "2025-06-15", "almoço", "food", "nu", []string{"work", "vr"}, 3)
		expected := `myfin add expense 1000/3 --account nubank --date 2025-06-15 --description almoço --category food --credit-card nu --times 3 --tags work,vr`
		assert.Equal(t, expected, cmd)
	})

	t.Run("minimal income", func(t *testing.T) {
		cmd := RenderCLI(entity.EntryTypeIncome, "5000", "nubank", "2025-06-15", "salário", "", "", nil, 0)
		expected := `myfin add income 5000 --account nubank --date 2025-06-15 --description salário`
		assert.Equal(t, expected, cmd)
	})

	t.Run("description without spaces does not get quoted", func(t *testing.T) {
		cmd := RenderCLI(entity.EntryTypeExpense, "10", "nubank", "2025-06-15", "café", "", "", nil, 0)
		expected := `myfin add expense 10 --account nubank --date 2025-06-15 --description café`
		assert.Equal(t, expected, cmd)
	})

	t.Run("description with spaces gets quoted", func(t *testing.T) {
		cmd := RenderCLI(entity.EntryTypeExpense, "10", "nubank", "2025-06-15", "comida japonesa", "", "", nil, 0)
		expected := `myfin add expense 10 --account nubank --date 2025-06-15 --description "comida japonesa"`
		assert.Equal(t, expected, cmd)
	})

	t.Run("no category, credit-card, or tags are omitted", func(t *testing.T) {
		cmd := RenderCLI(entity.EntryTypeExpense, "10", "nubank", "2025-06-15", "desc", "", "", nil, 0)
		assert.NotContains(t, cmd, "--category")
		assert.NotContains(t, cmd, "--credit-card")
		assert.NotContains(t, cmd, "--times")
		assert.NotContains(t, cmd, "--tags")
	})
}
