package services

import (
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// ComputeTrackedFieldDailySums computes daily sums of tracked fields from
// transactions and their linked inventory records. Fields must be marked
// trackInCalendar=true and be of number type. Sum = field_value × consumed_amount.
func ComputeTrackedFieldDailySums(
	transactions []*models.Transaction,
	indexes map[int64][]*models.TransactionInventoryIndex,
	records map[int64]*models.InventoryRecord,
	definitions map[int64]*models.ItemDefinition,
) map[int]*models.TrackedFieldDailySum {
	type sumKey struct {
		day        int
		itemDefId  int64
		fieldKey   string
	}
	accum := make(map[sumKey]struct {
		total float64
		unit  string
		name  string
	})

	for _, tx := range transactions {
		unixSec := tx.TransactionTime / 1000
		t := time.Unix(unixSec, 0).UTC()
		day := t.Day()

		txIndexes := indexes[tx.TransactionId]
		for _, idx := range txIndexes {
			record, ok := records[idx.InventoryRecordId]
			if !ok {
				continue
			}
			def, ok := definitions[record.ItemDefinitionId]
			if !ok || def.FieldSchema == nil {
				continue
			}

			for _, field := range def.FieldSchema.Fields {
				if !field.TrackInCalendar || field.FieldType != models.ITEM_FIELD_TYPE_NUMBER {
					continue
				}
				rawVal, exists := record.FieldValues.Values[field.Key]
				if !exists {
					continue
				}
				fv, ok := toFloat64(rawVal)
				if !ok {
					continue
				}

				sk := sumKey{day: day, itemDefId: def.ItemDefinitionId, fieldKey: field.Key}
				entry := accum[sk]
				entry.total += fv * idx.Amount
				entry.unit = field.Unit
				entry.name = def.Name
				accum[sk] = entry
			}
		}
	}

	// Build result: group by day → itemDefId → fields
	result := make(map[int]*models.TrackedFieldDailySum)
	for sk, entry := range accum {
		if _, ok := result[sk.day]; !ok {
			result[sk.day] = &models.TrackedFieldDailySum{}
		}
		daily := result[sk.day]

		// Find or create the ItemDef group for this day
		var itemDefGroup *models.TrackedFieldItemDefSum
		for i := range daily.ItemDefs {
			if daily.ItemDefs[i].ItemDefinitionId == sk.itemDefId {
				itemDefGroup = &daily.ItemDefs[i]
				break
			}
		}
		if itemDefGroup == nil {
			daily.ItemDefs = append(daily.ItemDefs, models.TrackedFieldItemDefSum{
				ItemDefinitionId:   sk.itemDefId,
				ItemDefinitionName: entry.name,
			})
			itemDefGroup = &daily.ItemDefs[len(daily.ItemDefs)-1]
		}

		itemDefGroup.Fields = append(itemDefGroup.Fields, models.TrackedFieldValue{
			Key:   sk.fieldKey,
			Value: entry.total,
			Unit:  entry.unit,
		})
	}

	return result
}
