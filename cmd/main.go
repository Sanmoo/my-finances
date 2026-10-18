package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-finances/internal/core/usecase"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"github.com/Sanmoo/my-finances/internal/infrastructure/cli"
	"github.com/Sanmoo/my-finances/internal/infrastructure/config"
	"github.com/Sanmoo/my-finances/internal/infrastructure/interactive"
	"github.com/Sanmoo/my-finances/internal/infrastructure/persistence"
)

var (
	printer   = cli.NewPrinter()
	cfgLoader = config.NewLoader()
	repo      *persistence.RepositoryFactory
)

func main() {
	rootCmd.SetArgs(normalizeNegativeAmountArgs(os.Args[1:]))
	if err := rootCmd.Execute(); err != nil {
		printer.PrintError(err.Error())
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "myfin",
	Short: "A personal finance management CLI tool",
}

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Start interactive wizard for adding entries",
	RunE:  runInteractive,
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new resource",
}

var addAccountCmd = &cobra.Command{
	Use:   "account <name>",
	Short: "Add a new account",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		factory, err := getFactory()
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		repo := factory.NewAccountsRepository()
		uc := usecase.NewAddAccount(repo)

		result, err := uc.Execute(usecase.AddAccountInput{
			Name: args[0],
		})
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		printer.PrintAccount(result.Account)
	},
}

var addCategoryCmd = &cobra.Command{
	Use:   "category <name> --account <account> --type <inc|exp> --alias <alias> [--emoji <emoji>]",
	Short: "Add a new category",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		factory, err := getFactory()
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		accountStr, _ := cmd.Flags().GetString("account")
		if accountStr == "" {
			printer.PrintError("--account is required")
			return
		}

		alias, _ := cmd.Flags().GetString("alias")
		if alias == "" {
			printer.PrintError("--alias is required")
			return
		}

		catType, _ := cmd.Flags().GetString("type")
		emoji, _ := cmd.Flags().GetString("emoji")

		repo := factory.NewCategoriesRepository()
		uc := usecase.NewAddCategory(repo)

		result, err := uc.Execute(usecase.AddCategoryInput{
			AccountName: accountStr,
			Name:        args[0],
			Type:        entity.CategoryType(catType),
			Alias:       alias,
			Emoji:       emoji,
		})
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		printer.PrintCategory(result.Category)
	},
}

var addCreditCardCmd = &cobra.Command{
	Use:   "credit-card <name> --closing-day <n> --due-day <n>",
	Short: "Add a new credit card",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		factory, err := getFactory()
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		closingDay, _ := cmd.Flags().GetInt("closing-day")
		dueDay, _ := cmd.Flags().GetInt("due-day")

		if closingDay == 0 || dueDay == 0 {
			printer.PrintError("closing-day and due-day are required")
			return
		}

		repo := factory.NewCreditCardsRepository()
		uc := usecase.NewAddCreditCard(repo)

		result, err := uc.Execute(usecase.AddCreditCardInput{
			Name:       args[0],
			ClosingDay: closingDay,
			DueDay:     dueDay,
		})
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		printer.PrintCreditCard(result.CreditCard)
	},
}

var addExpenseCmd = &cobra.Command{
	Use:   "expense [amount] --account <name> --date <YYYY-MM-DD> --description <text> [--tags x,y] [--category <alias>] [--credit-card <name>] [--times n]",
	Short: "Add a new expense",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		factory, err := getFactory()
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		amount := "0"
		if len(args) > 0 {
			amount = args[0]
		}

		accountStr, _ := cmd.Flags().GetString("account")
		tagsStr, _ := cmd.Flags().GetString("tags")
		dateStr, _ := cmd.Flags().GetString("date")
		categoryStr, _ := cmd.Flags().GetString("category")

		if dateStr == "" {
			printer.PrintError("--date is required")
			return
		}
		description, _ := cmd.Flags().GetString("description")
		creditCardStr, _ := cmd.Flags().GetString("credit-card")
		times, _ := cmd.Flags().GetInt("times")

		if accountStr == "" {
			printer.PrintError("--account is required")
			return
		}

		date, err := cli.ParseRealizationDate(dateStr, time.Now())
		if err != nil {
			printer.PrintError(err.Error())
			return
		}
		currency := getDefaultCurrency()

		entryRepo := factory.NewEntriesRepository()
		categoryRepo := factory.NewCategoriesRepository()
		tagRepo := factory.NewTagsRepository()
		ccRepo := factory.NewCreditCardsRepository()

		tags := parseCommaSeparated(tagsStr)
		uc := usecase.NewAddEntry(entryRepo, categoryRepo, tagRepo, ccRepo)

		result, err := uc.Execute(usecase.AddEntryInput{
			Type:           entity.EntryTypeExpense,
			Amount:         amount,
			Currency:       currency,
			Description:    description,
			CategoryAlias:  categoryStr,
			CreditCardName: creditCardStr,
			Tags:           tags,
			Times:          times,
			Date:           date,
			AccountName:    accountStr,
		})
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		for _, entry := range result.Entries {
			printer.PrintEntryWithCategory(entry, result.Category, accountStr)
		}
	},
}

var addIncomeCmd = &cobra.Command{
	Use:   "income [amount] --account <name> --date <YYYY-MM-DD> --description <text> [--category <alias>] [--tags x,y]",
	Short: "Add a new income",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		factory, err := getFactory()
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		amount := "0"
		if len(args) > 0 {
			amount = args[0]
		}

		accountStr, _ := cmd.Flags().GetString("account")
		tagsStr, _ := cmd.Flags().GetString("tags")
		dateStr, _ := cmd.Flags().GetString("date")
		categoryStr, _ := cmd.Flags().GetString("category")
		description, _ := cmd.Flags().GetString("description")

		if dateStr == "" {
			printer.PrintError("--date is required")
			return
		}

		if accountStr == "" {
			printer.PrintError("--account is required")
			return
		}

		date, err := cli.ParseRealizationDate(dateStr, time.Now())
		if err != nil {
			printer.PrintError(err.Error())
			return
		}
		currency := getDefaultCurrency()

		entryRepo := factory.NewEntriesRepository()
		categoryRepo := factory.NewCategoriesRepository()
		tagRepo := factory.NewTagsRepository()
		ccRepo := factory.NewCreditCardsRepository()

		tags := parseCommaSeparated(tagsStr)
		uc := usecase.NewAddEntry(entryRepo, categoryRepo, tagRepo, ccRepo)

		result, err := uc.Execute(usecase.AddEntryInput{
			Type:          entity.EntryTypeIncome,
			Amount:        amount,
			Currency:      currency,
			Description:   description,
			CategoryAlias: categoryStr,
			Tags:          tags,
			Date:          date,
			AccountName:   accountStr,
		})
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		for _, entry := range result.Entries {
			printer.PrintEntryWithCategory(entry, result.Category, accountStr)
		}
	},
}

var addTagCmd = &cobra.Command{
	Use:   "tag <name>",
	Short: "Add a new tag",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		factory, err := getFactory()
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		repo := factory.NewTagsRepository()
		uc := usecase.NewAddTag(repo)

		result, err := uc.Execute(usecase.AddTagInput{
			Name: args[0],
		})
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		fmt.Printf("Tag created: %s\n", result.Tag.Name)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List resources",
}

var listTagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List all registered tags",
	Run: func(cmd *cobra.Command, args []string) {
		factory, err := getFactory()
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		repo := factory.NewTagsRepository()
		tags, err := repo.GetAll()
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		if len(tags) == 0 {
			fmt.Println("No tags registered")
			return
		}

		fmt.Println("Registered tags:")
		for _, tag := range tags {
			fmt.Printf("  %s\n", tag.Name)
		}
	},
}

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate reports",
}

var reportEntriesCmd = &cobra.Command{
	Use:   "entries [--from YY-MM-DD] [--until YY-MM-DD] [--filter-tags x,y] [--filter-categories x,y] [--account name] [--format table|md] [--by-realization]",
	Short: "List entries",
	Run: func(cmd *cobra.Command, args []string) {
		factory, err := getFactory()
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		fromStr, _ := cmd.Flags().GetString("from")
		untilStr, _ := cmd.Flags().GetString("until")
		filterTagsStr, _ := cmd.Flags().GetString("filter-tags")
		filterCategoriesStr, _ := cmd.Flags().GetString("filter-categories")
		accountStr, _ := cmd.Flags().GetString("account")
		format, _ := cmd.Flags().GetString("format")
		byRealization, _ := cmd.Flags().GetBool("by-realization")

		var from, until *time.Time
		if fromStr != "" {
			t := cli.ParseDate(fromStr, time.Now())
			from = &t
		}
		if untilStr != "" {
			t := cli.ParseDate(untilStr, time.Now())
			until = &t
		}

		filterTags := parseCommaSeparated(filterTagsStr)
		filterCategories := parseCommaSeparated(filterCategoriesStr)

		entryRepo := factory.NewEntriesRepository()
		categoryRepo := factory.NewCategoriesRepository()
		accountRepo := factory.NewAccountsRepository()
		ccRepo := factory.NewCreditCardsRepository()

		preparation := usecase.NewReportPreparation(entryRepo, categoryRepo, accountRepo, ccRepo)

		result, err := preparation.PrepareReport(usecase.ReportInput{
			From:                    from,
			To:                      until,
			FilterTags:              filterTags,
			FilterCategories:        filterCategories,
			AccountName:             accountStr,
			FilterByRealizationDate: byRealization,
		})
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		if format == "md" {
			printer.PrintEntriesMarkdown(result.Entries, result.Categories, result.AccountsByName, accountStr)
		} else {
			printer.PrintEntriesTable(result.Entries, result.Categories, result.AccountsByName, accountStr)
		}
	},
}

var reportBalancesCmd = &cobra.Command{
	Use:   "balances [--account name] [--from YY-MM-DD] [--until YY-MM-DD] [--format table|md]",
	Short: "Show account balances",
	Run: func(cmd *cobra.Command, args []string) {
		factory, err := getFactory()
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		fromStr, _ := cmd.Flags().GetString("from")
		untilStr, _ := cmd.Flags().GetString("until")
		accountStr, _ := cmd.Flags().GetString("account")
		format, _ := cmd.Flags().GetString("format")

		var from *time.Time
		if fromStr != "" {
			t := cli.ParseDate(fromStr, time.Now())
			from = &t
		}

		var until *time.Time
		if untilStr != "" {
			t := cli.ParseDate(untilStr, time.Now())
			until = &t
		}

		entryRepo := factory.NewEntriesRepository()
		accountRepo := factory.NewAccountsRepository()

		balances, err := usecase.NewBalances(entryRepo, accountRepo).Execute(usecase.BalancesInput{
			AccountName: accountStr,
			From:        from,
			Until:       until,
		})
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		if format == "md" {
			printer.PrintBalancesMarkdown(balances, from, until)
		} else {
			printer.PrintBalancesTable(balances, from, until)
		}
	},
}

var reportStatementCmd = &cobra.Command{
	Use:   "statement --card <name> --month <MM|YY-MM|YYYY-MM> [--account name] [--format table|md]",
	Short: "Show a credit card invoice (fatura) for a month",
	Run: func(cmd *cobra.Command, args []string) {
		factory, err := getFactory()
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		cardStr, _ := cmd.Flags().GetString("card")
		monthStr, _ := cmd.Flags().GetString("month")
		accountStr, _ := cmd.Flags().GetString("account")
		format, _ := cmd.Flags().GetString("format")

		if cardStr == "" {
			printer.PrintError("--card is required")
			return
		}
		if monthStr == "" {
			printer.PrintError("--month is required")
			return
		}
		month, err := parseMonth(monthStr)
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		entryRepo := factory.NewEntriesRepository()
		categoryRepo := factory.NewCategoriesRepository()
		accountRepo := factory.NewAccountsRepository()
		ccRepo := factory.NewCreditCardsRepository()

		preparation := usecase.NewReportPreparation(entryRepo, categoryRepo, accountRepo, ccRepo)

		result, err := preparation.PrepareStatement(usecase.StatementInput{
			CreditCardName: cardStr,
			Year:           month.Year(),
			Month:          month.Month(),
			AccountName:    accountStr,
		})
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		if format == "md" {
			printer.PrintStatementMarkdown(result.Statement, result.Categories)
		} else {
			printer.PrintStatementTable(result.Statement, result.Categories)
		}
	},
}

var reportByCategoryCmd = &cobra.Command{
	Use:   "by-category [--from YY-MM-DD] [--until YY-MM-DD] [--account name] [--format table|md] [--by-realization]",
	Short: "List entries grouped by category",
	Run: func(cmd *cobra.Command, args []string) {
		factory, err := getFactory()
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		fromStr, _ := cmd.Flags().GetString("from")
		untilStr, _ := cmd.Flags().GetString("until")
		accountStr, _ := cmd.Flags().GetString("account")
		format, _ := cmd.Flags().GetString("format")
		byRealization, _ := cmd.Flags().GetBool("by-realization")

		var from, until *time.Time
		if fromStr != "" {
			t := cli.ParseDate(fromStr, time.Now())
			from = &t
		}
		if untilStr != "" {
			t := cli.ParseDate(untilStr, time.Now())
			until = &t
		}

		entryRepo := factory.NewEntriesRepository()
		categoryRepo := factory.NewCategoriesRepository()
		accountRepo := factory.NewAccountsRepository()
		ccRepo := factory.NewCreditCardsRepository()

		preparation := usecase.NewReportPreparation(entryRepo, categoryRepo, accountRepo, ccRepo)

		result, err := preparation.PrepareReport(usecase.ReportInput{
			From:                    from,
			To:                      until,
			AccountName:             accountStr,
			FilterByRealizationDate: byRealization,
		})
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		if format == "md" {
			printer.PrintEntriesByCategoryMarkdown(result.Entries, result.Categories, result.AccountsByName, accountStr)
		} else {
			printer.PrintEntriesByCategoryTable(result.Entries, result.Categories, result.AccountsByName, accountStr)
		}
	},
}

func init() {
	rootCmd.RunE = runInteractive
	rootCmd.AddCommand(interactiveCmd)

	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(reportCmd)

	addCmd.AddCommand(addAccountCmd)
	addCmd.AddCommand(addCategoryCmd)
	addCmd.AddCommand(addCreditCardCmd)
	addCmd.AddCommand(addExpenseCmd)
	addCmd.AddCommand(addIncomeCmd)
	addCmd.AddCommand(addTagCmd)

	rootCmd.AddCommand(listCmd)
	listCmd.AddCommand(listTagsCmd)

	reportCmd.AddCommand(reportEntriesCmd)
	reportCmd.AddCommand(reportBalancesCmd)
	reportCmd.AddCommand(reportStatementCmd)
	reportCmd.AddCommand(reportByCategoryCmd)

	addCategoryCmd.Flags().String("account", "", "account name (required)")
	addCategoryCmd.Flags().StringP("type", "t", "", "category type (inc or exp)")
	addCategoryCmd.Flags().String("alias", "", "category alias (required)")
	addCategoryCmd.Flags().String("emoji", "", "category emoji")

	addCreditCardCmd.Flags().Int("closing-day", 0, "closing day (1-31)")
	addCreditCardCmd.Flags().Int("due-day", 0, "due day (1-31)")

	addExpenseCmd.Flags().StringP("account", "a", "", "account name (required)")
	addExpenseCmd.Flags().StringP("tags", "t", "", "tag names (comma-separated)")
	addExpenseCmd.Flags().StringP("date", "d", "", "date (flexible: DD, MM-DD, YY-MM-DD, YYYY-MM-DD)")
	addExpenseCmd.Flags().StringP("category", "c", "", "category alias")
	addExpenseCmd.Flags().StringP("description", "D", "", "description")
	addExpenseCmd.Flags().StringP("credit-card", "C", "", "credit card name")
	addExpenseCmd.Flags().IntP("times", "n", 0, "number of installments")

	addIncomeCmd.Flags().StringP("account", "a", "", "account name (required)")
	addIncomeCmd.Flags().StringP("tags", "t", "", "tag names (comma-separated)")
	addIncomeCmd.Flags().StringP("date", "d", "", "date (flexible: DD, MM-DD, YY-MM-DD, YYYY-MM-DD)")
	addIncomeCmd.Flags().StringP("category", "c", "", "category alias")
	addIncomeCmd.Flags().StringP("description", "D", "", "description")

	reportEntriesCmd.Flags().String("from", "", "start date (YY-MM-DD)")
	reportEntriesCmd.Flags().String("until", "", "end date (YY-MM-DD)")
	reportEntriesCmd.Flags().String("filter-tags", "", "filter by tags (comma-separated)")
	reportEntriesCmd.Flags().String("filter-categories", "", "filter by categories (comma-separated)")
	reportEntriesCmd.Flags().String("account", "", "account name")
	reportEntriesCmd.Flags().String("format", "table", "output format (table or md)")
	reportEntriesCmd.Flags().Bool("by-realization", false, "filter by realization date instead of payment date")

	reportBalancesCmd.Flags().String("from", "", "start date (YY-MM-DD)")
	reportBalancesCmd.Flags().String("until", "", "end date (YY-MM-DD)")
	reportBalancesCmd.Flags().String("account", "", "account name")
	reportBalancesCmd.Flags().String("format", "table", "output format (table or md)")

	reportStatementCmd.Flags().String("card", "", "credit card name (required)")
	reportStatementCmd.Flags().String("month", "", "invoice month (MM, YY-MM or YYYY-MM)")
	reportStatementCmd.Flags().String("account", "", "account name")
	reportStatementCmd.Flags().String("format", "table", "output format (table or md)")

	reportByCategoryCmd.Flags().String("from", "", "start date (YY-MM-DD)")
	reportByCategoryCmd.Flags().String("until", "", "end date (YY-MM-DD)")
	reportByCategoryCmd.Flags().String("account", "", "account name")
	reportByCategoryCmd.Flags().String("format", "table", "output format (table or md)")
	reportByCategoryCmd.Flags().Bool("by-realization", false, "filter by realization date instead of payment date")
}

func normalizeNegativeAmountArgs(args []string) []string {
	if len(args) == 0 || args[len(args)-1] == "--" || !isNegativeNumber(args[len(args)-1]) {
		return args
	}
	if len(args) >= 2 && args[0] == "add" && args[1] == "expense" {
		normalized := make([]string, 0, len(args)+1)
		normalized = append(normalized, args[:len(args)-1]...)
		normalized = append(normalized, "--", args[len(args)-1])
		return normalized
	}
	return args
}

func isNegativeNumber(value string) bool {
	if len(value) < 2 || value[0] != '-' {
		return false
	}
	_, err := strconv.ParseFloat(value, 64)
	return err == nil
}

func getFactory() (*persistence.RepositoryFactory, error) {
	if repo != nil {
		return repo, nil
	}

	cfg, err := cfgLoader.Load()
	if err != nil {
		return nil, err
	}

	repo = persistence.NewRepositoryFactory(cfg)
	return repo, nil
}

func getDefaultCurrency() string {
	cfg, err := cfgLoader.Load()
	if err != nil || cfg == nil {
		return "BRL"
	}
	if cfg.DefaultCurrency == "" {
		return "BRL"
	}
	return cfg.DefaultCurrency
}

// parseMonth parses a month string in MM, YY-MM or YYYY-MM format and
// returns the first day of that month.
func parseMonth(monthStr string) (time.Time, error) {
	monthStr = strings.TrimSpace(monthStr)
	now := time.Now()

	// Format: YYYY-MM
	if t, err := time.Parse("2006-01", monthStr); err == nil {
		return t.UTC(), nil
	}

	// Format: YY-MM (abbreviated year, always 20YY)
	if t, err := time.Parse("06-01", monthStr); err == nil {
		return cli.ResolveAbbreviatedYear(t), nil
	}

	// Format: MM (use current year)
	if t, err := time.Parse("01", monthStr); err == nil {
		return time.Date(now.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
	}

	return time.Time{}, fmt.Errorf("invalid month: %s (expected MM, YY-MM or YYYY-MM)", monthStr)
}

func runInteractive(cmd *cobra.Command, args []string) error {
	factory, err := getFactory()
	if err != nil {
		return err
	}

	selector, err := interactive.NewFzfSelector(interactive.DefaultFzfOptions())
	if err != nil {
		return err
	}

	prompter := interactive.NewStdioPrompter()

	w := interactive.NewWizard(
		prompter,
		selector,
		factory.NewEntriesRepository(),
		factory.NewCategoriesRepository(),
		factory.NewTagsRepository(),
		factory.NewCreditCardsRepository(),
		factory.NewAccountsRepository(),
		getDefaultCurrency(),
		printer,
	)

	return w.Run()
}

func parseCommaSeparated(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
