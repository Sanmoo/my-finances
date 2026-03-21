package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-finances/internal/core/usecase"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"github.com/Sanmoo/my-finances/internal/infrastructure/cli"
	"github.com/Sanmoo/my-finances/internal/infrastructure/config"
	"github.com/Sanmoo/my-finances/internal/infrastructure/persistence"
	"github.com/Sanmoo/my-finances/internal/infrastructure/persistence/sqlite"
)

var (
	dbPath        string
	namespaceFlag string
	printer       = cli.NewPrinter()
	cfgLoader     = config.NewLoader()
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
		db, err := initDB()
		if err != nil {
			printer.PrintError(err.Error())
			return
		}
		defer db.Close()

		nsID := getNamespaceID(db)
		repo := sqlite.NewAccountsRepository(db)
		uc := usecase.NewAddAccount(repo)

		result, err := uc.Execute(usecase.AddAccountInput{
			NamespaceID: nsID,
			Name:        args[0],
		})
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		printer.PrintAccount(result.Account)
	},
}

var addCategoryCmd = &cobra.Command{
	Use:   "category --type <inc|exp> --name <name> [--alias <alias>] [--emoji <emoji>]",
	Short: "Add a new category",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := initDB()
		if err != nil {
			printer.PrintError(err.Error())
			return
		}
		defer db.Close()

		catType, _ := cmd.Flags().GetString("type")
		name, _ := cmd.Flags().GetString("name")
		alias, _ := cmd.Flags().GetString("alias")
		emoji, _ := cmd.Flags().GetString("emoji")

		if name == "" {
			printer.PrintError("name is required")
			return
		}

		nsID := getNamespaceID(db)
		repo := sqlite.NewCategoriesRepository(db)
		uc := usecase.NewAddCategory(repo)

		result, err := uc.Execute(usecase.AddCategoryInput{
			NamespaceID: nsID,
			Name:        name,
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
		db, err := initDB()
		if err != nil {
			printer.PrintError(err.Error())
			return
		}
		defer db.Close()

		closingDay, _ := cmd.Flags().GetInt("closing-day")
		dueDay, _ := cmd.Flags().GetInt("due-day")

		if closingDay == 0 || dueDay == 0 {
			printer.PrintError("closing-day and due-day are required")
			return
		}

		nsID := getNamespaceID(db)
		repo := sqlite.NewCreditCardsRepository(db)
		uc := usecase.NewAddCreditCard(repo)

		result, err := uc.Execute(usecase.AddCreditCardInput{
			NamespaceID: nsID,
			Name:        args[0],
			ClosingDay:  closingDay,
			DueDay:      dueDay,
		})
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		printer.PrintCreditCard(result.CreditCard)
	},
}

var addExpenseCmd = &cobra.Command{
	Use:   "expense [amount] [--tags x,y] [--date DD-MM-YY] [--category x] [--description x] [--credit-card x] [--times n]",
	Short: "Add a new expense",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		db, err := initDB()
		if err != nil {
			printer.PrintError(err.Error())
			return
		}
		defer db.Close()

		amount := "0"
		if len(args) > 0 {
			amount = args[0]
		}

		tags, _ := cmd.Flags().GetStringSlice("tags")
		dateStr, _ := cmd.Flags().GetString("date")
		categoryStr, _ := cmd.Flags().GetString("category")
		description, _ := cmd.Flags().GetString("description")
		creditCardStr, _ := cmd.Flags().GetString("credit-card")
		times, _ := cmd.Flags().GetInt("times")

		date := parseDate(dateStr)
		currency := getDefaultCurrency()
		nsID := getNamespaceID(db)

		entryRepo := sqlite.NewEntriesRepository(db)
		categoryRepo := sqlite.NewCategoriesRepository(db)
		ccRepo := sqlite.NewCreditCardsRepository(db)
		uc := usecase.NewAddEntry(entryRepo, categoryRepo, ccRepo)

		var categoryID *int64
		if categoryStr != "" {
			cat, _ := categoryRepo.GetByNameOrAlias(nsID, categoryStr)
			if cat != nil {
				categoryID = &cat.ID
			}
		}

		var cc *entity.CreditCard
		if creditCardStr != "" {
			ccs, _ := ccRepo.GetByNamespaceID(nsID)
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
			NamespaceID: nsID,
			Type:        entity.EntryTypeExpense,
			Amount:      amount,
			Currency:    currency,
			Description: description,
			CategoryID:  categoryID,
			CreditCard:  cc,
			Tags:        tags,
			Times:       times,
			Date:        date,
		})
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		for _, entry := range result.Entries {
			printer.PrintEntry(entry)
		}
	},
}

var addIncomeCmd = &cobra.Command{
	Use:   "income [amount] [--date x] [--category x] [--description x]",
	Short: "Add a new income",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		db, err := initDB()
		if err != nil {
			printer.PrintError(err.Error())
			return
		}
		defer db.Close()

		amount := "0"
		if len(args) > 0 {
			amount = args[0]
		}

		dateStr, _ := cmd.Flags().GetString("date")
		categoryStr, _ := cmd.Flags().GetString("category")
		description, _ := cmd.Flags().GetString("description")

		date := parseDate(dateStr)
		currency := getDefaultCurrency()
		nsID := getNamespaceID(db)

		entryRepo := sqlite.NewEntriesRepository(db)
		categoryRepo := sqlite.NewCategoriesRepository(db)
		ccRepo := sqlite.NewCreditCardsRepository(db)
		uc := usecase.NewAddEntry(entryRepo, categoryRepo, ccRepo)

		var categoryID *int64
		if categoryStr != "" {
			cat, _ := categoryRepo.GetByNameOrAlias(nsID, categoryStr)
			if cat != nil {
				categoryID = &cat.ID
			}
		}

		result, err := uc.Execute(usecase.AddEntryInput{
			NamespaceID: nsID,
			Type:        entity.EntryTypeIncome,
			Amount:      amount,
			Currency:    currency,
			Description: description,
			CategoryID:  categoryID,
			Date:        date,
		})
		if err != nil {
			printer.PrintError(err.Error())
			return
		}

		for _, entry := range result.Entries {
			printer.PrintEntry(entry)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.AddCommand(addAccountCmd)
	addCmd.AddCommand(addCategoryCmd)
	addCmd.AddCommand(addCreditCardCmd)
	addCmd.AddCommand(addExpenseCmd)
	addCmd.AddCommand(addIncomeCmd)

	rootCmd.PersistentFlags().StringVarP(&namespaceFlag, "namespace", "s", "", "namespace to use")
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "", "database path")

	addCategoryCmd.Flags().StringP("type", "t", "", "category type (inc or exp)")
	addCategoryCmd.Flags().StringP("name", "n", "", "category name")
	addCategoryCmd.Flags().String("alias", "", "category alias")
	addCategoryCmd.Flags().String("emoji", "", "category emoji")

	addCreditCardCmd.Flags().Int("closing-day", 0, "closing day (1-31)")
	addCreditCardCmd.Flags().Int("due-day", 0, "due day (1-31)")

	addExpenseCmd.Flags().StringSlice("tags", []string{}, "tags (comma-separated)")
	addExpenseCmd.Flags().String("date", "", "date (DD-MM-YY)")
	addExpenseCmd.Flags().String("category", "", "category name or alias")
	addExpenseCmd.Flags().String("description", "", "description")
	addExpenseCmd.Flags().String("credit-card", "", "credit card name")
	addExpenseCmd.Flags().Int("times", 0, "number of installments")

	addIncomeCmd.Flags().String("date", "", "date (DD-MM-YY)")
	addIncomeCmd.Flags().String("category", "", "category name or alias")
	addIncomeCmd.Flags().String("description", "", "description")
}

func initDB() (*sqlite.DB, error) {
	path := dbPath
	if path == "" {
		homeDir, _ := os.UserHomeDir()
		path = filepath.Join(homeDir, "myfin.db")
	}

	db, err := sqlite.Open(path)
	if err != nil {
		return nil, err
	}

	absMigrationsPath, _ := filepath.Abs("migrations")
	mm := persistence.NewMigrationManager(db.DB, "file://"+absMigrationsPath)
	if err := mm.Up(); err != nil {
		return nil, err
	}

	return db, nil
}

func getNamespaceID(db *sqlite.DB) int64 {
	nsName := namespaceFlag
	if nsName == "" {
		cfg, err := cfgLoader.Load()
		if err == nil && cfg != nil {
			nsName = cfg.DefaultNamespace
		}
	}
	if nsName == "" {
		nsName = "default"
	}

	nsRepo := sqlite.NewNamespacesRepository(db)
	ns, err := nsRepo.GetByName(nsName)
	if err != nil || ns == nil {
		ns, _ = entity.NewNamespace(nsName)
		if ns != nil {
			id, _ := nsRepo.Create(ns)
			return id
		}
		return 1
	}

	return ns.ID
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
		return time.Now().UTC()
	}

	dateStr = strings.TrimSpace(dateStr)

	formats := []string{"02-01-06", "02-01-2006", "2006-01-02"}
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t.UTC()
		}
	}

	return time.Now().UTC()
}
