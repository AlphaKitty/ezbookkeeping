# Computed item fields via `expr` on ItemField

ItemDefinitions have dynamic field schemas (number, text, enum, date). Some fields are not
manually entered — they are computed from other fields (e.g. `total_weight = unit_weight * quantity`).
We extended `ItemField` with an optional `expr` string that holds an arithmetic expression.

**Key decisions**:

- **Expression lives on ItemField itself** rather than a separate `computedFields` list. An
  ItemField is either manually filled (`editable=true`) or computed (`expr` set, `editable=false`).
  This keeps the schema flat and the producer/consumer boundary clear.

- **Expression syntax = the existing `EvaluateExpression` engine**: `+ - * / ( )` with field keys
  as variable names. Same syntax already used by `ExpensePricingExpr` and `IncomePricingExpr`.

- **Chaining allowed**: a computed field can reference another computed field (e.g. `tax = subtotal * 0.13`
  where `subtotal` itself is computed). Topological sort at evaluation time; cycle detection at
  ItemDefinition save time.

- **Dual computation**: frontend evaluates expressions in real time for instant preview; backend
  re-evaluates authoritatively on save. Same expression grammar on both sides.

- **Empty-on-incomplete**: if any referenced field's value is missing, the computed field
  resolves to empty rather than substituting zero or a default. Save validation rejects incomplete
  computed fields.

## Considered alternatives

- **Separate `computedFields` list**: would split the schema into two categories and force
  consumers to merge two sources. Rejected — a single `fields` list with `editable`/`expr`
  is simpler and the distinction is already captured by those fields.
- **No chaining**: would require duplicating sub-expressions (e.g. `tax = unit_weight * quantity * 0.13`
  instead of `tax = subtotal * 0.13`). Rejected — semantic clarity matters more than avoiding
  topological sort.
- **Backend-only evaluation**: would mean no live preview. Rejected — inventory entry is
  interactive and users need instant feedback.
- **Zero-fill for missing values**: would silently produce wrong results. Rejected — empty
  propagation makes missing data visible.
