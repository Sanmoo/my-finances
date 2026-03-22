package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type Printer struct {
	formatter *Formatter
	output    io.Writer
}

func NewPrinter() *Printer {
	return &Printer{
		formatter: NewFormatter(),
		output:    os.Stdout,
	}
}

func NewPrinterWithOutput(w io.Writer) *Printer {
	return &Printer{
		formatter: NewFormatter(),
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

func (p *Printer) PrintEntry(entry *entity.Entry, tags map[int64]*entity.Tag) {
	fmt.Fprintln(p.output, p.formatter.FormatEntry(entry, tags))
}

func (p *Printer) PrintEntryWithCategory(entry *entity.Entry, category *entity.Category, tags map[int64]*entity.Tag) {
	fmt.Fprintln(p.output, p.formatter.FormatEntryWithCategory(entry, category, tags))
}

func (p *Printer) PrintEntriesTable(entries []*entity.Entry, categories map[int64]*entity.Category, tags map[int64]*entity.Tag, totalInstallments map[int64]int) {
	fmt.Fprint(p.output, p.formatter.FormatEntriesTable(entries, categories, tags, totalInstallments))
}

func (p *Printer) PrintEntriesMarkdown(entries []*entity.Entry, categories map[int64]*entity.Category, tags map[int64]*entity.Tag, totalInstallments map[int64]int) {
	fmt.Fprint(p.output, p.formatter.FormatEntriesMarkdown(entries, categories, tags, totalInstallments))
}

func (p *Printer) PrintReportMarkdown(entries []*entity.Entry, categories map[int64]*entity.Category, tags map[int64]*entity.Tag, from, to *time.Time) {
	fmt.Fprint(p.output, p.formatter.FormatReportMarkdown(entries, categories, tags, from, to))
}

func (p *Printer) PrintBalancesTable(accounts []*entity.Account, accountBalances map[int64]*AccountBalance, from, to *time.Time) {
	fmt.Fprint(p.output, p.formatter.FormatBalancesTable(accounts, accountBalances, from, to))
}

func (p *Printer) PrintBalancesMarkdown(accounts []*entity.Account, accountBalances map[int64]*AccountBalance, from, to *time.Time) {
	fmt.Fprint(p.output, p.formatter.FormatBalancesMarkdown(accounts, accountBalances, from, to))
}

func (p *Printer) PrintError(msg string) {
	fmt.Fprintln(p.output, p.formatter.FormatError(msg))
}

func (p *Printer) PrintSuccess(msg string) {
	fmt.Fprintln(p.output, p.formatter.FormatSuccess(msg))
}
