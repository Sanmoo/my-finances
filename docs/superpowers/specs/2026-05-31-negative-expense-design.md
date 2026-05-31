# Negative Expense Design

## Goal

Allow CLI commands that create expenses to accept negative amounts, such as:

```bash
mae -d 18 -a sam -c rest -D "Restituição livup para o Edu" "-20"
```

The entry remains an expense with a negative amount. It must not be converted to income, because the intent is to undo or adjust part of a previous expense.

## Approach

Use the existing expense creation flow and relax only the validation that rejects negative amounts for expenses. Keep all other business rules unchanged, including tag registration, account/category rules, date parsing, credit card validations, YAML persistence, and amount expression parsing.

## Testing

Add a regression test that creates an expense with a negative amount through the existing CLI/usecase path and verifies that it is accepted and persisted or passed through as an expense with a negative amount.

## Out of Scope

- Adding a new entry type for refunds or adjustments.
- Converting negative expenses into income.
- Changing report semantics beyond the existing arithmetic effect of a negative expense.
