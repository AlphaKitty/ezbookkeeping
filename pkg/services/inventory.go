package services

import (
	"time"

	"xorm.io/xorm"

	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
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

	definition, err := ItemDefinitions.GetItemDefinitionById(c, uid, request.ItemDefinitionId)
	if err != nil {
		return nil, err
	}

	var fieldValues map[string]any
	if request.FieldValues != nil {
		fieldValues = request.FieldValues.Values
	}
	computedValues, err := computeRecordFields(definition.FieldSchema, fieldValues)
	if err != nil {
		return nil, err
	}
	request.FieldValues = &models.ItemFieldValues{Values: computedValues}

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

	_, err = s.UserDataDB(uid).NewSession(c).Insert(record)
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

	definition, err := ItemDefinitions.GetItemDefinitionById(c, uid, request.ItemDefinitionId)
	if err != nil {
		return nil, err
	}

	var modifyFieldValues map[string]any
	if request.FieldValues != nil {
		modifyFieldValues = request.FieldValues.Values
	}
	computedValues, err := computeRecordFields(definition.FieldSchema, modifyFieldValues)
	if err != nil {
		return nil, err
	}
	request.FieldValues = &models.ItemFieldValues{Values: computedValues}

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

// computeRecordFields computes computed field values from the schema and user-entered values.
// It returns the merged values map (manual + computed). Computed field values from the client
// are always overridden by authoritative server-side computation. Returns an error if any
// computed field cannot be resolved due to missing dependencies.
func computeRecordFields(schema *models.ItemFieldSchema, userValues map[string]any) (map[string]any, error) {
	if schema == nil || len(schema.Fields) == 0 {
		return userValues, nil
	}

	// Build FieldExpr list for computed fields
	var computedFields []utils.FieldExpr
	for _, f := range schema.Fields {
		if f.Expr != "" {
			computedFields = append(computedFields, utils.FieldExpr{Key: f.Key, Expr: f.Expr})
		}
	}

	if len(computedFields) == 0 {
		return userValues, nil
	}

	// Convert user values from map[string]any to map[string]float64
	floatValues := make(map[string]float64, len(userValues))
	for k, v := range userValues {
		if fv, ok := toFloat64(v); ok {
			floatValues[k] = fv
		}
	}

	// Evaluate computed fields
	computed, err := utils.EvaluateFields(computedFields, floatValues)
	if err != nil {
		return nil, fmt.Errorf("failed to compute fields: %w", err)
	}

	// Reject if any computed field could not be resolved
	for _, f := range computedFields {
		if _, ok := computed[f.Key]; !ok {
			return nil, fmt.Errorf("cannot compute field %q: missing dependencies", f.Key)
		}
	}

	// Merge computed values into result (copy user values, then overwrite with computed)
	result := make(map[string]any, len(userValues)+len(computed))
	for k, v := range userValues {
		result[k] = v
	}
	for k, v := range computed {
		result[k] = v
	}

	return result, nil
}

// toFloat64 converts a value to float64 if possible.
func toFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(val, 64)
		return f, err == nil
	default:
		return 0, false
	}
}
