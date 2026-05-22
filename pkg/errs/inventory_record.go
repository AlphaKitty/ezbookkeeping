package errs

import "net/http"

// Error codes related to inventory records
var (
	ErrInventoryRecordIdInvalid = NewNormalError(NormalSubcategoryInventory, 0, http.StatusBadRequest, "inventory record id is invalid")
	ErrInventoryRecordNotFound  = NewNormalError(NormalSubcategoryInventory, 1, http.StatusBadRequest, "inventory record not found")
)
