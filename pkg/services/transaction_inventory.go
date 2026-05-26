package services

import (
	"time"

	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

// TransactionInventoryService represents transaction inventory index service
type TransactionInventoryService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a transaction inventory service singleton instance
var (
	TransactionInventoryIndexes = &TransactionInventoryService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// CreateInventoryIndexes creates transaction-inventory join table rows
func (s *TransactionInventoryService) CreateInventoryIndexes(c core.Context, uid int64, transactionId int64, inventoryRecordIds []int64, amounts []float64, transactionTime int64, sess *xorm.Session) error {
	now := time.Now().Unix()
	n := len(inventoryRecordIds)
	if n > len(amounts) {
		n = len(amounts)
	}

	indexUuids := s.GenerateUuids(uuid.UUID_TYPE_INVENTORY_INDEX, uint16(n))
	if len(indexUuids) < n {
		return nil // not enough UUIDs, skip
	}

	indexes := make([]*models.TransactionInventoryIndex, 0, n)
	for i := 0; i < n; i++ {
		indexes = append(indexes, &models.TransactionInventoryIndex{
			IndexId:           indexUuids[i],
			Uid:               uid,
			Deleted:           false,
			TransactionTime:   transactionTime,
			TransactionId:     transactionId,
			InventoryRecordId: inventoryRecordIds[i],
			Amount:            amounts[i],
			CreatedUnixTime:   now,
			UpdatedUnixTime:   now,
		})
	}

	if len(indexes) > 0 {
		_, err := sess.Insert(indexes)
		return err
	}
	return nil
}

// DeleteInventoryIndexesByTransactionId soft-deletes all inventory indexes for a transaction
func (s *TransactionInventoryService) DeleteInventoryIndexesByTransactionId(c core.Context, uid int64, transactionId int64, sess *xorm.Session) error {
	now := time.Now().Unix()
	update := &models.TransactionInventoryIndex{
		Deleted:         true,
		DeletedUnixTime: now,
	}
	_, err := sess.Cols("deleted", "deleted_unix_time").
		Where("uid=? AND deleted=? AND transaction_id=?", uid, false, transactionId).
		Update(update)
	return err
}

// GetInventoryIndexesByTransactionId returns all inventory indexes for a transaction
func (s *TransactionInventoryService) GetInventoryIndexesByTransactionId(c core.Context, uid int64, transactionId int64) ([]*models.TransactionInventoryIndex, error) {
	var indexes []*models.TransactionInventoryIndex
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=? AND transaction_id=?", uid, false, transactionId).Find(&indexes)
	return indexes, err
}

// GetInventoryIndexesByTransactionIds returns inventory indexes for multiple transactions
func (s *TransactionInventoryService) GetInventoryIndexesByTransactionIds(c core.Context, uid int64, transactionIds []int64) ([]*models.TransactionInventoryIndex, error) {
	if len(transactionIds) == 0 {
		return nil, nil
	}
	var indexes []*models.TransactionInventoryIndex
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=? AND transaction_id IN ?", uid, false, transactionIds).Find(&indexes)
	return indexes, err
}
