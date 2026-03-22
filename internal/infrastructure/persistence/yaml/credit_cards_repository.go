package yaml

import (
	"os"
	"path/filepath"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type CreditCardsRepository struct {
	basePath string
}

func NewCreditCardsRepository(basePath string) *CreditCardsRepository {
	return &CreditCardsRepository{basePath: basePath}
}

func (r *CreditCardsRepository) filePath() string {
	return filepath.Join(r.basePath, "credit_cards.yaml")
}

func (r *CreditCardsRepository) Create(cc *entity.CreditCard) error {
	data, err := Read[CreditCardsData](r.filePath())
	if err != nil {
		return err
	}

	yamlCC := CreditCard{}
	yamlCC.FromEntity(cc)
	data.CreditCards = append(data.CreditCards, yamlCC)

	return Write(r.filePath(), data)
}

func (r *CreditCardsRepository) GetByName(name string) (*entity.CreditCard, error) {
	data, err := Read[CreditCardsData](r.filePath())
	if err != nil {
		return nil, err
	}

	for i := range data.CreditCards {
		if data.CreditCards[i].Name == name {
			return data.CreditCards[i].ToEntity(), nil
		}
	}

	return nil, nil
}

func (r *CreditCardsRepository) GetAll() ([]*entity.CreditCard, error) {
	data, err := Read[CreditCardsData](r.filePath())
	if err != nil {
		return nil, err
	}

	cards := make([]*entity.CreditCard, len(data.CreditCards))
	for i := range data.CreditCards {
		cards[i] = data.CreditCards[i].ToEntity()
	}

	return cards, nil
}

func (r *CreditCardsRepository) EnsureInitialized() error {
	path := r.filePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return Write(path, CreditCardsData{CreditCards: []CreditCard{}})
	}
	return nil
}

func (r *CreditCardsRepository) Port() port.CreditCardsRepository {
	return r
}
