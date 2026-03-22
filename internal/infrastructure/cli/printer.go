package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"github.com/Sanmoo/my-finances/internal/infrastructure/config"
	"github.com/Sanmoo/my-finances/internal/infrastructure/i18n"
)

type Printer struct {
	formatter *Formatter
	output    io.Writer
}

func NewPrinter() *Printer {
	cfg, err := config.NewLoader().Load()
	if err != nil {
		cfg = &config.Config{Locale: "pt-BR"}
	}
	if cfg.Locale == "" {
		cfg.Locale = "pt-BR"
	}
	return &Printer{
		formatter: NewFormatter(i18n.New(cfg.Locale)),
		output:    os.Stdout,
	}
}

func NewPrinterWithLocale(locale string) *Printer {
	if locale == "" {
		locale = "pt-BR"
	}
	return &Printer{
		formatter: NewFormatter(i18n.New(locale)),
		output:    os.Stdout,
	}
}

func NewPrinterWithOutput(w io.Writer) *Printer {
	cfg, err := config.NewLoader().Load()
	if err != nil {
		cfg = &config.Config{Locale: "pt-BR"}
	}
	if cfg.Locale == "" {
		cfg.Locale = "pt-BR"
	}
	return &Printer{
		formatter: NewFormatter(i18n.New(cfg.Locale)),
		output:    w,
	}
}

func (p *Printer) PrintAccount(acc *entity.Account) {
	fmt.Fprintln(p.output, p.formatter.FormatAccount(acc))
}

func (p *Printer) PrintCategory(cat *entity.Category) {
	fmt.Fprintln(p.output, p.formatter.FormatCategory(cat))
}

func (p *Printer) PrintCreditCard(cc *entity.CreditCard) {
	fmt.Fprintln(p.output, p.formatter.FormatCreditCard(cc))
}

func (p *Printer) PrintEntry(entry *entity.Entry) {
	fmt.Fprintln(p.output, p.formatter.FormatEntry(entry))
}

func (p *Printer) PrintEntryWithCategory(entry *entity.Entry, category *entity.Category) {
	fmt.Fprintln(p.output, p.formatter.FormatEntryWithCategory(entry, category))
}

func (p *Printer) PrintEntriesTable(entries []*entity.Entry, categories map[string]*entity.Category, accounts map[string]*entity.Account, filteredAccount string) {
	fmt.Fprint(p.output, p.formatter.FormatEntriesTable(entries, categories, accounts, filteredAccount))
}

func (p *Printer) PrintEntriesMarkdown(entries []*entity.Entry, categories map[string]*entity.Category, accounts map[string]*entity.Account, filteredAccount string) {
	fmt.Fprint(p.output, p.formatter.FormatEntriesMarkdown(entries, categories, accounts, filteredAccount))
}

func (p *Printer) PrintReportMarkdown(entries []*entity.Entry, categories map[string]*entity.Category, from, to *time.Time) {
	fmt.Fprint(p.output, p.formatter.FormatReportMarkdown(entries, categories, from, to))
}

func (p *Printer) PrintBalancesTable(accounts []*entity.Account, accountBalances map[string]*AccountBalance, from, to *time.Time) {
	fmt.Fprint(p.output, p.formatter.FormatBalancesTable(accounts, accountBalances, from, to))
}

func (p *Printer) PrintBalancesMarkdown(accounts []*entity.Account, accountBalances map[string]*AccountBalance, from, to *time.Time) {
	fmt.Fprint(p.output, p.formatter.FormatBalancesMarkdown(accounts, accountBalances, from, to))
}

func (p *Printer) PrintError(msg string) {
	fmt.Fprintln(p.output, p.formatter.FormatError(msg))
}

func (p *Printer) PrintSuccess(msg string) {
	fmt.Fprintln(p.output, p.formatter.FormatSuccess(msg))
}
