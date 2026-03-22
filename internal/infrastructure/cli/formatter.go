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
	return fmt.Sprintf("Account created: %s", acc.Name)
}

func (f *Formatter) FormatCategory(cat *entity.Category) string {
	emoji := ""
	if cat.Emoji != nil {
		emoji = *cat.Emoji + " "
	}
	return fmt.Sprintf("Category created: %s%s (Alias: %s, Type: %s)", emoji, cat.Name, cat.Alias, cat.Type)
}

func (f *Formatter) FormatCreditCard(cc *entity.CreditCard) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Credit card created: %s\n", cc.Name))
	sb.WriteString(fmt.Sprintf("  Closing day: %d, Due day: %d", cc.ClosingDay, cc.DueDay))
	return sb.String()
}

func (f *Formatter) FormatEntry(entry *entity.Entry) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Entry created"))
	lines = append(lines, fmt.Sprintf("  Type: %s, Amount: %s", entry.Type, f.locale.FormatCurrency(entry.Amount, entry.Currency)))
	lines = append(lines, fmt.Sprintf("  Date: %s", f.locale.FormatDate(entry.RealizationDate)))
	if entry.PaymentDate != nil {
		lines = append(lines, fmt.Sprintf("  Payment Date: %s", f.locale.FormatDate(*entry.PaymentDate)))
	}
	if entry.Description != "" {
		lines = append(lines, fmt.Sprintf("  Description: %s", entry.Description))
	}
	if len(entry.Tags) > 0 {
		lines = append(lines, fmt.Sprintf("  Tags: %s", strings.Join(entry.Tags, ", ")))
	}
	return strings.Join(lines, "\n")
}

func (f *Formatter) FormatEntryWithCategory(entry *entity.Entry, category *entity.Category) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Entry created"))
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
	if len(entry.Tags) > 0 {
		lines = append(lines, fmt.Sprintf("  Tags: %s", strings.Join(entry.Tags, ", ")))
	}
	return strings.Join(lines, "\n")
}

func (f *Formatter) FormatEntryMarkdown(entry *entity.Entry, category *entity.Category) string {
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
	if len(entry.Tags) > 0 {
		tagStr = fmt.Sprintf(" `[%s]`", strings.Join(entry.Tags, ", "))
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

func (f *Formatter) FormatEntryDescription(entry *entity.Entry) string {
	desc := entry.Description
	if entry.CreditCardName != nil {
		ccPart := "[CC] (" + f.locale.FormatDate(entry.RealizationDate) + ")"
		if entry.InstallmentTotal > 1 {
			ccPart += fmt.Sprintf(" (%d/%d)", entry.InstallmentNumber, entry.InstallmentTotal)
		}
		if desc != "" {
			ccPart += " - " + desc
		}
		desc = ccPart
	}
	return desc
}

func (f *Formatter) FormatEntriesTable(entries []*entity.Entry, categories map[string]*entity.Category, accounts map[string]*entity.Account, filteredAccount string) string {
	expenses, incomes := f.separateByTypeForReport(entries)
	var sb strings.Builder

	if len(expenses) > 0 {
		catWidth := f.calculateCategoryWidthForReport(expenses, categories)
		sb.WriteString("=== Expenses ===\n")
		headerFormat := fmt.Sprintf("%%-12s | %%%ds | %%-12s | %%s\n", -catWidth)
		sb.WriteString(fmt.Sprintf(headerFormat, "Date", "Category", "Amount", "Description"))
		separatorLen := 12 + 3 + catWidth + 3 + 12 + 3 + 11
		sb.WriteString(strings.Repeat("-", separatorLen) + "\n")

		for _, entry := range expenses {
			sb.WriteString(f.formatReportEntryRow(entry, categories, catWidth))
		}
		sb.WriteString("\n")
	}

	if len(incomes) > 0 {
		catWidth := f.calculateCategoryWidthForReport(incomes, categories)
		sb.WriteString("=== Incomes ===\n")
		headerFormat := fmt.Sprintf("%%-12s | %%%ds | %%-12s | %%s\n", -catWidth)
		sb.WriteString(fmt.Sprintf(headerFormat, "Date", "Category", "Amount", "Description"))
		separatorLen := 12 + 3 + catWidth + 3 + 12 + 3 + 11
		sb.WriteString(strings.Repeat("-", separatorLen) + "\n")

		for _, entry := range incomes {
			sb.WriteString(f.formatReportEntryRow(entry, categories, catWidth))
		}
	}

	return sb.String()
}

func (f *Formatter) calculateCategoryWidthForReport(entries []*entity.Entry, categories map[string]*entity.Category) int {
	minWidth := len("Category")
	for _, entry := range entries {
		if entry.CategoryAlias != nil {
			if cat, ok := categories[*entry.CategoryAlias]; ok {
				displayName := f.getCategoryDisplayName(cat)
				if len(displayName) > minWidth {
					minWidth = len(displayName)
				}
			}
		}
	}
	return minWidth
}

func (f *Formatter) separateByTypeForReport(entries []*entity.Entry) (expenses, incomes []*entity.Entry) {
	for _, e := range entries {
		if e.IsExpense() {
			expenses = append(expenses, e)
		} else {
			incomes = append(incomes, e)
		}
	}
	return
}

func (f *Formatter) formatReportEntryRow(entry *entity.Entry, categories map[string]*entity.Category, catWidth int) string {
	dateStr := f.locale.FormatDate(entry.RealizationDate)

	catName := ""
	if entry.CategoryAlias != nil {
		if cat, ok := categories[*entry.CategoryAlias]; ok {
			catName = f.getCategoryDisplayName(cat)
		}
	}

	amountStr := f.locale.FormatCurrency(entry.Amount, entry.Currency)

	desc := f.FormatEntryDescription(entry)

	tagsStr := ""
	if len(entry.Tags) > 0 {
		tagsStr = fmt.Sprintf(" [%s]", strings.Join(entry.Tags, ", "))
	}

	formatStr := fmt.Sprintf("%%-12s | %%%ds | %%-12s | %%s%%s\n", -catWidth)
	return fmt.Sprintf(formatStr, dateStr, catName, amountStr, desc, tagsStr)
}

func (f *Formatter) FormatEntriesMarkdown(entries []*entity.Entry, categories map[string]*entity.Category, accounts map[string]*entity.Account, filteredAccount string) string {
	expenses, incomes := f.separateByTypeForReport(entries)
	var sb strings.Builder

	if len(expenses) > 0 {
		sb.WriteString("## Expenses\n\n")
		sb.WriteString("| Date | Category | Amount | Description |\n")
		sb.WriteString("|------|----------|--------|-------------|\n")

		for _, entry := range expenses {
			sb.WriteString(f.formatReportEntryMarkdownRow(entry, categories))
		}
		sb.WriteString("\n")
	}

	if len(incomes) > 0 {
		sb.WriteString("## Incomes\n\n")
		sb.WriteString("| Date | Category | Amount | Description |\n")
		sb.WriteString("|------|----------|--------|-------------|\n")

		for _, entry := range incomes {
			sb.WriteString(f.formatReportEntryMarkdownRow(entry, categories))
		}
	}

	return sb.String()
}

func (f *Formatter) formatReportEntryMarkdownRow(entry *entity.Entry, categories map[string]*entity.Category) string {
	dateStr := f.locale.FormatDate(entry.RealizationDate)

	catName := ""
	if entry.CategoryAlias != nil {
		if cat, ok := categories[*entry.CategoryAlias]; ok {
			catName = f.getCategoryDisplayName(cat)
		}
	}

	amountStr := f.locale.FormatCurrency(entry.Amount, entry.Currency)
	desc := f.FormatEntryDescription(entry)

	tagsStr := ""
	if len(entry.Tags) > 0 {
		tagsStr = fmt.Sprintf(" `[%s]`", strings.Join(entry.Tags, ", "))
	}

	return fmt.Sprintf("| %s | %s | %s | %s%s |\n", dateStr, catName, amountStr, desc, tagsStr)
}

func (f *Formatter) FormatReportMarkdown(entries []*entity.Entry, categories map[string]*entity.Category, from, to *time.Time) string {
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
		var cat *entity.Category
		if entry.CategoryAlias != nil {
			cat = categories[*entry.CategoryAlias]
		}
		sb.WriteString(f.FormatEntryMarkdown(entry, cat) + "\n")
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

func (f *Formatter) FormatBalancesTable(accounts []*entity.Account, accountBalances map[string]*AccountBalance, from, to *time.Time) string {
	var sb strings.Builder

	for _, acc := range accounts {
		ab := accountBalances[acc.Name]
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

func (f *Formatter) FormatBalancesMarkdown(accounts []*entity.Account, accountBalances map[string]*AccountBalance, from, to *time.Time) string {
	var sb strings.Builder

	sb.WriteString("# Balances\n\n")

	for _, acc := range accounts {
		ab := accountBalances[acc.Name]
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

func (f *Formatter) getCategoryDisplayName(cat *entity.Category) string {
	if cat == nil {
		return ""
	}
	if cat.Emoji != nil && *cat.Emoji != "" {
		return *cat.Emoji + " " + cat.Name
	}
	return cat.Name
}
