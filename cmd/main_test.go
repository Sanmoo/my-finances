package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeNegativeAmountArgs(t *testing.T) {
	args := []string{"add", "expense", "-d", "18", "-a", "sam", "-c", "rest", "-D", "Restituição livup para o Edu", "-20"}

	got := normalizeNegativeAmountArgs(args)

	assert.Equal(t, []string{"add", "expense", "-d", "18", "-a", "sam", "-c", "rest", "-D", "Restituição livup para o Edu", "--", "-20"}, got)
}
