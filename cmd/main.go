package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/core/usecase"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"github.com/Sanmoo/my-finances/internal/infrastructure/cli"
	"github.com/Sanmoo/my-finances/internal/infrastructure/config"
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

		date := parseDate(dateStr)
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

		date := parseDate(dateStr)
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
	Use:   "entries [--from DD-MM-YY] [--until DD-MM-YY] [--filter-tags x,y] [--filter-categories x,y] [--account name] [--format table|md] [--by-realization]",
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
			t := parseDate(fromStr)
			from = &t
		}
		if untilStr != "" {
			t := parseDate(untilStr)
			until = &t
		}

		filterTags := parseCommaSeparated(filterTagsStr)
		filterCategories := parseCommaSeparated(filterCategoriesStr)

		entryRepo := factory.NewEntriesRepository()
		categoryRepo := factory.NewCategoriesRepository()
		accountRepo := factory.NewAccountsRepository()

		report := usecase.NewReport(entryRepo, categoryRepo, accountRepo)

		result, err := report.Execute(usecase.ReportInput{
			Format:                  format,
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

		categoryMap := make(map[string]*entity.Category)
		for _, acc := range result.Accounts {
			categories, err := categoryRepo.GetAll(acc.Name)
			if err != nil {
				continue
			}
			for _, cat := range categories {
				categoryMap[cat.Alias] = cat
			}
		}

		entries := make([]*entity.Entry, 0)
		for _, e := range result.Entries {
			entry := &entity.Entry{
				Type:              entity.EntryType(e.Type),
				Amount:            e.Amount,
				Currency:          e.Currency,
				Description:       e.Description,
				CategoryAlias:     e.CategoryAlias,
				CreditCardName:    e.CreditCardName,
				Tags:              e.Tags,
				InstallmentNumber: e.InstallmentNumber,
				InstallmentTotal:  e.InstallmentTotal,
				RealizationDate:   e.RealizationDate,
				PaymentDate:       e.PaymentDate,
			}
			entries = append(entries, entry)
		}

		accounts := make(map[string]*entity.Account)
		for _, acc := range result.Accounts {
			accounts[acc.Name] = acc
		}

		if format == "md" {
			printer.PrintEntriesMarkdown(entries, categoryMap, accounts, accountStr)
		} else {
			printer.PrintEntriesTable(entries, categoryMap, accounts, accountStr)
		}
	},
}

var reportBalancesCmd = &cobra.Command{
	Use:   "balances [--account name] [--from DD-MM-YY] [--until DD-MM-YY] [--format table|md]",
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
			t := parseDate(fromStr)
			from = &t
		}

		var until *time.Time
		if untilStr != "" {
			t := parseDate(untilStr)
			until = &t
		}

		entryRepo := factory.NewEntriesRepository()
		accountRepo := factory.NewAccountsRepository()

		var accounts []*entity.Account
		if accountStr != "" {
			account, err := accountRepo.GetByName(accountStr)
			if err != nil {
				printer.PrintError(err.Error())
				return
			}
			if account == nil {
				printer.PrintError("account not found: " + accountStr)
				return
			}
			accounts = []*entity.Account{account}
		} else {
			accounts, err = accountRepo.GetAll()
			if err != nil {
				printer.PrintError(err.Error())
				return
			}
		}

		accountBalances := make(map[string]*cli.AccountBalance)

		for _, acc := range accounts {
			// Get entries within the filtered period (respecting --from and --until)
			entries, err := entryRepo.GetAll(&port.EntryFilters{
				FromDate:    from,
				ToDate:      until,
				AccountName: acc.Name,
			})
			if err != nil {
				printer.PrintError(err.Error())
				return
			}

			var totalIncome, totalExpense float64
			for _, entry := range entries {
				if entry.IsIncome() {
					totalIncome += entry.Amount
				} else {
					totalExpense += entry.Amount
				}
			}

			// Get accumulated entries up to --until (ignoring --from)
			accumulatedEntries, err := entryRepo.GetAll(&port.EntryFilters{
				ToDate:      until,
				AccountName: acc.Name,
			})
			if err != nil {
				printer.PrintError(err.Error())
				return
			}

			var accumulatedBalance float64
			for _, entry := range accumulatedEntries {
				if entry.IsIncome() {
					accumulatedBalance += entry.Amount
				} else {
					accumulatedBalance -= entry.Amount
				}
			}

			accountBalances[acc.Name] = &cli.AccountBalance{
				Account:            acc,
				TotalIncome:        totalIncome,
				TotalExpense:       totalExpense,
				Balance:            totalIncome - totalExpense,
				AccumulatedBalance: accumulatedBalance,
			}
		}

		if format == "md" {
			printer.PrintBalancesMarkdown(accounts, accountBalances, from, until)
		} else {
			printer.PrintBalancesTable(accounts, accountBalances, from, until)
		}
	},
}

var reportByCategoryCmd = &cobra.Command{
	Use:   "by-category [--from DD-MM-YY] [--until DD-MM-YY] [--account name] [--format table|md] [--by-realization]",
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
			t := parseDate(fromStr)
			from = &t
		}
		if untilStr != "" {
			t := parseDate(untilStr)
			until = &t
		}

		entryRepo := factory.NewEntriesRepository()
		categoryRepo := factory.NewCategoriesRepository()
		accountRepo := factory.NewAccountsRepository()

		report := usecase.NewReport(entryRepo, categoryRepo, accountRepo)

		result, err := report.Execute(usecase.ReportInput{
			Format:                  format,
			From:                    from,
			To:                      until,
			AccountName:             accountStr,
			FilterByRealizationDate: byRealization,
		})
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		categoryMap := make(map[string]*entity.Category)
		for _, acc := range result.Accounts {
			categories, err := categoryRepo.GetAll(acc.Name)
			if err != nil {
				continue
			}
			for _, cat := range categories {
				categoryMap[cat.Alias] = cat
			}
		}

		entries := make([]*entity.Entry, 0)
		for _, e := range result.Entries {
			entry := &entity.Entry{
				Type:              entity.EntryType(e.Type),
				Amount:            e.Amount,
				Currency:          e.Currency,
				Description:       e.Description,
				CategoryAlias:     e.CategoryAlias,
				CreditCardName:    e.CreditCardName,
				Tags:              e.Tags,
				InstallmentNumber: e.InstallmentNumber,
				InstallmentTotal:  e.InstallmentTotal,
				RealizationDate:   e.RealizationDate,
				PaymentDate:       e.PaymentDate,
			}
			entries = append(entries, entry)
		}

		accounts := make(map[string]*entity.Account)
		for _, acc := range result.Accounts {
			accounts[acc.Name] = acc
		}

		if format == "md" {
			printer.PrintEntriesByCategoryMarkdown(entries, categoryMap, accounts, accountStr)
		} else {
			printer.PrintEntriesByCategoryTable(entries, categoryMap, accounts, accountStr)
		}
	},
}

func init() {
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

	reportEntriesCmd.Flags().String("from", "", "start date (DD-MM-YY)")
	reportEntriesCmd.Flags().String("until", "", "end date (DD-MM-YY)")
	reportEntriesCmd.Flags().String("filter-tags", "", "filter by tags (comma-separated)")
	reportEntriesCmd.Flags().String("filter-categories", "", "filter by categories (comma-separated)")
	reportEntriesCmd.Flags().String("account", "", "account name")
	reportEntriesCmd.Flags().String("format", "table", "output format (table or md)")
	reportEntriesCmd.Flags().Bool("by-realization", false, "filter by realization date instead of payment date")

	reportBalancesCmd.Flags().String("from", "", "start date (DD-MM-YY)")
	reportBalancesCmd.Flags().String("until", "", "end date (DD-MM-YY)")
	reportBalancesCmd.Flags().String("account", "", "account name")
	reportBalancesCmd.Flags().String("format", "table", "output format (table or md)")

	reportByCategoryCmd.Flags().String("from", "", "start date (DD-MM-YY)")
	reportByCategoryCmd.Flags().String("until", "", "end date (DD-MM-YY)")
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

func parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}

	dateStr = strings.TrimSpace(dateStr)
	now := time.Now()

	// Format: YYYY-MM-DD
	if t, err := time.Parse("2006-01-02", dateStr); err == nil {
		return t.UTC()
	}

	// Format: YY-MM-DD
	if t, err := time.Parse("06-01-02", dateStr); err == nil {
		if t.Year() < 100 {
			t = t.AddDate(2000, 0, 0)
		}
		return t.UTC()
	}

	// Format: MM-DD (use current year)
	if t, err := time.Parse("01-02", dateStr); err == nil {
		return time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	}

	// Format: DD (use current month and year)
	if t, err := time.Parse("2", dateStr); err == nil {
		return time.Date(now.Year(), now.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	}

	return time.Time{}
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
