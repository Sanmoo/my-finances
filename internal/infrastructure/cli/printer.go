package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type Printer struct{}

func NewPrinter() *Printer {
	return &Printer{}
}

func tagNames(ids []int64, tags map[int64]*entity.Tag) string {
	if len(ids) == 0 {
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

func (p *Printer) PrintAccount(acc *entity.Account) {
	fmt.Printf("Account created: %s (ID: %d)\n", acc.Name, acc.ID)
}

func (p *Printer) PrintCategory(cat *entity.Category) {
	emoji := ""
	if cat.Emoji != nil {
		emoji = *cat.Emoji + " "
	}
	fmt.Printf("Category created: %s%s (ID: %d, Type: %s)\n", emoji, cat.Name, cat.ID, cat.Type)
}

func (p *Printer) PrintCreditCard(cc *entity.CreditCard) {
	fmt.Printf("Credit card created: %s (ID: %d)\n", cc.Name, cc.ID)
	fmt.Printf("  Closing day: %d, Due day: %d\n", cc.ClosingDay, cc.DueDay)
}

func (p *Printer) PrintEntry(entry *entity.Entry, tags map[int64]*entity.Tag) {
	fmt.Printf("Entry created (ID: %d)\n", entry.ID)
	fmt.Printf("  Type: %s, Amount: %.2f %s\n", entry.Type, entry.Amount, entry.Currency)
	fmt.Printf("  Date: %s\n", entry.RealizationDate.Format("02-01-2006"))
	if entry.PaymentDate != nil {
		fmt.Printf("  Payment Date: %s\n", entry.PaymentDate.Format("02-01-2006"))
	}
	if entry.Description != "" {
		fmt.Printf("  Description: %s\n", entry.Description)
	}
	if len(entry.TagIDs) > 0 {
		fmt.Printf("  Tags: %s\n", tagNames(entry.TagIDs, tags))
	}
}

func (p *Printer) PrintEntryWithCategory(entry *entity.Entry, category *entity.Category, tags map[int64]*entity.Tag) {
	fmt.Printf("Entry created (ID: %d)\n", entry.ID)
	fmt.Printf("  Type: %s, Amount: %.2f %s\n", entry.Type, entry.Amount, entry.Currency)
	fmt.Printf("  Date: %s\n", entry.RealizationDate.Format("02-01-2006"))
	if entry.PaymentDate != nil {
		fmt.Printf("  Payment Date: %s\n", entry.PaymentDate.Format("02-01-2006"))
	}
	if category != nil {
		emoji := ""
		if category.Emoji != nil {
			emoji = *category.Emoji + " "
		}
		fmt.Printf("  Category: %s%s\n", emoji, category.Name)
	}
	if entry.Description != "" {
		fmt.Printf("  Description: %s\n", entry.Description)
	}
	if len(entry.TagIDs) > 0 {
		fmt.Printf("  Tags: %s\n", tagNames(entry.TagIDs, tags))
	}
}

func (p *Printer) PrintEntryMarkdown(entry *entity.Entry, category *entity.Category, tags map[int64]*entity.Tag) {
	emoji := ""
	if category != nil && category.Emoji != nil {
		emoji = *category.Emoji + " "
	}

	dateStr := entry.RealizationDate.Format("02/01/2006")
	paymentStr := ""
	if entry.PaymentDate != nil {
		paymentStr = fmt.Sprintf(" → %s", entry.PaymentDate.Format("02/01/2006"))
	}

	descStr := ""
	if entry.Description != "" {
		descStr = fmt.Sprintf(" — _%s_", entry.Description)
	}

	tagStr := ""
	if len(entry.TagIDs) > 0 {
		tagStr = fmt.Sprintf(" `[%s]`", tagNames(entry.TagIDs, tags))
	}

	catName := ""
	if category != nil {
		catName = emoji + category.Name
	}

	entryType := "💰"
	if entry.IsExpense() {
		entryType = "💸"
	}

	fmt.Printf("| %s | %s | %s | %.2f |%s%s%s |\n",
		dateStr+paymentStr,
		catName,
		entryType,
		entry.Amount,
		descStr,
		tagStr,
		"",
	)
}

func (p *Printer) PrintReportMarkdown(entries []*entity.Entry, categories map[int64]*entity.Category, tags map[int64]*entity.Tag, from, to *time.Time) {
	fmt.Println("# Financial Report")
	fmt.Println()

	if from != nil || to != nil {
		period := ""
		if from != nil {
			period += fmt.Sprintf("from %s ", from.Format("02/01/2006"))
		}
		if to != nil {
			period += fmt.Sprintf("until %s", to.Format("02/01/2006"))
		}
		fmt.Printf("**Period:** %s\n\n", strings.TrimSpace(period))
	}

	fmt.Println("| Date | Category | Type | Amount |")
	fmt.Println("|------|----------|------|--------|")

	for _, entry := range entries {
		cat := categories[*entry.CategoryID]
		p.PrintEntryMarkdown(entry, cat, tags)
	}

	fmt.Println()
	p.PrintSummary(entries)
}

func (p *Printer) PrintSummary(entries []*entity.Entry) {
	var totalIncome, totalExpense float64

	for _, entry := range entries {
		if entry.IsIncome() {
			totalIncome += entry.Amount
		} else {
			totalExpense += entry.Amount
		}
	}

	fmt.Println("## Summary")
	fmt.Printf("- **Total Income:** %.2f\n", totalIncome)
	fmt.Printf("- **Total Expense:** %.2f\n", totalExpense)
	fmt.Printf("- **Balance:** %.2f\n", totalIncome-totalExpense)
}

func (p *Printer) PrintBalances(namespace string, accounts []*entity.Account, entries []*entity.Entry) {
	fmt.Printf("# Balances for namespace: %s\n\n", namespace)

	balanceByAccount := make(map[int64]float64)

	for _, entry := range entries {
		if entry.IsIncome() {
			balanceByAccount[0] += entry.Amount
		} else {
			if entry.PaymentDate != nil {
				balanceByAccount[0] -= entry.Amount
			} else {
				balanceByAccount[0] -= entry.Amount
			}
		}
	}

	fmt.Println("## Total Balance")
	fmt.Printf("**%.2f**\n", balanceByAccount[0])
}

func (p *Printer) PrintError(msg string) {
	fmt.Printf("Error: %s\n", msg)
}

func (p *Printer) PrintSuccess(msg string) {
	fmt.Printf("Success: %s\n", msg)
}

func (p *Printer) FormatEntryDescription(entry *entity.Entry, totalInstallments int) string {
	desc := entry.Description
	if entry.CreditCardID != nil {
		ccPart := "[CC] (" + entry.RealizationDate.Format("02/01/2006") + ")"
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

func (p *Printer) PrintEntriesTable(entries []*entity.Entry, categories map[int64]*entity.Category, tags map[int64]*entity.Tag, totalInstallments map[int64]int) {
	expenses := make([]*entity.Entry, 0)
	incomes := make([]*entity.Entry, 0)

	for _, entry := range entries {
		if entry.IsExpense() {
			expenses = append(expenses, entry)
		} else {
			incomes = append(incomes, entry)
		}
	}

	if len(expenses) > 0 {
		fmt.Println("=== Expenses ===")
		fmt.Printf("%-12s | %-15s | %-12s | %s\n", "Date", "Category", "Amount", "Description")
		fmt.Println(strings.Repeat("-", 80))

		for _, entry := range expenses {
			dateStr := entry.RealizationDate.Format("02-01-2006")
			catName := ""
			if entry.CategoryID != nil {
				if cat, ok := categories[*entry.CategoryID]; ok {
					catName = cat.Name
				}
			}

			currency := entry.Currency
			amountStr := fmt.Sprintf("%s %.2f", currency, entry.Amount)

			ti := 0
			if entry.ParentEntryID != nil {
				ti = totalInstallments[*entry.ParentEntryID]
			} else {
				ti = totalInstallments[entry.ID]
			}

			desc := p.FormatEntryDescription(entry, ti)

			tagsStr := ""
			if len(entry.TagIDs) > 0 {
				tagsStr = tagNames(entry.TagIDs, tags)
			}

			fmt.Printf("%-12s | %-15s | %-12s | %s\n", dateStr, catName, amountStr, desc)
			if tagsStr != "" {
				fmt.Printf("%-12s   %-15s   %-12s   Tags: %s\n", "", "", "", tagsStr)
			}
		}
		fmt.Println()
	}

	if len(incomes) > 0 {
		fmt.Println("=== Incomes ===")
		fmt.Printf("%-12s | %-15s | %-12s | %s\n", "Date", "Category", "Amount", "Description")
		fmt.Println(strings.Repeat("-", 80))

		for _, entry := range incomes {
			dateStr := entry.RealizationDate.Format("02-01-2006")
			catName := ""
			if entry.CategoryID != nil {
				if cat, ok := categories[*entry.CategoryID]; ok {
					catName = cat.Name
				}
			}

			currency := entry.Currency
			amountStr := fmt.Sprintf("%s %.2f", currency, entry.Amount)

			desc := entry.Description

			tagsStr := ""
			if len(entry.TagIDs) > 0 {
				tagsStr = tagNames(entry.TagIDs, tags)
			}

			fmt.Printf("%-12s | %-15s | %-12s | %s\n", dateStr, catName, amountStr, desc)
			if tagsStr != "" {
				fmt.Printf("%-12s   %-15s   %-12s   Tags: %s\n", "", "", "", tagsStr)
			}
		}
		fmt.Println()
	}
}

func (p *Printer) PrintEntriesMarkdown(entries []*entity.Entry, categories map[int64]*entity.Category, tags map[int64]*entity.Tag, totalInstallments map[int64]int) {
	expenses := make([]*entity.Entry, 0)
	incomes := make([]*entity.Entry, 0)

	for _, entry := range entries {
		if entry.IsExpense() {
			expenses = append(expenses, entry)
		} else {
			incomes = append(incomes, entry)
		}
	}

	if len(expenses) > 0 {
		fmt.Println("## Expenses")
		fmt.Println()
		fmt.Println("| Date | Category | Amount | Description |")
		fmt.Println("|------|----------|--------|-------------|")

		for _, entry := range expenses {
			dateStr := entry.RealizationDate.Format("02/01/2006")
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

			currency := entry.Currency
			amountStr := fmt.Sprintf("%s %.2f", currency, entry.Amount)

			ti := 0
			if entry.ParentEntryID != nil {
				ti = totalInstallments[*entry.ParentEntryID]
			} else {
				ti = totalInstallments[entry.ID]
			}

			desc := p.FormatEntryDescription(entry, ti)

			tagsStr := ""
			if len(entry.TagIDs) > 0 {
				tagsStr = fmt.Sprintf("`[%s]`", tagNames(entry.TagIDs, tags))
			}

			fmt.Printf("| %s | %s | %s | %s %s |\n", dateStr, catName, amountStr, desc, tagsStr)
		}
		fmt.Println()
	}

	if len(incomes) > 0 {
		fmt.Println("## Incomes")
		fmt.Println()
		fmt.Println("| Date | Category | Amount | Description |")
		fmt.Println("|------|----------|--------|-------------|")

		for _, entry := range incomes {
			dateStr := entry.RealizationDate.Format("02/01/2006")
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

			currency := entry.Currency
			amountStr := fmt.Sprintf("%s %.2f", currency, entry.Amount)

			desc := entry.Description

			tagsStr := ""
			if len(entry.TagIDs) > 0 {
				tagsStr = fmt.Sprintf("`[%s]`", tagNames(entry.TagIDs, tags))
			}

			fmt.Printf("| %s | %s | %s | %s %s |\n", dateStr, catName, amountStr, desc, tagsStr)
		}
		fmt.Println()
	}
}

type AccountBalance struct {
	Account      *entity.Account
	TotalIncome  float64
	TotalExpense float64
	Balance      float64
}

func (p *Printer) PrintBalancesTable(accounts []*entity.Account, accountBalances map[int64]*AccountBalance, from, to *time.Time) {
	for _, acc := range accounts {
		ab := accountBalances[acc.ID]
		if ab == nil {
			ab = &AccountBalance{Account: acc}
		}

		fmt.Printf("=== %s ===\n", acc.Name)
		fmt.Println()

		if from != nil || to != nil {
			period := ""
			if from != nil {
				period += from.Format("02/01/2006")
			}
			if from != nil && to != nil {
				period += " - "
			}
			if to != nil {
				period += to.Format("02/01/2006")
			}
			fmt.Printf("Period: %s\n", period)
			fmt.Println()
		}

		fmt.Printf("  Total Income:  %.2f\n", ab.TotalIncome)
		fmt.Printf("  Total Expense: %.2f\n", ab.TotalExpense)
		fmt.Println("  " + strings.Repeat("-", 30))
		fmt.Printf("  Balance:       %.2f\n", ab.Balance)
		fmt.Println()
	}
}

func (p *Printer) PrintBalancesMarkdown(accounts []*entity.Account, accountBalances map[int64]*AccountBalance, from, to *time.Time) {
	fmt.Println("# Balances")
	fmt.Println()

	for _, acc := range accounts {
		ab := accountBalances[acc.ID]
		if ab == nil {
			ab = &AccountBalance{Account: acc}
		}

		fmt.Printf("## %s\n", acc.Name)
		fmt.Println()

		if from != nil || to != nil {
			period := ""
			if from != nil {
				period += from.Format("02/01/2006")
			}
			if from != nil && to != nil {
				period += " - "
			}
			if to != nil {
				period += to.Format("02/01/2006")
			}
			fmt.Printf("**Period:** %s\n", period)
			fmt.Println()
		}

		fmt.Println("| | Amount |")
		fmt.Println("|---|---:|")
		fmt.Printf("| Total Income | %.2f |\n", ab.TotalIncome)
		fmt.Printf("| Total Expense | %.2f |\n", ab.TotalExpense)
		fmt.Printf("| **Balance** | **%.2f** |\n", ab.Balance)
		fmt.Println()
	}
}
