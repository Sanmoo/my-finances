package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/core/usecase"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"github.com/Sanmoo/my-finances/internal/infrastructure/cli"
	"github.com/Sanmoo/my-finances/internal/infrastructure/config"
	"github.com/Sanmoo/my-finances/internal/infrastructure/database"
	"github.com/Sanmoo/my-finances/internal/infrastructure/persistence"
	"github.com/Sanmoo/my-finances/internal/infrastructure/persistence/sqlite"
)

var (
	dbFlag      string
	printer     = cli.NewPrinter()
	cfgLoader   = config.NewLoader()
	dbManager   *database.Manager
	repoFactory *persistence.RepositoryFactory
)

func main() {
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
	Use:   "category <name> --type <inc|exp> [--alias <alias>] [--emoji <emoji>]",
	Short: "Add a new category",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		factory, err := getFactory()
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		catType, _ := cmd.Flags().GetString("type")
		alias, _ := cmd.Flags().GetString("alias")
		emoji, _ := cmd.Flags().GetString("emoji")

		repo := factory.NewCategoriesRepository()
		uc := usecase.NewAddCategory(repo)

		result, err := uc.Execute(usecase.AddCategoryInput{
			Name:  args[0],
			Type:  entity.CategoryType(catType),
			Alias: alias,
			Emoji: emoji,
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
	Use:   "expense [amount] --account <name> [--tags x,y] --date <YYYY-MM-DD> [--category x] --description <text> [--credit-card x] [--times n]",
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
		tags, _ := cmd.Flags().GetStringSlice("tags")
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

		accountRepo := factory.NewAccountsRepository()
		account, err := accountRepo.GetByName(accountStr)
		if err != nil {
			printer.PrintError(err.Error())
			return
		}
		if account == nil {
			printer.PrintError("account not found: " + accountStr)
			return
		}

		date := parseDate(dateStr)
		currency := getDefaultCurrency()

		entryRepo := factory.NewEntriesRepository()
		categoryRepo := factory.NewCategoriesRepository()
		ccRepo := factory.NewCreditCardsRepository()
		uc := usecase.NewAddEntry(entryRepo, categoryRepo, ccRepo)

		var cc *entity.CreditCard
		if creditCardStr != "" {
			ccs, _ := ccRepo.GetAll()
			for _, c := range ccs {
				if c.Name == creditCardStr {
					cc = c
					break
				}
			}
			if times <= 0 {
				printer.PrintError("--times is required when using --credit-card")
				return
			}
		}

		if times <= 0 {
			times = 1
		}

		result, err := uc.Execute(usecase.AddEntryInput{
			Type:                entity.EntryTypeExpense,
			Amount:              amount,
			Currency:            currency,
			Description:         description,
			CategoryNameOrAlias: categoryStr,
			CreditCard:          cc,
			Tags:                tags,
			Times:               times,
			Date:                date,
			AccountID:           account.ID,
		})
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		for _, entry := range result.Entries {
			printer.PrintEntryWithCategory(entry, result.Category)
		}
	},
}

var addIncomeCmd = &cobra.Command{
	Use:   "income [amount] --account <name> --date <YYYY-MM-DD> [--category x] --description <text>",
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

		accountRepo := factory.NewAccountsRepository()
		account, err := accountRepo.GetByName(accountStr)
		if err != nil {
			printer.PrintError(err.Error())
			return
		}
		if account == nil {
			printer.PrintError("account not found: " + accountStr)
			return
		}

		date := parseDate(dateStr)
		currency := getDefaultCurrency()

		entryRepo := factory.NewEntriesRepository()
		categoryRepo := factory.NewCategoriesRepository()
		ccRepo := factory.NewCreditCardsRepository()
		uc := usecase.NewAddEntry(entryRepo, categoryRepo, ccRepo)

		result, err := uc.Execute(usecase.AddEntryInput{
			Type:                entity.EntryTypeIncome,
			Amount:              amount,
			Currency:            currency,
			Description:         description,
			CategoryNameOrAlias: categoryStr,
			Date:                date,
			AccountID:           account.ID,
		})
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		for _, entry := range result.Entries {
			printer.PrintEntryWithCategory(entry, result.Category)
		}
	},
}

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate reports",
}

var reportEntriesCmd = &cobra.Command{
	Use:   "entries [--from DD-MM-YY] [--until DD-MM-YY] [--filter-tags x,y] [--filter-categories x,y] [--account name] [--format table|md]",
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

		var accountID *int64
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
			accountID = &account.ID
		}

		report := usecase.NewReport(entryRepo, categoryRepo)

		result, err := report.Execute(usecase.ReportInput{
			Format:           format,
			From:             from,
			To:               until,
			FilterTags:       filterTags,
			FilterCategories: filterCategories,
			AccountID:        accountID,
		})
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		categories, err := categoryRepo.GetAll()
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		categoryMap := make(map[int64]*entity.Category)
		for _, cat := range categories {
			categoryMap[cat.ID] = cat
		}

		entries := make([]*entity.Entry, 0)
		for _, e := range result.Entries {
			entry := &entity.Entry{
				ID:              e.EntryID,
				Type:            entity.EntryType(e.Type),
				Amount:          e.Amount,
				Currency:        e.Currency,
				Description:     e.Description,
				CategoryID:      e.CategoryID,
				CreditCardID:    e.CreditCardID,
				AccountID:       e.AccountID,
				Installment:     e.Installment,
				ParentEntryID:   e.ParentEntryID,
				RealizationDate: e.RealizationDate,
				PaymentDate:     e.PaymentDate,
				Tags:            e.Tags,
			}
			entries = append(entries, entry)
		}

		if format == "md" {
			printer.PrintEntriesMarkdown(entries, categoryMap, result.TotalInstallments)
		} else {
			printer.PrintEntriesTable(entries, categoryMap, result.TotalInstallments)
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

		accountBalances := make(map[int64]*cli.AccountBalance)

		for _, acc := range accounts {
			var accountID *int64
			accountID = &acc.ID

			entries, err := entryRepo.GetAll(&port.EntryFilters{
				FromDate:  from,
				ToDate:    until,
				AccountID: accountID,
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

			accountBalances[acc.ID] = &cli.AccountBalance{
				Account:      acc,
				TotalIncome:  totalIncome,
				TotalExpense: totalExpense,
				Balance:      totalIncome - totalExpense,
			}
		}

		if format == "md" {
			printer.PrintBalancesMarkdown(accounts, accountBalances, from, until)
		} else {
			printer.PrintBalancesTable(accounts, accountBalances, from, until)
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

	reportCmd.AddCommand(reportEntriesCmd)
	reportCmd.AddCommand(reportBalancesCmd)

	rootCmd.PersistentFlags().StringVar(&dbFlag, "db", "", "database name (default or custom)")

	addCategoryCmd.Flags().StringP("type", "t", "", "category type (inc or exp)")
	addCategoryCmd.Flags().String("alias", "", "category alias")
	addCategoryCmd.Flags().String("emoji", "", "category emoji")

	addCreditCardCmd.Flags().Int("closing-day", 0, "closing day (1-31)")
	addCreditCardCmd.Flags().Int("due-day", 0, "due day (1-31)")

	addExpenseCmd.Flags().String("account", "", "account name (required)")
	addExpenseCmd.Flags().StringSlice("tags", []string{}, "tags (comma-separated)")
	addExpenseCmd.Flags().String("date", "", "date (DD-MM-YY)")
	addExpenseCmd.Flags().String("category", "", "category name or alias")
	addExpenseCmd.Flags().String("description", "", "description")
	addExpenseCmd.Flags().String("credit-card", "", "credit card name")
	addExpenseCmd.Flags().Int("times", 0, "number of installments")

	addIncomeCmd.Flags().String("account", "", "account name (required)")
	addIncomeCmd.Flags().String("date", "", "date (DD-MM-YY)")
	addIncomeCmd.Flags().String("category", "", "category name or alias")
	addIncomeCmd.Flags().String("description", "", "description")

	reportEntriesCmd.Flags().String("from", "", "start date (DD-MM-YY)")
	reportEntriesCmd.Flags().String("until", "", "end date (DD-MM-YY)")
	reportEntriesCmd.Flags().String("filter-tags", "", "filter by tags (comma-separated)")
	reportEntriesCmd.Flags().String("filter-categories", "", "filter by categories (comma-separated)")
	reportEntriesCmd.Flags().String("account", "", "account name")
	reportEntriesCmd.Flags().String("format", "table", "output format (table or md)")

	reportBalancesCmd.Flags().String("from", "", "start date (DD-MM-YY)")
	reportBalancesCmd.Flags().String("until", "", "end date (DD-MM-YY)")
	reportBalancesCmd.Flags().String("account", "", "account name")
	reportBalancesCmd.Flags().String("format", "table", "output format (table or md)")
}

func getDBManager() *database.Manager {
	if dbManager == nil {
		dbManager = database.NewManager(cfgLoader)
	}
	return dbManager
}

func getDB() (*sqlite.DB, error) {
	mgr := getDBManager()
	return mgr.GetDatabase(dbFlag)
}

func getFactory() (*persistence.RepositoryFactory, error) {
	if repoFactory != nil {
		return repoFactory, nil
	}

	cfg, err := cfgLoader.Load()
	if err != nil {
		return nil, err
	}

	if cfg.StorageDriver == config.DriverSQLite {
		db, err := getDB()
		if err != nil {
			return nil, err
		}
		repoFactory = persistence.NewRepositoryFactory(cfg, db.DB)
	} else {
		repoFactory = persistence.NewRepositoryFactory(cfg, nil)
	}

	return repoFactory, nil
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

	formats := []string{"2006-01-02", "06-01-02"}
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			if t.Year() < 100 {
				t = t.AddDate(2000, 0, 0)
			}
			return t.UTC()
		}
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

func init() {
	_ = filepath.Base
	_ = persistence.NewMigrationManager
}
