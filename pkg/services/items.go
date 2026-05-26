package services

import (
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

// ItemDefinitionService represents item definition service
type ItemDefinitionService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a item definition service singleton instance
var (
	ItemDefinitions = &ItemDefinitionService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetAllItemDefinitionsByUid returns all item definition models of user
func (s *ItemDefinitionService) GetAllItemDefinitionsByUid(c core.Context, uid int64) ([]*models.ItemDefinition, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var definitions []*models.ItemDefinition
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).Find(&definitions)

	return definitions, err
}

// GetItemDefinitionById returns an item definition model according to item definition id
func (s *ItemDefinitionService) GetItemDefinitionById(c core.Context, uid int64, id int64) (*models.ItemDefinition, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if id <= 0 {
		return nil, errs.ErrItemDefinitionIdInvalid
	}

	definition := &models.ItemDefinition{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(id).Where("uid=? AND deleted=?", uid, false).Get(definition)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrItemDefinitionNotFound
	}

	return definition, nil
}

// CreateItemDefinition creates a new item definition
func (s *ItemDefinitionService) CreateItemDefinition(c core.Context, uid int64, request *models.ItemDefinitionCreateRequest) (*models.ItemDefinition, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()
	definition := &models.ItemDefinition{
		ItemDefinitionId:  s.GenerateUuid(uuid.UUID_TYPE_ITEM_DEF),
		Uid:               uid,
		Name:              request.Name,
		Icon:              request.Icon,
		FieldSchema:       request.FieldSchema,
		ExpensePricingExpr: request.ExpensePricingExpr,
		IncomePricingExpr:  request.IncomePricingExpr,
		IncomeCategoryId:  request.IncomeCategoryId,
		ExpenseCategoryId: request.ExpenseCategoryId,
		CreatedUnixTime:   now,
		UpdatedUnixTime:   now,
	}

	_, err := s.UserDataDB(uid).NewSession(c).Insert(definition)
	if err != nil {
		return nil, err
	}

	return definition, nil
}

// ModifyItemDefinition modifies an existing item definition
func (s *ItemDefinitionService) ModifyItemDefinition(c core.Context, uid int64, request *models.ItemDefinitionModifyRequest) (*models.ItemDefinition, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	definition, err := s.GetItemDefinitionById(c, uid, request.Id)
	if err != nil {
		return nil, err
	}

	definition.Name = request.Name
	definition.Icon = request.Icon
	definition.FieldSchema = request.FieldSchema
	definition.ExpensePricingExpr = request.ExpensePricingExpr
	definition.IncomePricingExpr = request.IncomePricingExpr
	definition.IncomeCategoryId = request.IncomeCategoryId
	definition.ExpenseCategoryId = request.ExpenseCategoryId
	definition.UpdatedUnixTime = time.Now().Unix()

	_, err = s.UserDataDB(uid).NewSession(c).ID(definition.ItemDefinitionId).Cols("name", "icon", "field_schema", "expense_pricing_expr", "income_pricing_expr", "income_category_id", "expense_category_id", "updated_unix_time").Update(definition)

	return definition, err
}

// DeleteItemDefinition deletes an item definition (soft delete)
func (s *ItemDefinitionService) DeleteItemDefinition(c core.Context, uid int64, id int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if id <= 0 {
		return errs.ErrItemDefinitionIdInvalid
	}

	now := time.Now().Unix()
	_, err := s.UserDataDB(uid).NewSession(c).ID(id).Where("uid=? AND deleted=?", uid, false).Cols("deleted", "deleted_unix_time").Update(&models.ItemDefinition{
		Deleted:         true,
		DeletedUnixTime: now,
	})

	return err
}
