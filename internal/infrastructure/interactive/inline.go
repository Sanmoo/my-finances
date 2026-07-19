package interactive

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Sanmoo/my-finances/internal/core/usecase"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

// inlineAddAccount prompts for account name and creates it.
// Returns the account name to use.
func (w *Wizard) inlineAddAccount() (string, error) {
	fmt.Println("\n--- Nova conta ---")
	for {
		name, err := w.prompter.Text("Nome da conta", "")
		if err != nil {
			return "", err
		}
		if name == "" {
			continue
		}

		uc := usecase.NewAddAccount(w.accountRepo)
		result, err := uc.Execute(usecase.AddAccountInput{Name: name})
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				fmt.Println("Conta já existe.")
				return name, nil
			}
			return "", fmt.Errorf("erro ao criar conta: %w", err)
		}
		w.printer.PrintAccount(result.Account)
		return name, nil
	}
}

// inlineAddCategory prompts for name, alias, and emoji, then creates the category.
// Returns the alias to use.
func (w *Wizard) inlineAddCategory(accountName string, catType entity.CategoryType) (string, error) {
	fmt.Println("\n--- Nova categoria ---")
	for {
		name, err := w.prompter.Text("Nome", "")
		if err != nil {
			return "", err
		}
		if name == "" {
			continue
		}

		for {
			alias, err := w.prompter.Text("Alias", "")
			if err != nil {
				return "", err
			}
			if alias == "" {
				continue
			}

			emoji, err := w.prompter.Text("Emoji (opcional)", "")
			if err != nil {
				return "", err
			}

			uc := usecase.NewAddCategory(w.categoryRepo)
			result, err := uc.Execute(usecase.AddCategoryInput{
				AccountName: accountName,
				Name:        name,
				Type:        catType,
				Alias:       alias,
				Emoji:       emoji,
			})
			if err != nil {
				return "", fmt.Errorf("erro ao criar categoria: %w", err)
			}
			w.printer.PrintCategory(result.Category)
			return alias, nil
		}
	}
}

// inlineAddTag prompts for tag name and creates it.
// Returns the tag name.
func (w *Wizard) inlineAddTag() (string, error) {
	fmt.Println("\n--- Nova tag ---")
	for {
		name, err := w.prompter.Text("Nome", "")
		if err != nil {
			return "", err
		}
		if name == "" {
			continue
		}

		uc := usecase.NewAddTag(w.tagRepo)
		_, err = uc.Execute(usecase.AddTagInput{Name: name})
		if err != nil {
			if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "already registered") {
				fmt.Println("Tag já existe, usando.")
				return name, nil
			}
			return "", fmt.Errorf("erro ao criar tag: %w", err)
		}
		fmt.Printf("Tag criada: %s\n", name)
		return name, nil
	}
}

// inlineAddCreditCard prompts for name, closing day, due day and creates it.
// Returns the credit card name.
func (w *Wizard) inlineAddCreditCard() (string, error) {
	fmt.Println("\n--- Novo cartão de crédito ---")
	for {
		name, err := w.prompter.Text("Nome", "")
		if err != nil {
			return "", err
		}
		if name == "" {
			continue
		}

		for {
			closingStr, err := w.prompter.Text("Dia de fechamento", "")
			if err != nil {
				return "", err
			}
			closing, cerr := strconv.Atoi(closingStr)
			if cerr != nil || closing < 1 || closing > 31 {
				fmt.Println("Dia inválido. Informe um número entre 1 e 31.")
				continue
			}

			dueStr, err := w.prompter.Text("Dia de vencimento", "")
			if err != nil {
				return "", err
			}
			due, derr := strconv.Atoi(dueStr)
			if derr != nil || due < 1 || due > 31 {
				fmt.Println("Dia inválido. Informe um número entre 1 e 31.")
				continue
			}

			uc := usecase.NewAddCreditCard(w.ccRepo)
			result, err := uc.Execute(usecase.AddCreditCardInput{
				Name:       name,
				ClosingDay: closing,
				DueDay:     due,
			})
			if err != nil {
				return "", fmt.Errorf("erro ao criar cartão: %w", err)
			}
			w.printer.PrintCreditCard(result.CreditCard)
			return name, nil
		}
	}
}
