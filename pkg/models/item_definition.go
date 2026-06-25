package models

import "encoding/json"

// ItemFieldType represents the type of a field in an item definition
type ItemFieldType string

const (
	ITEM_FIELD_TYPE_NUMBER ItemFieldType = "number"
	ITEM_FIELD_TYPE_TEXT   ItemFieldType = "text"
	ITEM_FIELD_TYPE_ENUM   ItemFieldType = "enum"
	ITEM_FIELD_TYPE_DATE   ItemFieldType = "date"
)

// ItemField represents a single field definition in an item schema
type ItemField struct {
	Key                  string        `json:"key"`
	Label                string        `json:"label"`
	FieldType            ItemFieldType `json:"fieldType"`
	Required             bool          `json:"required"`
	Editable             bool          `json:"editable"`
	ParticipateInNaming  bool          `json:"participateInNaming"`
	Options              []string      `json:"options,omitempty"`
	Unit                 string        `json:"unit,omitempty"`
	Format               string        `json:"format,omitempty"`
	DefaultValue         string        `json:"defaultValue,omitempty"`
	Expr                 string        `json:"expr,omitempty"`
	TrackInCalendar      bool          `json:"trackInCalendar"`
	SortOrder            int           `json:"sortOrder"`
}

// ItemFieldSchema wraps a list of fields for BLOB storage
type ItemFieldSchema struct {
	Fields []*ItemField `json:"fields"`
}

// FromDB deserializes data from database
func (s *ItemFieldSchema) FromDB(data []byte) error {
	return json.Unmarshal(data, s)
}

// ToDB serializes data to database
func (s *ItemFieldSchema) ToDB() ([]byte, error) {
	return json.Marshal(s)
}

// ItemDefinition represents an item type definition stored in database
type ItemDefinition struct {
	ItemDefinitionId int64            `xorm:"PK"`
	Uid              int64            `xorm:"INDEX(IDX_item_def_uid_deleted) NOT NULL"`
	Deleted          bool             `xorm:"INDEX(IDX_item_def_uid_deleted) NOT NULL"`
	Name             string           `xorm:"VARCHAR(64) NOT NULL"`
	Icon             string           `xorm:"VARCHAR(64) NOT NULL"`
	FieldSchema      *ItemFieldSchema `xorm:"BLOB"`
	ExpensePricingExpr string           `xorm:"VARCHAR(255) NOT NULL DEFAULT ''"`
	IncomePricingExpr  string           `xorm:"VARCHAR(255) NOT NULL DEFAULT ''"`
	IncomeCategoryId int64            `xorm:"NOT NULL DEFAULT 0"`
	ExpenseCategoryId int64            `xorm:"NOT NULL DEFAULT 0"`
	CreatedUnixTime  int64
	UpdatedUnixTime  int64
	DeletedUnixTime  int64
}

// ItemDefinitionCreateRequest represents all parameters of item definition creation request
type ItemDefinitionCreateRequest struct {
	Name             string           `json:"name" binding:"required,notBlank,max=64"`
	Icon             string           `json:"icon" binding:"max=64"`
	FieldSchema      *ItemFieldSchema `json:"fieldSchema" binding:"required"`
	ExpensePricingExpr string           `json:"expensePricingExpr" binding:"max=255"`
	IncomePricingExpr  string           `json:"incomePricingExpr" binding:"max=255"`
	IncomeCategoryId   int64            `json:"incomeCategoryId,string" binding:"required,min=1"`
	ExpenseCategoryId  int64            `json:"expenseCategoryId,string" binding:"required,min=1"`
}

// ItemDefinitionModifyRequest represents all parameters of item definition modification request
type ItemDefinitionModifyRequest struct {
	Id                 int64            `json:"id,string" binding:"required,min=1"`
	Name               string           `json:"name" binding:"required,notBlank,max=64"`
	Icon               string           `json:"icon" binding:"max=64"`
	FieldSchema        *ItemFieldSchema `json:"fieldSchema" binding:"required"`
	ExpensePricingExpr string           `json:"expensePricingExpr" binding:"max=255"`
	IncomePricingExpr  string           `json:"incomePricingExpr" binding:"max=255"`
	IncomeCategoryId   int64            `json:"incomeCategoryId,string" binding:"required,min=1"`
	ExpenseCategoryId  int64            `json:"expenseCategoryId,string" binding:"required,min=1"`
}

// ItemDefinitionGetRequest represents all parameters of item definition getting request
type ItemDefinitionGetRequest struct {
	Id int64 `form:"id,string" binding:"required,min=1"`
}

// ItemDefinitionDeleteRequest represents all parameters of item definition deleting request
type ItemDefinitionDeleteRequest struct {
	Id int64 `json:"id,string" binding:"required,min=1"`
}

// ItemDefinitionInfoResponse represents a view-object of item definition
type ItemDefinitionInfoResponse struct {
	Id               int64            `json:"id,string"`
	Name             string           `json:"name"`
	Icon             string           `json:"icon"`
	FieldSchema      *ItemFieldSchema `json:"fieldSchema"`
	ExpensePricingExpr string           `json:"expensePricingExpr"`
	IncomePricingExpr  string           `json:"incomePricingExpr"`
	IncomeCategoryId int64            `json:"incomeCategoryId,string"`
	ExpenseCategoryId int64            `json:"expenseCategoryId,string"`
}

// ToItemDefinitionInfoResponse returns a view-object according to database model
func (d *ItemDefinition) ToItemDefinitionInfoResponse() *ItemDefinitionInfoResponse {
	return &ItemDefinitionInfoResponse{
		Id:               d.ItemDefinitionId,
		Name:             d.Name,
		Icon:             d.Icon,
		FieldSchema:      d.FieldSchema,
		ExpensePricingExpr: d.ExpensePricingExpr,
			IncomePricingExpr:  d.IncomePricingExpr,
		IncomeCategoryId: d.IncomeCategoryId,
		ExpenseCategoryId: d.ExpenseCategoryId,
	}
}

// ItemDefinitionInfoResponseSlice represents the slice data structure of ItemDefinitionInfoResponse
type ItemDefinitionInfoResponseSlice []*ItemDefinitionInfoResponse

// Len returns the count of items
func (s ItemDefinitionInfoResponseSlice) Len() int {
	return len(s)
}

// Swap swaps two items
func (s ItemDefinitionInfoResponseSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less reports whether the first item is less than the second one
func (s ItemDefinitionInfoResponseSlice) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}
