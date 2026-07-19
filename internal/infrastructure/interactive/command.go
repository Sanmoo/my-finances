package interactive

import (
	"fmt"
	"strings"

	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

// RenderCLI returns the equivalent myfin CLI command for the given parameters.
// The output is copy-pasteable for non-interactive use.
func RenderCLI(typ entity.EntryType, amount, account, date, description, category, creditCard string, tags []string, times int) string {
	var parts []string

	parts = append(parts, "myfin", "add", string(typ), amount)
	parts = append(parts, "--account", account)
	parts = append(parts, "--date", date)
	parts = append(parts, "--description", formatDescription(description))

	if category != "" {
		parts = append(parts, "--category", category)
	}
	if creditCard != "" {
		parts = append(parts, "--credit-card", creditCard)
	}
	if creditCard != "" && times > 0 {
		parts = append(parts, "--times", fmt.Sprintf("%d", times))
	}
	if len(tags) > 0 {
		parts = append(parts, "--tags", strings.Join(tags, ","))
	}

	return strings.Join(parts, " ")
}

func formatDescription(desc string) string {
	if strings.Contains(desc, " ") {
		return fmt.Sprintf("%q", desc)
	}
	return desc
}
