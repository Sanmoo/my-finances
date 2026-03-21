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

func (p *Printer) PrintEntry(entry *entity.Entry) {
	fmt.Printf("Entry created (ID: %d)\n", entry.ID)
	fmt.Printf("  Type: %s, Amount: %.2f %s\n", entry.Type, entry.Amount, entry.Currency)
	fmt.Printf("  Date: %s\n", entry.RealizationDate.Format("02-01-2006"))
	if entry.PaymentDate != nil {
		fmt.Printf("  Payment Date: %s\n", entry.PaymentDate.Format("02-01-2006"))
	}
	if entry.Description != "" {
		fmt.Printf("  Description: %s\n", entry.Description)
	}
	if len(entry.Tags) > 0 {
		fmt.Printf("  Tags: %s\n", strings.Join(entry.Tags, ", "))
	}
}

func (p *Printer) PrintEntryMarkdown(entry *entity.Entry, category *entity.Category) {
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
	if len(entry.Tags) > 0 {
		tagStr = fmt.Sprintf(" `[%s]`", strings.Join(entry.Tags, "`, `"))
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

func (p *Printer) PrintReportMarkdown(entries []*entity.Entry, categories map[int64]*entity.Category, from, to *time.Time) {
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
		p.PrintEntryMarkdown(entry, cat)
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
