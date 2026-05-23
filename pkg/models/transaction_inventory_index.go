package models

// TransactionInventoryIndex represents transaction and inventory record relation stored in database
type TransactionInventoryIndex struct {
	IndexId           int64 `xorm:"PK"`
	Uid               int64 `xorm:"INDEX(IDX_transaction_inventory_index_uid_deleted_transaction_id) INDEX(IDX_transaction_inventory_index_uid_deleted_inventory_record_id) NOT NULL"`
	Deleted           bool  `xorm:"INDEX(IDX_transaction_inventory_index_uid_deleted_transaction_id) INDEX(IDX_transaction_inventory_index_uid_deleted_inventory_record_id) NOT NULL"`
	TransactionTime   int64 `xorm:"NOT NULL"`
	TransactionId     int64 `xorm:"INDEX(IDX_transaction_inventory_index_uid_deleted_transaction_id)"`
	InventoryRecordId int64 `xorm:"INDEX(IDX_transaction_inventory_index_uid_deleted_inventory_record_id)"`
	Amount            float64
	CreatedUnixTime   int64
	UpdatedUnixTime   int64
	DeletedUnixTime   int64
}
