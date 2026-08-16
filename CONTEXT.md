# My-Finances

A CLI for personal finance tracking. Users record income and expense entries in accounts, optionally linking them to credit cards; reports expose balances and card invoices.

## Language

**Entry**:
A single income or expense record, always attached to an account and stored under the account's data path.
_Avoid_: transaction, lançamento

**Realization date**:
The date the purchase actually happened — the `--date` flag passed on the CLI.
_Avoid_: purchase date

**Payment date**:
For credit card entries, the date the charge lands on the invoice (the day it is paid to the bank). Derived from the realization date plus the card's closing and due days, never entered by the user. Entries without a card have no payment date.
_Avoid_: due date, charge date

**Closing day**:
Day of the month on which the card's invoice closes. A purchase realized on or after the closing day belongs to the next invoice.
_Avoid_: cutoff day

**Due day**:
Day of the month on which the card's invoice is due (vencimento).
_Avoid_: payment day

**Invoice (Fatura)**:
The set of credit card entries whose payment date falls in a given month. An invoice is identified by its due month: "June invoice" means the invoice due in June. An invoice covers purchases from the previous month's closing day up to the day before the current month's closing day (its purchase window).
_Avoid_: bill, statement, fatura

**Statement (Extrato)**:
The report that lists the entries of a single invoice, optionally restricted to one account.
_Avoid_: extract, fatura

**Purchase window**:
The realization-date interval covered by an invoice: from the closing day of the previous month through the day before the closing day of the invoice's month (e.g. closing day 9 → window 09/05–08/06 for the June invoice).
_Avoid_: period, billing cycle

**Installment (Parcela)**:
Each of the N entries created by `--times N`. The CLI amount is the per-installment value, so the total purchase value is amount × N; each installment lands on a consecutive invoice.
_Avoid_: recurrence, monthly charge

**Account**:
A named scope that owns entries and categories (e.g. a bank account or a person's context). Credit cards and tags are global, not scoped to an account.

**Category**:
A per-account classifier for entries, with a display name, an alias used in commands, and a type (income or expense).

**Tag**:
A global label applied to entries; tags must be registered before use.
