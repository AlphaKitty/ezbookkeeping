# Tracked inventory fields in the transaction calendar

The transaction calendar currently shows daily income/expense totals. We added
the ability to track sums of specific ItemField values (e.g. total weight of steel
plates consumed each day) by marking fields with `trackInCalendar: true` on the
ItemDefinition.

**Key decisions:**

- **Configuration lives on `ItemField`** as a `TrackInCalendar` flag rather than
  a separate user-level setting. The field is an intrinsic property of the item type
  — if "weight" matters for steel plates, it always matters regardless of which
  user is viewing.

- **Aggregation is `field_value × consumed_quantity`**, not raw field value.
  A single inventory record has a fixed field value (e.g. unit weight = 5 kg).
  The transaction consumes a quantity (e.g. 3 units). The meaningful daily sum
  is `5 × 3 = 15 kg`, not `5 kg`.

- **Aggregated via `TransactionInventoryIndex`**, the many-to-many join table.
  A single transaction can consume multiple inventory records, possibly from
  different ItemDefinitions. Each join row's `Amount` is the multiplier.

- **Computed server-side** in the monthly transaction list response
  (`GET /v1/transactions/list/by_month.json`). The response includes a
  `trackedFieldDailySums` block keyed by day-of-month, then by ItemDefinition,
  then by field key → `{value, unit}`.

- **Grouped by ItemDefinition.** Mixing tracked fields from different item types
  (e.g. weight of steel + length of pipe) has no business meaning.

## Considered alternatives

- **Global user config for tracked fields**: would require a separate settings UI
  and storage. Rejected — tracking is a property of the item type, not the user.
- **Client-side computation**: would require N+1 fetches for inventory record
  field values. Rejected — server-side is a single query.
- **Direct field value sum (without ×Amount multiplier)**: would give the same
  result whether 1 or 100 units were consumed. Rejected — quantity matters.
