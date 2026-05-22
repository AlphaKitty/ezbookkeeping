package api

import (
	"sort"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
)

// ItemDefinitionsApi represents item definition api
type ItemDefinitionsApi struct {
	items *services.ItemDefinitionService
}

// Initialize an item definition api singleton instance
var (
	ItemDefinitions = &ItemDefinitionsApi{
		items: services.ItemDefinitions,
	}
)

// ItemDefinitionListHandler returns item definition list of current user
func (a *ItemDefinitionsApi) ItemDefinitionListHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()
	definitions, err := a.items.GetAllItemDefinitionsByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[item_definitions.ItemDefinitionListHandler] failed to get item definitions for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	resps := make(models.ItemDefinitionInfoResponseSlice, len(definitions))
	for i := 0; i < len(definitions); i++ {
		resps[i] = definitions[i].ToItemDefinitionInfoResponse()
	}

	sort.Sort(resps)

	return resps, nil
}

// ItemDefinitionGetHandler returns one specific item definition of current user
func (a *ItemDefinitionsApi) ItemDefinitionGetHandler(c *core.WebContext) (any, *errs.Error) {
	var getReq models.ItemDefinitionGetRequest
	err := c.ShouldBindQuery(&getReq)

	if err != nil {
		log.Warnf(c, "[item_definitions.ItemDefinitionGetHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	definition, err := a.items.GetItemDefinitionById(c, uid, getReq.Id)

	if err != nil {
		log.Errorf(c, "[item_definitions.ItemDefinitionGetHandler] failed to get item definition \"id:%d\" for user \"uid:%d\", because %s", getReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return definition.ToItemDefinitionInfoResponse(), nil
}

// ItemDefinitionCreateHandler saves a new item definition by request parameters for current user
func (a *ItemDefinitionsApi) ItemDefinitionCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var createReq models.ItemDefinitionCreateRequest
	err := c.ShouldBindJSON(&createReq)

	if err != nil {
		log.Warnf(c, "[item_definitions.ItemDefinitionCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	definition, err := a.items.CreateItemDefinition(c, uid, &createReq)

	if err != nil {
		log.Errorf(c, "[item_definitions.ItemDefinitionCreateHandler] failed to create item definition for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[item_definitions.ItemDefinitionCreateHandler] user \"uid:%d\" has created item definition \"id:%d\"", uid, definition.ItemDefinitionId)

	return definition.ToItemDefinitionInfoResponse(), nil
}

// ItemDefinitionModifyHandler saves an existed item definition by request parameters for current user
func (a *ItemDefinitionsApi) ItemDefinitionModifyHandler(c *core.WebContext) (any, *errs.Error) {
	var modifyReq models.ItemDefinitionModifyRequest
	err := c.ShouldBindJSON(&modifyReq)

	if err != nil {
		log.Warnf(c, "[item_definitions.ItemDefinitionModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	definition, err := a.items.ModifyItemDefinition(c, uid, &modifyReq)

	if err != nil {
		log.Errorf(c, "[item_definitions.ItemDefinitionModifyHandler] failed to update item definition \"id:%d\" for user \"uid:%d\", because %s", modifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[item_definitions.ItemDefinitionModifyHandler] user \"uid:%d\" has updated item definition \"id:%d\"", uid, definition.ItemDefinitionId)

	return definition.ToItemDefinitionInfoResponse(), nil
}

// ItemDefinitionDeleteHandler deletes an existed item definition by request parameters for current user
func (a *ItemDefinitionsApi) ItemDefinitionDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var deleteReq models.ItemDefinitionDeleteRequest
	err := c.ShouldBindJSON(&deleteReq)

	if err != nil {
		log.Warnf(c, "[item_definitions.ItemDefinitionDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.items.DeleteItemDefinition(c, uid, deleteReq.Id)

	if err != nil {
		log.Errorf(c, "[item_definitions.ItemDefinitionDeleteHandler] failed to delete item definition \"id:%d\" for user \"uid:%d\", because %s", deleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[item_definitions.ItemDefinitionDeleteHandler] user \"uid:%d\" has deleted item definition \"id:%d\"", uid, deleteReq.Id)
	return true, nil
}
