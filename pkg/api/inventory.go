package api

import (
	"sort"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
)

// InventoryRecordsApi represents inventory record api
type InventoryRecordsApi struct {
	records *services.InventoryRecordService
	items   *services.ItemDefinitionService
}

// Initialize an inventory record api singleton instance
var (
	InventoryRecords = &InventoryRecordsApi{
		records: services.InventoryRecords,
		items:   services.ItemDefinitions,
	}
)

// InventoryRecordListHandler returns inventory record list of current user
func (a *InventoryRecordsApi) InventoryRecordListHandler(c *core.WebContext) (any, *errs.Error) {
	var listReq models.InventoryRecordListRequest
	err := c.ShouldBindQuery(&listReq)

	if err != nil {
		log.Warnf(c, "[inventory_records.InventoryRecordListHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	records, err := a.records.GetInventoryRecordsByUid(c, uid, &listReq)

	if err != nil {
		log.Errorf(c, "[inventory_records.InventoryRecordListHandler] failed to get inventory records for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	resps := make(models.InventoryRecordInfoResponseSlice, len(records))
	for i := 0; i < len(records); i++ {
		resps[i] = records[i].ToInventoryRecordInfoResponse()
		a.fillItemDefinitionName(c, uid, records[i].ItemDefinitionId, resps[i])
	}

	sort.Sort(resps)

	return resps, nil
}

// InventoryRecordGetHandler returns one specific inventory record of current user
func (a *InventoryRecordsApi) InventoryRecordGetHandler(c *core.WebContext) (any, *errs.Error) {
	var getReq models.InventoryRecordGetRequest
	err := c.ShouldBindQuery(&getReq)

	if err != nil {
		log.Warnf(c, "[inventory_records.InventoryRecordGetHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	record, err := a.records.GetInventoryRecordById(c, uid, getReq.Id)

	if err != nil {
		log.Errorf(c, "[inventory_records.InventoryRecordGetHandler] failed to get inventory record \"id:%d\" for user \"uid:%d\", because %s", getReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	resp := record.ToInventoryRecordInfoResponse()
	a.fillItemDefinitionName(c, uid, record.ItemDefinitionId, resp)

	return resp, nil
}

// InventoryRecordCreateHandler saves a new inventory record by request parameters for current user
func (a *InventoryRecordsApi) InventoryRecordCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var createReq models.InventoryRecordCreateRequest
	err := c.ShouldBindJSON(&createReq)

	if err != nil {
		log.Warnf(c, "[inventory_records.InventoryRecordCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	record, err := a.records.CreateInventoryRecord(c, uid, &createReq)

	if err != nil {
		log.Errorf(c, "[inventory_records.InventoryRecordCreateHandler] failed to create inventory record for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[inventory_records.InventoryRecordCreateHandler] user \"uid:%d\" has created inventory record \"id:%d\"", uid, record.InventoryRecordId)

	resp := record.ToInventoryRecordInfoResponse()
	a.fillItemDefinitionName(c, uid, record.ItemDefinitionId, resp)

	return resp, nil
}

// InventoryRecordModifyHandler saves an existed inventory record by request parameters for current user
func (a *InventoryRecordsApi) InventoryRecordModifyHandler(c *core.WebContext) (any, *errs.Error) {
	var modifyReq models.InventoryRecordModifyRequest
	err := c.ShouldBindJSON(&modifyReq)

	if err != nil {
		log.Warnf(c, "[inventory_records.InventoryRecordModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	record, err := a.records.ModifyInventoryRecord(c, uid, &modifyReq)

	if err != nil {
		log.Errorf(c, "[inventory_records.InventoryRecordModifyHandler] failed to update inventory record \"id:%d\" for user \"uid:%d\", because %s", modifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[inventory_records.InventoryRecordModifyHandler] user \"uid:%d\" has updated inventory record \"id:%d\"", uid, record.InventoryRecordId)

	resp := record.ToInventoryRecordInfoResponse()
	a.fillItemDefinitionName(c, uid, record.ItemDefinitionId, resp)

	return resp, nil
}

// InventoryRecordDeleteHandler deletes an existed inventory record by request parameters for current user
func (a *InventoryRecordsApi) InventoryRecordDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var deleteReq models.InventoryRecordDeleteRequest
	err := c.ShouldBindJSON(&deleteReq)

	if err != nil {
		log.Warnf(c, "[inventory_records.InventoryRecordDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.records.DeleteInventoryRecord(c, uid, deleteReq.Id)

	if err != nil {
		log.Errorf(c, "[inventory_records.InventoryRecordDeleteHandler] failed to delete inventory record \"id:%d\" for user \"uid:%d\", because %s", deleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[inventory_records.InventoryRecordDeleteHandler] user \"uid:%d\" has deleted inventory record \"id:%d\"", uid, deleteReq.Id)
	return true, nil
}

func (a *InventoryRecordsApi) fillItemDefinitionName(c core.Context, uid int64, itemDefinitionId int64, resp *models.InventoryRecordInfoResponse) {
	if itemDefinitionId <= 0 {
		return
	}

	definition, err := a.items.GetItemDefinitionById(c, uid, itemDefinitionId)
	if err != nil {
		return
	}

	resp.ItemDefinitionName = definition.Name
}
