package usecase

import (
	"fmt"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

// PreparedStatement bundles the statement report with the resolved
// Categories, so callers never coordinate category loading themselves.
type PreparedStatement struct {
	Statement *StatementOutput
	// Categories is the legacy map keyed by alias: all accounts in repository
	// order, last account's category wins on alias collisions, read errors
	// ignored. See Report.Execute for the full compatibility notes.
	Categories map[string]*entity.Category
}

// ReportPreparation is the application seam consumed by the report callers
// (and by the tests): it concentrates the coordination needed to render
// `report entries`, `report by-category` and `report statement` — querying
// entries, resolving Categories across accounts and computing totals.
//
// Callers are responsible for flags/dates/configuration, wiring the adapters
// into NewReportPreparation, invoking the preparation and choosing the output
// format. They never rebuild entries nor coordinate repository reads.
type ReportPreparation interface {
	PrepareReport(input ReportInput) (*ReportOutput, error)
	PrepareStatement(input StatementInput) (*PreparedStatement, error)
}

type preparation struct {
	report       *Report
	statement    *Statement
	categoryRepo port.CategoriesRepository
	accountRepo  port.AccountsRepository
}

// NewReportPreparation wires the preparation seam over the persistence
// adapters (this wiring is the caller's job, not the core's).
func NewReportPreparation(
	entryRepo port.EntriesRepository,
	categoryRepo port.CategoriesRepository,
	accountRepo port.AccountsRepository,
	ccRepo port.CreditCardsRepository,
) ReportPreparation {
	return &preparation{
		report:       NewReport(entryRepo, categoryRepo, accountRepo),
		statement:    NewStatement(entryRepo, ccRepo, accountRepo),
		categoryRepo: categoryRepo,
		accountRepo:  accountRepo,
	}
}

func (p *preparation) PrepareReport(input ReportInput) (*ReportOutput, error) {
	return p.report.Execute(input)
}

// PrepareStatement runs the Statement usecase and then resolves Categories
// the same legacy way as the entries reports.
func (p *preparation) PrepareStatement(input StatementInput) (*PreparedStatement, error) {
	result, err := p.statement.Execute(input)
	if err != nil {
		return nil, err
	}

	accounts, err := p.accountRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	return &PreparedStatement{
		Statement:  result,
		Categories: buildLegacyCategoryMap(p.categoryRepo, accounts),
	}, nil
}
