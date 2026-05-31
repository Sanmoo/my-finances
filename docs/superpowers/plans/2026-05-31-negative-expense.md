# Negative Expense Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Allow expenses to be created with negative amounts while keeping them as expenses.

**Architecture:** The behavior belongs in the domain entry validation because all entry creation paths call `entity.NewEntry`. Update only amount validation so zero remains invalid and negative income remains invalid, while negative expense becomes valid.

**Tech Stack:** Go, Cobra CLI, YAML persistence, `stretchr/testify` tests.

---

### Task 1: Domain validation

**Files:**
- Modify: `internal/domain/entity/entry_test.go`
- Modify: `internal/domain/entity/entry.go`

- [ ] **Step 1: Write the failing test**

In `internal/domain/entity/entry_test.go`, add this table case to `TestNewEntry` after `valid expense entry`:

```go
{
    name:      "valid negative expense entry",
    entryType: EntryTypeExpense,
    amount:    -20.00,
    currency:  "BRL",
    date:      baseDate,
    opts:      []EntryOption{},
    wantErr:   nil,
},
```

Also rename the existing `negative amount` case to `negative income amount` to keep income validation explicit.

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/domain/entity -run TestNewEntry/valid_negative_expense_entry -count=1
```

Expected: FAIL with `amount must be greater than 0`.

- [ ] **Step 3: Write minimal implementation**

In `internal/domain/entity/entry.go`, replace:

```go
if amount <= 0 {
    return nil, ErrInvalidAmount
}
```

with:

```go
if amount == 0 || (entryType == EntryTypeIncome && amount < 0) {
    return nil, ErrInvalidAmount
}
```

- [ ] **Step 4: Run domain tests**

Run:

```bash
go test ./internal/domain/entity -count=1
```

Expected: PASS.

### Task 2: Usecase regression

**Files:**
- Modify: `internal/core/usecase/add_entry_test.go`

- [ ] **Step 1: Write the failing regression test**

In `TestAddEntry_Execute`, add a subtest that calls `AddEntry.Execute` with `Type: entity.EntryTypeExpense` and `Amount: "-20"`, then asserts `result.Entries[0].Type == entity.EntryTypeExpense` and `result.Entries[0].Amount == -20.00`.

- [ ] **Step 2: Run test**

Run:

```bash
go test ./internal/core/usecase -run TestAddEntry_Execute/success_with_negative_expense_amount -count=1
```

Expected after Task 1: PASS. If run before Task 1, expected FAIL with `amount must be greater than 0`.

### Task 3: Full verification

- [ ] **Step 1: Run all tests**

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 2: Commit implementation**

```bash
git add internal/domain/entity/entry.go internal/domain/entity/entry_test.go internal/core/usecase/add_entry_test.go docs/superpowers/plans/2026-05-31-negative-expense.md
git commit -m "feat: allow negative expense amounts"
```
