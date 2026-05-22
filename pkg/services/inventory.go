package services

import (
	"time"

	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

// InventoryRecordService represents inventory record service
type InventoryRecordService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a inventory record service singleton instance
var (
	InventoryRecords = &InventoryRecordService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetInventoryRecordsByUid returns inventory records of user with optional filters
func (s *InventoryRecordService) GetInventoryRecordsByUid(c core.Context, uid int64, request *models.InventoryRecordListRequest) ([]*models.InventoryRecord, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	sess := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false)

	if request.ItemDefinitionId > 0 {
		sess = sess.Where("item_definition_id=?", request.ItemDefinitionId)
	}
	if request.WarehouseId > 0 {
		sess = sess.Where("warehouse_id=?", request.WarehouseId)
	}
	if request.Status != "" {
		sess = sess.Where("status=?", request.Status)
	}

	var records []*models.InventoryRecord
	err := sess.OrderBy("updated_unix_time desc").Find(&records)

	return records, err
}

// GetInventoryRecordById returns an inventory record model according to inventory record id
func (s *InventoryRecordService) GetInventoryRecordById(c core.Context, uid int64, id int64) (*models.InventoryRecord, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if id <= 0 {
		return nil, errs.ErrInventoryRecordIdInvalid
	}

	record := &models.InventoryRecord{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(id).Where("uid=? AND deleted=?", uid, false).Get(record)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrInventoryRecordNotFound
	}

	return record, nil
}

// CreateInventoryRecord creates a new inventory record
func (s *InventoryRecordService) CreateInventoryRecord(c core.Context, uid int64, request *models.InventoryRecordCreateRequest) (*models.InventoryRecord, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()
	status := models.INVENTORY_STATUS_IN_STOCK

	record := &models.InventoryRecord{
		InventoryRecordId: s.GenerateUuid(uuid.UUID_TYPE_INVENTORY),
		Uid:               uid,
		ItemDefinitionId:  request.ItemDefinitionId,
		WarehouseId:       request.WarehouseId,
		FieldValues:       request.FieldValues,
		Quantity:          request.Quantity,
		Unit:              request.Unit,
		UnitPrice:         request.UnitPrice,
		Transporter:       request.Transporter,
		BatchNo:           request.BatchNo,
		Status:            status,
		Comment:           request.Comment,
		CreatedUnixTime:   now,
		UpdatedUnixTime:   now,
	}

	_, err := s.UserDataDB(uid).NewSession(c).Insert(record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

// ModifyInventoryRecord modifies an existing inventory record
func (s *InventoryRecordService) ModifyInventoryRecord(c core.Context, uid int64, request *models.InventoryRecordModifyRequest) (*models.InventoryRecord, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	record, err := s.GetInventoryRecordById(c, uid, request.Id)
	if err != nil {
		return nil, err
	}

	record.ItemDefinitionId = request.ItemDefinitionId
	record.WarehouseId = request.WarehouseId
	record.FieldValues = request.FieldValues
	record.Quantity = request.Quantity
	record.Unit = request.Unit
	record.UnitPrice = request.UnitPrice
	record.Transporter = request.Transporter
	record.BatchNo = request.BatchNo
	record.Status = request.Status
	record.Comment = request.Comment
	record.UpdatedUnixTime = time.Now().Unix()

	_, err = s.UserDataDB(uid).NewSession(c).ID(record.InventoryRecordId).Cols(
		"item_definition_id", "warehouse_id", "field_values", "quantity", "unit",
		"unit_price", "transporter", "batch_no", "status", "comment", "updated_unix_time",
	).Update(record)

	return record, err
}

// DeleteInventoryRecord deletes an inventory record (soft delete)
func (s *InventoryRecordService) DeleteInventoryRecord(c core.Context, uid int64, id int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if id <= 0 {
		return errs.ErrInventoryRecordIdInvalid
	}

	now := time.Now().Unix()
	_, err := s.UserDataDB(uid).NewSession(c).ID(id).Where("uid=? AND deleted=?", uid, false).Cols("deleted", "deleted_unix_time").Update(&models.InventoryRecord{
		Deleted:         true,
		DeletedUnixTime: now,
	})

	return err
}

// UpdateInventoryQuantity updates the quantity of an inventory record within a transaction session
func (s *InventoryRecordService) UpdateInventoryQuantity(c core.Context, uid int64, id int64, delta float64, newStatus models.InventoryStatus, sess *xorm.Session) error {
	record, err := s.GetInventoryRecordById(c, uid, id)
	if err != nil {
		return err
	}

	newQuantity := record.Quantity + delta
	if newQuantity < 0 {
		newQuantity = 0
	}

	update := &models.InventoryRecord{
		Quantity:        newQuantity,
		Status:          newStatus,
		UpdatedUnixTime: time.Now().Unix(),
	}

	_, err = sess.ID(id).Cols("quantity", "status", "updated_unix_time").Update(update)
	return err
}
