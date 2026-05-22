package errs

import "net/http"

// Error codes related to item definitions
var (
	ErrItemDefinitionIdInvalid = NewNormalError(NormalSubcategoryItemDefinition, 0, http.StatusBadRequest, "item definition id is invalid")
	ErrItemDefinitionNotFound  = NewNormalError(NormalSubcategoryItemDefinition, 1, http.StatusBadRequest, "item definition not found")
)
