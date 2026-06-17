# ezbookkeeping

A self-hosted double-entry bookkeeping application for personal and small business use.

## Language

### Core Entities

**User**:
A person who owns accounts, transactions, and inventory in a single tenant.
_Avoid_: Account holder, member, employee

**Account**:
A financial account that holds a balance in a specific currency. Can be an asset (cash, bank, investment, savings, receivables, virtual wallet, certificate of deposit) or a liability (credit card, debt). Accounts form a two-level hierarchy via parent accounts.
_Avoid_: Bank account, wallet, ledger

**Transaction**:
A record of money movement. Four types: income (money in), expense (money out), transfer (money moved between accounts), and balance modification (direct balance adjustment).
_Avoid_: Entry, record, journal entry

**TransactionCategory**:
A classification label for transactions (income, expense, or transfer type). Two-level hierarchy.
_Avoid_: Category, tag, label

**TransactionTag**:
An additional freeform label attached to a transaction. Tags belong to tag groups. Up to 10 tags per transaction.
_Avoid_: Label, flag

**TransactionTemplate**:
A reusable transaction preset. Can be a one-time quick-entry template or a scheduled recurring template (daily/weekly/monthly/yearly).
_Avoid_: Recurring transaction, scheduled entry

### Inventory

**ItemDefinition**:
A product or service type definition with a dynamic field schema. Defines custom fields (number, text, enum, date), pricing expressions for income and expense transactions, and default categories.
_Avoid_: Item type, product type, SKU definition

**InventoryRecord**:
A concrete stock record of an ItemDefinition. Stores field values per the item's schema, plus quantity, unit, unit price, batch number, transporter, and status.
_Avoid_: Stock entry, inventory item, warehouse record

**ItemField**:
A single field within an ItemDefinition's schema. Has a key, label, type (number/text/enum/date), and optional attributes: required, editable, unit, format, default value, sort order, and a computed expression (expr).

**Computed Field**:
An ItemField whose value is derived from other fields via an arithmetic expression (`expr`) rather than manual input. Computed fields can reference other computed fields (chaining). If a referenced field's value is empty, the computed field evaluates to empty (empty-on-incomplete). Expressions use `+ - * / ( )` with field keys as variables.
_Avoid_: Calculated field, derived field, formula field

**Pricing Expression**:
An arithmetic expression (`ExpensePricingExpr` / `IncomePricingExpr`) on an ItemDefinition that calculates the transaction amount when an inventory record is consumed in an expense or income transaction. Same expression syntax as computed fields.
_Avoid_: Price formula, amount expression
