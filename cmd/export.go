package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-finances/internal/domain/entity"
	"github.com/Sanmoo/my-finances/internal/infrastructure/config"
	"github.com/Sanmoo/my-finances/internal/infrastructure/database"
	"github.com/Sanmoo/my-finances/internal/infrastructure/persistence"
	"github.com/Sanmoo/my-finances/internal/infrastructure/persistence/yaml"
)

func init() {
	rootCmd.AddCommand(exportCmd)
}

var exportCmd = &cobra.Command{
	Use:   "export --from <sqlite|yaml> --to <sqlite|yaml>",
	Short: "Export data between storage formats",
	Run: func(cmd *cobra.Command, args []string) {
		fromDriver, _ := cmd.Flags().GetString("from")
		toDriver, _ := cmd.Flags().GetString("to")
		accountStr, _ := cmd.Flags().GetString("account")

		if fromDriver == "" || toDriver == "" {
			fmt.Println("Error: --from and --to flags are required")
			os.Exit(1)
		}

		if fromDriver == toDriver {
			fmt.Printf("Error: source and destination are the same format: %s\n", fromDriver)
			os.Exit(1)
		}

		cfg, err := cfgLoader.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		if fromDriver == config.DriverSQLite {
			dbManager := database.NewManager(cfgLoader)
			db, err := dbManager.GetDatabase(dbFlag)
			if err != nil {
				fmt.Printf("Error opening SQLite database: %v\n", err)
				os.Exit(1)
			}
			defer db.Close()
		}

		if fromDriver == config.DriverYAML {
			yaml.Init(cfg.DataPath)
		}

		if toDriver == config.DriverYAML {
			yaml.Init(cfg.DataPath)
		}

		fmt.Printf("Exporting from %s to %s...\n", fromDriver, toDriver)

		var sqliteRepoFactory, yamlRepoFactory *persistence.RepositoryFactory

		if fromDriver == config.DriverSQLite {
			dbManager := database.NewManager(cfgLoader)
			db, err := dbManager.GetDatabase(dbFlag)
			if err != nil {
				fmt.Printf("Error opening SQLite database: %v\n", err)
				os.Exit(1)
			}
			defer db.Close()

			sqliteCfg := &config.Config{
				StorageDriver: config.DriverSQLite,
				DatabasesPath: cfg.DatabasesPath,
			}
			sqliteRepoFactory = persistence.NewRepositoryFactory(sqliteCfg, db.DB)
		}

		if toDriver == config.DriverYAML {
			yamlCfg := &config.Config{
				StorageDriver: config.DriverYAML,
				DataPath:      cfg.DataPath,
			}
			yamlRepoFactory = persistence.NewRepositoryFactory(yamlCfg, nil)
		}

		fmt.Printf("SQLite factory: %v\n", sqliteRepoFactory != nil)
		fmt.Printf("YAML factory: %v\n", yamlRepoFactory != nil)

		if sqliteRepoFactory != nil && yamlRepoFactory != nil {
			err = exportFromSQLiteToYAML(sqliteRepoFactory, yamlRepoFactory, accountStr)
			if err != nil {
				fmt.Printf("Error exporting: %v\n", err)
				os.Exit(1)
			}
		} else if yamlRepoFactory != nil && sqliteRepoFactory != nil {
			err = exportFromYAMLToSQLite(yamlRepoFactory, sqliteRepoFactory, accountStr)
			if err != nil {
				fmt.Printf("Error exporting: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("Error: unsupported export direction")
			os.Exit(1)
		}

		fmt.Println("Export completed successfully!")
	},
}

func exportFromSQLiteToYAML(sqliteFactory, yamlFactory *persistence.RepositoryFactory, accountStr string) error {
	sqliteAccounts := sqliteFactory.NewAccountsRepository()
	sqliteCategories := sqliteFactory.NewCategoriesRepository()
	sqliteCreditCards := sqliteFactory.NewCreditCardsRepository()
	sqliteEntries := sqliteFactory.NewEntriesRepository()

	yamlAccounts := yamlFactory.NewAccountsRepository()
	yamlCategories := yamlFactory.NewCategoriesRepository()
	yamlCreditCards := yamlFactory.NewCreditCardsRepository()
	yamlEntries := yamlFactory.NewEntriesRepository()

	fmt.Println("Exporting accounts...")
	accounts, err := sqliteAccounts.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get accounts: %w", err)
	}
	fmt.Printf("  Found %d accounts in SQLite\n", len(accounts))
	for _, acc := range accounts {
		fmt.Printf("  - Account: %s (ID: %d)\n", acc.Name, acc.ID)
		newAcc := &entity.Account{Name: acc.Name}
		_, err := yamlAccounts.Create(newAcc)
		if err != nil {
			return fmt.Errorf("failed to create account %s: %w", acc.Name, err)
		}
	}

	fmt.Println("Exporting categories...")
	categories, err := sqliteCategories.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get categories: %w", err)
	}
	for _, cat := range categories {
		newCat := &entity.Category{
			Name:  cat.Name,
			Type:  cat.Type,
			Alias: cat.Alias,
			Emoji: cat.Emoji,
		}
		_, err := yamlCategories.Create(newCat)
		if err != nil {
			return fmt.Errorf("failed to create category %s: %w", cat.Name, err)
		}
		fmt.Printf("  - Category: %s\n", cat.Name)
	}

	fmt.Println("Exporting credit cards...")
	creditCards, err := sqliteCreditCards.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get credit cards: %w", err)
	}
	for _, cc := range creditCards {
		newCC := &entity.CreditCard{
			Name:       cc.Name,
			ClosingDay: cc.ClosingDay,
			DueDay:     cc.DueDay,
		}
		_, err := yamlCreditCards.Create(newCC)
		if err != nil {
			return fmt.Errorf("failed to create credit card %s: %w", cc.Name, err)
		}
		fmt.Printf("  - Credit card: %s\n", cc.Name)
	}

	fmt.Println("Exporting entries...")
	entries, err := sqliteEntries.GetAll(nil)
	if err != nil {
		return fmt.Errorf("failed to get entries: %w", err)
	}
	for _, entry := range entries {
		_, err := yamlEntries.Create(entry)
		if err != nil {
			return fmt.Errorf("failed to create entry: %w", err)
		}
	}
	fmt.Printf("  - Exported %d entries\n", len(entries))

	return nil
}

func exportFromYAMLToSQLite(yamlFactory, sqliteFactory *persistence.RepositoryFactory, accountStr string) error {
	return fmt.Errorf("YAML to SQLite export not implemented yet")
}

func init() {
	exportCmd.Flags().String("from", "", "source format (sqlite or yaml)")
	exportCmd.Flags().String("to", "", "destination format (sqlite or yaml)")
	exportCmd.Flags().String("account", "", "account name (optional, exports all if not specified)")
}
