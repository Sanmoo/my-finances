package interactive

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/core/usecase"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"github.com/Sanmoo/my-finances/internal/infrastructure/cli"
	"github.com/Sanmoo/my-finances/pkg/expr"
)

// Special constants for selector options.
const (
	selectorNewAccount  = "+ Criar nova conta"
	selectorNone        = "(nenhum(a))"
	selectorNewCategory = "+ Criar nova categoria"
	selectorNewTag      = "+ Criar nova tag"
	selectorNewCard     = "+ Criar novo cartão"

	choiceExpense = "💸 Despesa"
	choiceIncome  = "💰 Receita"
)

var errCancelled = errors.New("cancelled")

// Wizard orchestrates the interactive entry creation flow.
type Wizard struct {
	prompter        Prompter
	selector        Selector
	entryRepo       port.EntriesRepository
	categoryRepo    port.CategoriesRepository
	tagRepo         port.TagsRepository
	ccRepo          port.CreditCardsRepository
	accountRepo     port.AccountsRepository
	defaultCurrency string
	printer         *cli.Printer
}

// NewWizard creates a wizard with all required dependencies.
func NewWizard(
	prompter Prompter,
	selector Selector,
	entryRepo port.EntriesRepository,
	categoryRepo port.CategoriesRepository,
	tagRepo port.TagsRepository,
	ccRepo port.CreditCardsRepository,
	accountRepo port.AccountsRepository,
	defaultCurrency string,
	printer *cli.Printer,
) *Wizard {
	return &Wizard{
		prompter:        prompter,
		selector:        selector,
		entryRepo:       entryRepo,
		categoryRepo:    categoryRepo,
		tagRepo:         tagRepo,
		ccRepo:          ccRepo,
		accountRepo:     accountRepo,
		defaultCurrency: defaultCurrency,
		printer:         printer,
	}
}

// Run starts the interactive loop. It returns nil on normal exit.
func (w *Wizard) Run() error {
	if !isTerminal(os.Stdin) {
		return fmt.Errorf("modo interativo requer um terminal")
	}

	fmt.Println("myfin — modo interativo")
	fmt.Println()

	for {
		err := w.runOneCycle()
		if err != nil {
			if errors.Is(err, errCancelled) {
				fmt.Println("\nCancelado.")
				return nil
			}
			w.printer.PrintError(err.Error())
		}

		fmt.Println()
		again, err := w.prompter.Confirm("Cadastrar novo registro?", true)
		if err != nil {
			return err
		}
		if !again {
			break
		}
		fmt.Println()
	}
	return nil
}

// runOneCycle collects all inputs, confirms, and executes one entry.
func (w *Wizard) runOneCycle() error {
	// Step 1: Type
	entryType, err := w.promptType()
	if err != nil {
		return err
	}

	// Step 2: Account
	accountName, err := w.promptAccount()
	if err != nil {
		return err
	}

	// Step 3: Amount
	amount, err := w.promptAmount()
	if err != nil {
		return err
	}

	// Step 4: Date
	date, err := w.promptDate()
	if err != nil {
		return err
	}

	// Step 5: Description
	description, err := w.promptDescription()
	if err != nil {
		return err
	}

	// Step 6: Category
	categoryAlias, err := w.promptCategory(accountName, entryType)
	if err != nil {
		return err
	}

	// Step 7: Tags
	tags, err := w.promptTags()
	if err != nil {
		return err
	}

	// Step 8: Credit card (expense only)
	var creditCardName string
	var times int
	if entryType == entity.EntryTypeExpense {
		cc, t, err := w.promptCreditCard()
		if err != nil {
			return err
		}
		creditCardName = cc
		times = t
	}

	// Build command for display
	cmd := RenderCLI(entryType, amount, accountName, date.Format("2006-01-02"), description, categoryAlias, creditCardName, tags, times)

	// Confirmation
	fmt.Println()
	fmt.Println("────────────────────────────────────────")
	fmt.Println("Comando a executar:")
	fmt.Printf("  %s\n", cmd)
	fmt.Println()

	confirmed, err := w.prompter.Confirm("Executar?", true)
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("Descartado.")
		return nil
	}

	// Execute via use case
	return w.execute(entryType, amount, accountName, date, description, categoryAlias, creditCardName, tags, times)
}

// isTerminal checks if fd is a character device (terminal).
func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func (w *Wizard) promptType() (entity.EntryType, error) {
	choice, ok, err := w.selector.Select("Tipo", []string{choiceExpense, choiceIncome})
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errCancelled
	}
	fmt.Printf("Tipo: %s\n", choice)
	if choice == choiceExpense {
		return entity.EntryTypeExpense, nil
	}
	return entity.EntryTypeIncome, nil
}

func (w *Wizard) promptAccount() (string, error) {
	accounts, err := w.accountRepo.GetAll()
	if err != nil {
		return "", fmt.Errorf("erro ao carregar contas: %w", err)
	}

	options := make([]string, 0, len(accounts)+1)
	for _, a := range accounts {
		options = append(options, a.Name)
	}
	options = append(options, selectorNewAccount)

	choice, ok, err := w.selector.Select("Conta", options)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errCancelled
	}

	if choice == selectorNewAccount {
		name, err := w.inlineAddAccount()
		if err != nil {
			return "", err
		}
		fmt.Printf("Conta: %s\n", name)
		return name, nil
	}
	fmt.Printf("Conta: %s\n", choice)
	return choice, nil
}

func (w *Wizard) promptAmount() (string, error) {
	for {
		input, err := w.prompter.Text("Valor (ex.: 1000/3)", "")
		if err != nil {
			return "", err
		}
		if input == "" {
			fmt.Println("O valor é obrigatório.")
			continue
		}
		// Validate that the expression parses
		_, parseErr := evalAmount(input)
		if parseErr != nil {
			fmt.Printf("Valor inválido: %v\n", parseErr)
			continue
		}
		return input, nil
	}
}

func (w *Wizard) promptDate() (time.Time, error) {
	today := time.Now().Format("2006-01-02")
	for {
		input, err := w.prompter.Text("Data", today)
		if err != nil {
			return time.Time{}, err
		}
		date, err := cli.ParseRealizationDate(input, time.Now())
		if err != nil {
			if errors.Is(err, cli.ErrImpossibleRollback) {
				fmt.Printf("Dia %s não existe no mês anterior. Use MM-DD ou YYYY-MM-DD.\n", strings.TrimSpace(input))
			} else {
				fmt.Println("Data inválida. Use DD, MM-DD, YY-MM-DD ou YYYY-MM-DD.")
			}
			continue
		}
		fmt.Printf("Data: %s\n", date.Format("2006-01-02"))
		return date, nil
	}
}

func (w *Wizard) promptDescription() (string, error) {
	return w.prompter.Text("Descrição", "")
}

func (w *Wizard) promptCategory(accountName string, entryType entity.EntryType) (string, error) {
	categories, err := w.categoryRepo.GetAll(accountName)
	if err != nil {
		return "", fmt.Errorf("erro ao carregar categorias: %w", err)
	}

	wantType := entity.CategoryTypeExpense
	if entryType == entity.EntryTypeIncome {
		wantType = entity.CategoryTypeIncome
	}

	options := []string{selectorNone, selectorNewCategory}
	for _, c := range categories {
		if c.Type == wantType {
			options = append(options, c.Alias)
		}
	}

	choice, ok, err := w.selector.Select("Categoria", options)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errCancelled
	}

	switch choice {
	case selectorNone:
		fmt.Printf("Categoria: %s\n", selectorNone)
		return "", nil
	case selectorNewCategory:
		alias, err := w.inlineAddCategory(accountName, wantType)
		if err != nil {
			return "", err
		}
		fmt.Printf("Categoria: %s\n", alias)
		return alias, nil
	default:
		fmt.Printf("Categoria: %s\n", choice)
		return choice, nil
	}
}

func (w *Wizard) promptTags() ([]string, error) {
	allTags, err := w.tagRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar tags: %w", err)
	}

	options := []string{selectorNone, selectorNewTag}
	for _, t := range allTags {
		options = append(options, t.Name)
	}

	selected, err := w.selector.MultiSelect("Tags", options)
	if err != nil {
		return nil, err
	}

	// If "none" was selected, ignore everything else and return empty
	for _, s := range selected {
		if s == selectorNone {
			fmt.Printf("Tags: %s\n", selectorNone)
			return []string{}, nil
		}
	}

	var tags []string
	for _, s := range selected {
		if s == selectorNewTag {
			name, err := w.inlineAddTag()
			if err != nil {
				return nil, err
			}
			tags = append(tags, name)
		} else {
			tags = append(tags, s)
		}
	}
	if len(tags) == 0 {
		fmt.Printf("Tags: %s\n", selectorNone)
	} else {
		fmt.Printf("Tags: %s\n", strings.Join(tags, ", "))
	}
	return tags, nil
}

func (w *Wizard) promptCreditCard() (string, int, error) {
	cards, err := w.ccRepo.GetAll()
	if err != nil {
		return "", 0, fmt.Errorf("erro ao carregar cartões: %w", err)
	}

	options := []string{selectorNone, selectorNewCard}
	for _, cc := range cards {
		options = append(options, cc.Name)
	}

	choice, ok, err := w.selector.Select("Cartão de crédito", options)
	if err != nil {
		return "", 0, err
	}
	if !ok {
		return "", 0, errCancelled
	}

	switch choice {
	case selectorNone:
		fmt.Printf("Cartão de crédito: %s\n", selectorNone)
		return "", 1, nil
	case selectorNewCard:
		name, err := w.inlineAddCreditCard()
		if err != nil {
			return "", 0, err
		}
		choice = name
	}
	fmt.Printf("Cartão de crédito: %s\n", choice)

	// Prompt for times
	times, err := w.promptTimes()
	if err != nil {
		return "", 0, err
	}
	return choice, times, nil
}

func (w *Wizard) promptTimes() (int, error) {
	for {
		input, err := w.prompter.Text("Parcelas", "1")
		if err != nil {
			return 0, err
		}
		if input == "" {
			return 1, nil
		}
		var n int
		if _, scanErr := fmt.Sscanf(input, "%d", &n); scanErr != nil || n < 1 {
			fmt.Println("Número de parcelas inválido. Informe um número >= 1.")
			continue
		}
		return n, nil
	}
}

func (w *Wizard) execute(
	entryType entity.EntryType,
	amount string,
	accountName string,
	date time.Time,
	description string,
	categoryAlias string,
	creditCardName string,
	tags []string,
	times int,
) error {
	uc := usecase.NewAddEntry(w.entryRepo, w.categoryRepo, w.tagRepo, w.ccRepo)

	result, err := uc.Execute(usecase.AddEntryInput{
		Type:           entryType,
		Amount:         amount,
		Currency:       w.defaultCurrency,
		Description:    description,
		CategoryAlias:  categoryAlias,
		CreditCardName: creditCardName,
		Tags:           tags,
		Times:          times,
		Date:           date,
		AccountName:    accountName,
	})
	if err != nil {
		return err
	}

	// Print output (same format as CLI)
	for _, entry := range result.Entries {
		w.printer.PrintEntryWithCategory(entry, result.Category, accountName)
	}

	return nil
}

// evalAmount validates a math expression.
func evalAmount(exprStr string) (float64, error) {
	return expr.Parse(exprStr)
}
