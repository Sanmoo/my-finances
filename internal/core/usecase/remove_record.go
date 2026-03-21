package usecase

import (
	"fmt"

	"github.com/Sanmoo/my-finances/internal/core/port"
)

type RemoveRecordInput struct {
	RecordID int64
}

type RemoveRecord struct {
	entryRepo port.EntriesRepository
}

func NewRemoveRecord(entryRepo port.EntriesRepository) *RemoveRecord {
	return &RemoveRecord{entryRepo: entryRepo}
}

func (uc *RemoveRecord) Execute(input RemoveRecordInput) error {
	entry, err := uc.entryRepo.GetByID(input.RecordID)
	if err != nil {
		return fmt.Errorf("failed to get entry: %w", err)
	}

	if entry == nil {
		return fmt.Errorf("entry with id %d not found", input.RecordID)
	}

	if err := uc.entryRepo.Delete(input.RecordID); err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	return nil
}
