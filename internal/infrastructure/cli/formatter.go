package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"github.com/Sanmoo/my-finances/internal/infrastructure/i18n"
)

type Formatter struct {
	locale *i18n.Locale
}

type AccountBalance struct {
	Account      *entity.Account
	TotalIncome  float64
	TotalExpense float64
	Balance      float64
}

func NewFormatter(locale *i18n.Locale) *Formatter {
	return &Formatter{locale: locale}
}

func (f *Formatter) FormatAccount(acc *entity.Account) string {
	return fmt.Sprintf("Account created: %s (ID: %d)", acc.Name, acc.ID)
}

func (f *Formatter) FormatCategory(cat *entity.Category) string {
	emoji := ""
	if cat.Emoji != nil {
		emoji = *cat.Emoji + " "
	}
	return fmt.Sprintf("Category created: %s%s (ID: %d, Type: %s)", emoji, cat.Name, cat.ID, cat.Type)
}

func (f *Formatter) FormatCreditCard(cc *entity.CreditCard) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Credit card created: %s (ID: %d)\n", cc.Name, cc.ID))
	sb.WriteString(fmt.Sprintf("  Closing day: %d, Due day: %d", cc.ClosingDay, cc.DueDay))
	return sb.String()
}

func (f *Formatter) FormatEntry(entry *entity.Entry, tags map[int64]*entity.Tag) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Entry created (ID: %d)", entry.ID))
	lines = append(lines, fmt.Sprintf("  Type: %s, Amount: %s", entry.Type, f.locale.FormatCurrency(entry.Amount, entry.Currency)))
	lines = append(lines, fmt.Sprintf("  Date: %s", f.locale.FormatDate(entry.RealizationDate)))
	if entry.PaymentDate != nil {
		lines = append(lines, fmt.Sprintf("  Payment Date: %s", f.locale.FormatDate(*entry.PaymentDate)))
	}
	if entry.Description != "" {
		lines = append(lines, fmt.Sprintf("  Description: %s", entry.Description))
	}
	if len(entry.TagIDs) > 0 {
		lines = append(lines, fmt.Sprintf("  Tags: %s", f.formatTagNames(entry.TagIDs, tags)))
	}
	return strings.Join(lines, "\n")
}

func (f *Formatter) FormatEntryWithCategory(entry *entity.Entry, category *entity.Category, tags map[int64]*entity.Tag) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Entry created (ID: %d)", entry.ID))
	lines = append(lines, fmt.Sprintf("  Type: %s, Amount: %s", entry.Type, f.locale.FormatCurrency(entry.Amount, entry.Currency)))
	lines = append(lines, fmt.Sprintf("  Date: %s", f.locale.FormatDate(entry.RealizationDate)))
	if entry.PaymentDate != nil {
		lines = append(lines, fmt.Sprintf("  Payment Date: %s", f.locale.FormatDate(*entry.PaymentDate)))
	}
	if category != nil {
		emoji := ""
		if category.Emoji != nil {
			emoji = *category.Emoji + " "
		}
		lines = append(lines, fmt.Sprintf("  Category: %s%s", emoji, category.Name))
	}
	if entry.Description != "" {
		lines = append(lines, fmt.Sprintf("  Description: %s", entry.Description))
	}
	if len(entry.TagIDs) > 0 {
		lines = append(lines, fmt.Sprintf("  Tags: %s", f.formatTagNames(entry.TagIDs, tags)))
	}
	return strings.Join(lines, "\n")
}

func (f *Formatter) FormatEntryMarkdown(entry *entity.Entry, category *entity.Category, tags map[int64]*entity.Tag) string {
	emoji := ""
	if category != nil && category.Emoji != nil {
		emoji = *category.Emoji + " "
	}

	dateStr := f.locale.FormatDate(entry.RealizationDate)
	paymentStr := ""
	if entry.PaymentDate != nil {
		paymentStr = fmt.Sprintf(" → %s", f.locale.FormatDate(*entry.PaymentDate))
	}

	descStr := ""
	if entry.Description != "" {
		descStr = fmt.Sprintf(" — _%s_", entry.Description)
	}

	tagStr := ""
	if len(entry.TagIDs) > 0 {
		tagStr = fmt.Sprintf(" `[%s]`", f.formatTagNames(entry.TagIDs, tags))
	}

	catName := ""
	if category != nil {
		catName = emoji + category.Name
	}

	entryType := "💰"
	if entry.IsExpense() {
		entryType = "💸"
	}

	return fmt.Sprintf("| %s | %s | %s | %s |%s%s%s |",
		dateStr+paymentStr,
		catName,
		entryType,
		f.locale.FormatCurrency(entry.Amount, entry.Currency),
		descStr,
		tagStr,
		"")
}

func (f *Formatter) FormatEntryDescription(entry *entity.Entry, totalInstallments int) string {
	desc := entry.Description
	if entry.CreditCardID != nil {
		ccPart := "[CC] (" + f.locale.FormatDate(entry.RealizationDate) + ")"
		if totalInstallments > 1 {
			ccPart += fmt.Sprintf(" (%d/%d)", entry.Installment, totalInstallments)
		}
		if desc != "" {
			ccPart += " - " + desc
		}
		desc = ccPart
	}
	return desc
}

func (f *Formatter) FormatEntriesTable(entries []*entity.Entry, categories map[int64]*entity.Category, tags map[int64]*entity.Tag, totalInstallments map[int64]int) string {
	var sb strings.Builder

	expenses, incomes := f.separateByType(entries)

	if len(expenses) > 0 {
		sb.WriteString("=== Expenses ===\n")
		sb.WriteString(fmt.Sprintf("%-12s | %-15s | %-12s | %s\n", "Date", "Category", "Amount", "Description"))
		sb.WriteString(strings.Repeat("-", 80) + "\n")

		for _, entry := range expenses {
			sb.WriteString(f.formatEntryTableRow(entry, categories, tags, totalInstallments))
		}
		sb.WriteString("\n")
	}

	if len(incomes) > 0 {
		sb.WriteString("=== Incomes ===\n")
		sb.WriteString(fmt.Sprintf("%-12s | %-15s | %-12s | %s\n", "Date", "Category", "Amount", "Description"))
		sb.WriteString(strings.Repeat("-", 80) + "\n")

		for _, entry := range incomes {
			sb.WriteString(f.formatEntryTableRow(entry, categories, tags, nil))
		}
	}

	return sb.String()
}

func (f *Formatter) FormatEntriesMarkdown(entries []*entity.Entry, categories map[int64]*entity.Category, tags map[int64]*entity.Tag, totalInstallments map[int64]int) string {
	var sb strings.Builder

	expenses, incomes := f.separateByType(entries)

	if len(expenses) > 0 {
		sb.WriteString("## Expenses\n\n")
		sb.WriteString("| Date | Category | Amount | Description |\n")
		sb.WriteString("|------|----------|--------|-------------|\n")

		for _, entry := range expenses {
			sb.WriteString(f.formatEntryMarkdownRow(entry, categories, tags, totalInstallments))
		}
		sb.WriteString("\n")
	}

	if len(incomes) > 0 {
		sb.WriteString("## Incomes\n\n")
		sb.WriteString("| Date | Category | Amount | Description |\n")
		sb.WriteString("|------|----------|--------|-------------|\n")

		for _, entry := range incomes {
			sb.WriteString(f.formatEntryMarkdownRow(entry, categories, tags, nil))
		}
	}

	return sb.String()
}

func (f *Formatter) FormatReportMarkdown(entries []*entity.Entry, categories map[int64]*entity.Category, tags map[int64]*entity.Tag, from, to *time.Time) string {
	var sb strings.Builder

	sb.WriteString("# Financial Report\n\n")

	if from != nil || to != nil {
		period := ""
		if from != nil {
			period += fmt.Sprintf("from %s ", f.locale.FormatDate(*from))
		}
		if to != nil {
			period += fmt.Sprintf("until %s", f.locale.FormatDate(*to))
		}
		sb.WriteString(fmt.Sprintf("**Period:** %s\n\n", strings.TrimSpace(period)))
	}

	sb.WriteString("| Date | Category | Type | Amount |\n")
	sb.WriteString("|------|----------|------|--------|\n")

	for _, entry := range entries {
		cat := categories[*entry.CategoryID]
		sb.WriteString(f.FormatEntryMarkdown(entry, cat, tags) + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString(f.FormatSummary(entries))

	return sb.String()
}

func (f *Formatter) FormatSummary(entries []*entity.Entry) string {
	var totalIncome, totalExpense float64

	for _, entry := range entries {
		if entry.IsIncome() {
			totalIncome += entry.Amount
		} else {
			totalExpense += entry.Amount
		}
	}

	var sb strings.Builder
	sb.WriteString("## Summary\n")
	sb.WriteString(fmt.Sprintf("- **Total Income:** %s\n", f.locale.FormatNumber(totalIncome)))
	sb.WriteString(fmt.Sprintf("- **Total Expense:** %s\n", f.locale.FormatNumber(totalExpense)))
	sb.WriteString(fmt.Sprintf("- **Balance:** %s\n", f.locale.FormatNumber(totalIncome-totalExpense)))

	return sb.String()
}

func (f *Formatter) FormatBalancesTable(accounts []*entity.Account, accountBalances map[int64]*AccountBalance, from, to *time.Time) string {
	var sb strings.Builder

	for _, acc := range accounts {
		ab := accountBalances[acc.ID]
		if ab == nil {
			ab = &AccountBalance{Account: acc}
		}

		sb.WriteString(fmt.Sprintf("=== %s ===\n\n", acc.Name))

		if from != nil || to != nil {
			period := ""
			if from != nil {
				period += f.locale.FormatDate(*from)
			}
			if from != nil && to != nil {
				period += " - "
			}
			if to != nil {
				period += f.locale.FormatDate(*to)
			}
			sb.WriteString(fmt.Sprintf("Period: %s\n\n", period))
		}

		sb.WriteString(fmt.Sprintf("  Total Income:  %s\n", f.locale.FormatNumber(ab.TotalIncome)))
		sb.WriteString(fmt.Sprintf("  Total Expense: %s\n", f.locale.FormatNumber(ab.TotalExpense)))
		sb.WriteString("  " + strings.Repeat("-", 30) + "\n")
		sb.WriteString(fmt.Sprintf("  Balance:       %s\n\n", f.locale.FormatNumber(ab.Balance)))
	}

	return sb.String()
}

func (f *Formatter) FormatBalancesMarkdown(accounts []*entity.Account, accountBalances map[int64]*AccountBalance, from, to *time.Time) string {
	var sb strings.Builder

	sb.WriteString("# Balances\n\n")

	for _, acc := range accounts {
		ab := accountBalances[acc.ID]
		if ab == nil {
			ab = &AccountBalance{Account: acc}
		}

		sb.WriteString(fmt.Sprintf("## %s\n\n", acc.Name))

		if from != nil || to != nil {
			period := ""
			if from != nil {
				period += f.locale.FormatDate(*from)
			}
			if from != nil && to != nil {
				period += " - "
			}
			if to != nil {
				period += f.locale.FormatDate(*to)
			}
			sb.WriteString(fmt.Sprintf("**Period:** %s\n\n", period))
		}

		sb.WriteString("| | Amount |\n")
		sb.WriteString("|---|---:|\n")
		sb.WriteString(fmt.Sprintf("| Total Income | %s |\n", f.locale.FormatNumber(ab.TotalIncome)))
		sb.WriteString(fmt.Sprintf("| Total Expense | %s |\n", f.locale.FormatNumber(ab.TotalExpense)))
		sb.WriteString(fmt.Sprintf("| **Balance** | **%s** |\n\n", f.locale.FormatNumber(ab.Balance)))
	}

	return sb.String()
}

func (f *Formatter) FormatError(msg string) string {
	return fmt.Sprintf("Error: %s", msg)
}

func (f *Formatter) FormatSuccess(msg string) string {
	return fmt.Sprintf("Success: %s", msg)
}

func (f *Formatter) formatTagNames(ids []int64, tags map[int64]*entity.Tag) string {
	if len(ids) == 0 || tags == nil {
		return ""
	}
	strs := make([]string, 0, len(ids))
	for _, id := range ids {
		if tag, ok := tags[id]; ok {
			strs = append(strs, tag.Name)
		} else {
			strs = append(strs, fmt.Sprintf("tag_%d", id))
		}
	}
	return strings.Join(strs, ", ")
}

func (f *Formatter) separateByType(entries []*entity.Entry) (expenses, incomes []*entity.Entry) {
	for _, e := range entries {
		if e.IsExpense() {
			expenses = append(expenses, e)
		} else {
			incomes = append(incomes, e)
		}
	}
	return
}

func (f *Formatter) formatEntryTableRow(entry *entity.Entry, categories map[int64]*entity.Category, tags map[int64]*entity.Tag, totalInstallments map[int64]int) string {
	var sb strings.Builder

	dateStr := f.locale.FormatDate(entry.RealizationDate)
	catName := ""
	if entry.CategoryID != nil {
		if cat, ok := categories[*entry.CategoryID]; ok {
			catName = cat.Name
		}
	}

	amountStr := f.locale.FormatCurrency(entry.Amount, entry.Currency)

	ti := 0
	if totalInstallments != nil {
		if entry.ParentEntryID != nil {
			ti = totalInstallments[*entry.ParentEntryID]
		} else {
			ti = totalInstallments[entry.ID]
		}
	}

	desc := f.FormatEntryDescription(entry, ti)

	tagsStr := ""
	if len(entry.TagIDs) > 0 && tags != nil {
		tagsStr = f.formatTagNames(entry.TagIDs, tags)
	}

	sb.WriteString(fmt.Sprintf("%-12s | %-15s | %-12s | %s\n", dateStr, catName, amountStr, desc))
	if tagsStr != "" {
		sb.WriteString(fmt.Sprintf("%-12s   %-15s   %-12s   Tags: %s\n", "", "", "", tagsStr))
	}

	return sb.String()
}

func (f *Formatter) formatEntryMarkdownRow(entry *entity.Entry, categories map[int64]*entity.Category, tags map[int64]*entity.Tag, totalInstallments map[int64]int) string {
	dateStr := f.locale.FormatDate(entry.RealizationDate)
	catName := ""
	if entry.CategoryID != nil {
		if cat, ok := categories[*entry.CategoryID]; ok {
			if cat.Emoji != nil {
				catName = *cat.Emoji + " " + cat.Name
			} else {
				catName = cat.Name
			}
		}
	}

	amountStr := f.locale.FormatCurrency(entry.Amount, entry.Currency)

	ti := 0
	if totalInstallments != nil {
		if entry.ParentEntryID != nil {
			ti = totalInstallments[*entry.ParentEntryID]
		} else {
			ti = totalInstallments[entry.ID]
		}
	}

	desc := f.FormatEntryDescription(entry, ti)

	tagsStr := ""
	if len(entry.TagIDs) > 0 && tags != nil {
		tagsStr = fmt.Sprintf("`[%s]`", f.formatTagNames(entry.TagIDs, tags))
	}

	return fmt.Sprintf("| %s | %s | %s | %s %s |\n", dateStr, catName, amountStr, desc, tagsStr)
}
