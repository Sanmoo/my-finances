package yaml

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Sanmoo/my-finances/internal/core/port"
	"github.com/Sanmoo/my-finances/internal/domain/entity"
)

type CreditCardsRepository struct {
	basePath string
}

type CreditCardsData struct {
	CreditCards []CreditCard `yaml:"credit_cards"`
}

func NewCreditCardsRepository(basePath string) *CreditCardsRepository {
	return &CreditCardsRepository{basePath: basePath}
}

func (r *CreditCardsRepository) filePath() string {
	return filepath.Join(r.basePath, "credit_cards.yaml")
}

func (r *CreditCardsRepository) metaPath() string {
	return filepath.Join(r.basePath, "_meta.yaml")
}

func (r *CreditCardsRepository) Create(cc *entity.CreditCard) (int64, error) {
	if err := EnsureMetaFile(r.metaPath()); err != nil {
		return 0, err
	}

	nextID, err := GetNextID(r.metaPath(), "credit_cards")
	if err != nil {
		return 0, err
	}

	cc.ID = nextID

	data, err := Read[CreditCardsData](r.filePath())
	if err != nil {
		return 0, err
	}

	yamlCC := CreditCard{}
	yamlCC.FromEntity(cc)
	data.CreditCards = append(data.CreditCards, yamlCC)

	if err := Write(r.filePath(), data); err != nil {
		return 0, err
	}

	return cc.ID, nil
}

func (r *CreditCardsRepository) GetByID(id int64) (*entity.CreditCard, error) {
	data, err := Read[CreditCardsData](r.filePath())
	if err != nil {
		return nil, err
	}

	for i := range data.CreditCards {
		if data.CreditCards[i].ID == id {
			return data.CreditCards[i].ToEntity(), nil
		}
	}

	return nil, nil
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

func (r *CreditCardsRepository) Update(cc *entity.CreditCard) error {
	data, err := Read[CreditCardsData](r.filePath())
	if err != nil {
		return err
	}

	found := false
	for i := range data.CreditCards {
		if data.CreditCards[i].ID == cc.ID {
			data.CreditCards[i].FromEntity(cc)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("credit card not found: %d", cc.ID)
	}

	return Write(r.filePath(), data)
}

func (r *CreditCardsRepository) Delete(id int64) error {
	data, err := Read[CreditCardsData](r.filePath())
	if err != nil {
		return err
	}

	newCards := make([]CreditCard, 0)
	for _, cc := range data.CreditCards {
		if cc.ID != id {
			newCards = append(newCards, cc)
		}
	}

	data.CreditCards = newCards
	return Write(r.filePath(), data)
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
