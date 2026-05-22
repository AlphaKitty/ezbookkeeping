package models

import "encoding/json"

// InventoryStatus represents the status of an inventory record
type InventoryStatus string

const (
	INVENTORY_STATUS_IN_STOCK  InventoryStatus = "in_stock"
	INVENTORY_STATUS_RESERVED  InventoryStatus = "reserved"
	INVENTORY_STATUS_SOLD_OUT  InventoryStatus = "sold_out"
)

// InventoryAction represents the direction of inventory change in a transaction
type InventoryAction string

const (
	INVENTORY_ACTION_NONE     InventoryAction = "none"
	INVENTORY_ACTION_STOCK_IN  InventoryAction = "stock_in"
	INVENTORY_ACTION_STOCK_OUT InventoryAction = "stock_out"
)

// ItemFieldValues stores the actual values for item fields
type ItemFieldValues struct {
	Values map[string]any `json:"values"`
}

// FromDB deserializes data from database
func (v *ItemFieldValues) FromDB(data []byte) error {
	if len(data) == 0 {
		v.Values = make(map[string]any)
		return nil
	}
	return json.Unmarshal(data, v)
}

// ToDB serializes data to database
func (v *ItemFieldValues) ToDB() ([]byte, error) {
	if v.Values == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(v)
}

// InventoryRecord represents an inventory item stored in a warehouse
type InventoryRecord struct {
	InventoryRecordId int64            `xorm:"PK"`
	Uid               int64            `xorm:"INDEX(IDX_inventory_uid_deleted_warehouse) NOT NULL"`
	Deleted           bool             `xorm:"INDEX(IDX_inventory_uid_deleted_warehouse) NOT NULL"`
	ItemDefinitionId  int64            `xorm:"INDEX NOT NULL"`
	WarehouseId       int64            `xorm:"INDEX(IDX_inventory_uid_deleted_warehouse) NOT NULL DEFAULT 0"`
	FieldValues       *ItemFieldValues `xorm:"BLOB"`
	Quantity          float64          `xorm:"NOT NULL DEFAULT 0"`
	Unit              string           `xorm:"VARCHAR(32) NOT NULL DEFAULT ''"`
	UnitPrice         float64          `xorm:"NOT NULL DEFAULT 0"`
	Transporter       string           `xorm:"VARCHAR(64) NOT NULL DEFAULT ''"`
	BatchNo           string           `xorm:"VARCHAR(64) NOT NULL DEFAULT ''"`
	Status            InventoryStatus  `xorm:"VARCHAR(16) NOT NULL DEFAULT 'in_stock'"`
	Comment           string           `xorm:"VARCHAR(255) NOT NULL DEFAULT ''"`
	CreatedUnixTime   int64
	UpdatedUnixTime   int64
	DeletedUnixTime   int64
}

// InventoryRecordCreateRequest represents all parameters of inventory record creation request
type InventoryRecordCreateRequest struct {
	ItemDefinitionId int64            `json:"itemDefinitionId,string" binding:"required,min=1"`
	WarehouseId      int64            `json:"warehouseId,string"`
	FieldValues      *ItemFieldValues `json:"fieldValues"`
	Quantity         float64          `json:"quantity" binding:"min=0"`
	Unit             string           `json:"unit" binding:"max=32"`
	UnitPrice        float64          `json:"unitPrice" binding:"min=0"`
	Transporter      string           `json:"transporter" binding:"max=64"`
	BatchNo          string           `json:"batchNo" binding:"max=64"`
	Comment          string           `json:"comment" binding:"max=255"`
}

// InventoryRecordModifyRequest represents all parameters of inventory record modification request
type InventoryRecordModifyRequest struct {
	Id               int64            `json:"id,string" binding:"required,min=1"`
	ItemDefinitionId int64            `json:"itemDefinitionId,string"`
	WarehouseId      int64            `json:"warehouseId,string"`
	FieldValues      *ItemFieldValues `json:"fieldValues"`
	Quantity         float64          `json:"quantity" binding:"min=0"`
	Unit             string           `json:"unit" binding:"max=32"`
	UnitPrice        float64          `json:"unitPrice" binding:"min=0"`
	Transporter      string           `json:"transporter" binding:"max=64"`
	BatchNo          string           `json:"batchNo" binding:"max=64"`
	Status           InventoryStatus  `json:"status"`
	Comment          string           `json:"comment" binding:"max=255"`
}

// InventoryRecordGetRequest represents all parameters of inventory record getting request
type InventoryRecordGetRequest struct {
	Id int64 `form:"id,string" binding:"required,min=1"`
}

// InventoryRecordListRequest represents all parameters of inventory record listing request
type InventoryRecordListRequest struct {
	ItemDefinitionId int64           `form:"itemDefinitionId,string"`
	WarehouseId      int64           `form:"warehouseId,string"`
	Status           InventoryStatus `form:"status"`
}

// InventoryRecordDeleteRequest represents all parameters of inventory record deleting request
type InventoryRecordDeleteRequest struct {
	Id int64 `json:"id,string" binding:"required,min=1"`
}

// InventoryRecordInfoResponse represents a view-object of inventory record
type InventoryRecordInfoResponse struct {
	Id                 int64            `json:"id,string"`
	ItemDefinitionId   int64            `json:"itemDefinitionId,string"`
	ItemDefinitionName string           `json:"itemDefinitionName"`
	WarehouseId        int64            `json:"warehouseId,string"`
	FieldValues        *ItemFieldValues `json:"fieldValues"`
	Quantity           float64          `json:"quantity"`
	Unit               string           `json:"unit"`
	UnitPrice          float64          `json:"unitPrice"`
	Transporter        string           `json:"transporter"`
	BatchNo            string           `json:"batchNo"`
	Status             InventoryStatus  `json:"status"`
	Comment            string           `json:"comment"`
	CreatedUnixTime    int64            `json:"createdUnixTime"`
	UpdatedUnixTime    int64            `json:"updatedUnixTime"`
}

// ToInventoryRecordInfoResponse returns a view-object according to database model
func (r *InventoryRecord) ToInventoryRecordInfoResponse() *InventoryRecordInfoResponse {
	return &InventoryRecordInfoResponse{
		Id:               r.InventoryRecordId,
		ItemDefinitionId: r.ItemDefinitionId,
		WarehouseId:      r.WarehouseId,
		FieldValues:      r.FieldValues,
		Quantity:         r.Quantity,
		Unit:             r.Unit,
		UnitPrice:        r.UnitPrice,
		Transporter:      r.Transporter,
		BatchNo:          r.BatchNo,
		Status:           r.Status,
		Comment:          r.Comment,
		CreatedUnixTime:  r.CreatedUnixTime,
		UpdatedUnixTime:  r.UpdatedUnixTime,
	}
}

// InventoryRecordInfoResponseSlice represents the slice data structure of InventoryRecordInfoResponse
type InventoryRecordInfoResponseSlice []*InventoryRecordInfoResponse

// Len returns the count of items
func (s InventoryRecordInfoResponseSlice) Len() int {
	return len(s)
}

// Swap swaps two items
func (s InventoryRecordInfoResponseSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less reports whether the first item is less than the second one
func (s InventoryRecordInfoResponseSlice) Less(i, j int) bool {
	return s[i].UpdatedUnixTime > s[j].UpdatedUnixTime
}
