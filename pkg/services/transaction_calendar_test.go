package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func TestComputeTrackedFieldDailySums(t *testing.T) {
	// Helper: make a simple field schema with tracked fields
	makeSchema := func(trackedKeys ...string) *models.ItemFieldSchema {
		fields := make([]*models.ItemField, 0)
		for _, k := range trackedKeys {
			fields = append(fields, &models.ItemField{
				Key:             k,
				FieldType:       models.ITEM_FIELD_TYPE_NUMBER,
				TrackInCalendar: true,
				Unit:            "kg",
			})
		}
		return &models.ItemFieldSchema{Fields: fields}
	}

	tests := []struct {
		name        string
		transactions []*models.Transaction
		indexes     map[int64][]*models.TransactionInventoryIndex // txId → indexes
		records     map[int64]*models.InventoryRecord
		definitions map[int64]*models.ItemDefinition
		want        map[int]*models.TrackedFieldDailySum
	}{
		{
			name:         "empty transactions → empty sums",
			transactions: []*models.Transaction{},
			indexes:      map[int64][]*models.TransactionInventoryIndex{},
			records:      map[int64]*models.InventoryRecord{},
			definitions:  map[int64]*models.ItemDefinition{},
			want:         map[int]*models.TrackedFieldDailySum{},
		},
		{
			name: "single transaction with tracked field → correct sum",
			transactions: []*models.Transaction{
				{
					TransactionId:   1,
					TransactionTime: makeTxTime(15, 0), // day 15
				},
			},
			indexes: map[int64][]*models.TransactionInventoryIndex{
				1: {
					{TransactionId: 1, InventoryRecordId: 10, Amount: 3},
				},
			},
			records: map[int64]*models.InventoryRecord{
				10: {
					InventoryRecordId: 10,
					ItemDefinitionId:  100,
					FieldValues:       &models.ItemFieldValues{Values: map[string]any{"weight": 5.0}},
				},
			},
			definitions: map[int64]*models.ItemDefinition{
				100: {
					ItemDefinitionId: 100,
					Name:             "钢板",
					FieldSchema:      makeSchema("weight"),
				},
			},
			want: map[int]*models.TrackedFieldDailySum{
				15: {
					ItemDefs: []models.TrackedFieldItemDefSum{
						{
							ItemDefinitionId:   100,
							ItemDefinitionName: "钢板",
							Fields: []models.TrackedFieldValue{
								{Key: "weight", Value: 15.0, Unit: "kg"}, // 5 * 3
							},
						},
					},
				},
			},
		},
		{
			name: "two transactions same day → sums merged",
			transactions: []*models.Transaction{
				{TransactionId: 1, TransactionTime: makeTxTime(10, 8)},
				{TransactionId: 2, TransactionTime: makeTxTime(10, 14)},
			},
			indexes: map[int64][]*models.TransactionInventoryIndex{
				1: {{TransactionId: 1, InventoryRecordId: 10, Amount: 2}},
				2: {{TransactionId: 2, InventoryRecordId: 10, Amount: 3}},
			},
			records: map[int64]*models.InventoryRecord{
				10: {
					InventoryRecordId: 10,
					ItemDefinitionId:  100,
					FieldValues:       &models.ItemFieldValues{Values: map[string]any{"weight": 5.0}},
				},
			},
			definitions: map[int64]*models.ItemDefinition{
				100: {ItemDefinitionId: 100, Name: "钢板", FieldSchema: makeSchema("weight")},
			},
			want: map[int]*models.TrackedFieldDailySum{
				10: {
					ItemDefs: []models.TrackedFieldItemDefSum{
						{ItemDefinitionId: 100, ItemDefinitionName: "钢板",
							Fields: []models.TrackedFieldValue{
								{Key: "weight", Value: 25.0, Unit: "kg"}, // 5*(2+3)
							},
						},
					},
				},
			},
		},
		{
			name: "different ItemDefinitions → separate groups",
			transactions: []*models.Transaction{
				{TransactionId: 1, TransactionTime: makeTxTime(5, 0)},
			},
			indexes: map[int64][]*models.TransactionInventoryIndex{
				1: {
					{TransactionId: 1, InventoryRecordId: 10, Amount: 2},
					{TransactionId: 1, InventoryRecordId: 20, Amount: 3},
				},
			},
			records: map[int64]*models.InventoryRecord{
				10: {InventoryRecordId: 10, ItemDefinitionId: 100, FieldValues: &models.ItemFieldValues{Values: map[string]any{"weight": 5.0}}},
				20: {InventoryRecordId: 20, ItemDefinitionId: 200, FieldValues: &models.ItemFieldValues{Values: map[string]any{"length": 10.0}}},
			},
			definitions: map[int64]*models.ItemDefinition{
				100: {ItemDefinitionId: 100, Name: "钢板", FieldSchema: makeSchema("weight")},
				200: {ItemDefinitionId: 200, Name: "钢管",
					FieldSchema: &models.ItemFieldSchema{Fields: []*models.ItemField{
						{Key: "length", FieldType: models.ITEM_FIELD_TYPE_NUMBER, TrackInCalendar: true, Unit: "m"},
					}},
				},
			},
			want: map[int]*models.TrackedFieldDailySum{
				5: {
					ItemDefs: []models.TrackedFieldItemDefSum{
						{ItemDefinitionId: 100, ItemDefinitionName: "钢板", Fields: []models.TrackedFieldValue{{Key: "weight", Value: 10.0, Unit: "kg"}}},
						{ItemDefinitionId: 200, ItemDefinitionName: "钢管", Fields: []models.TrackedFieldValue{{Key: "length", Value: 30.0, Unit: "m"}}},
					},
				},
			},
		},
		{
			name: "no tracked fields → empty sums",
			transactions: []*models.Transaction{
				{TransactionId: 1, TransactionTime: makeTxTime(1, 0)},
			},
			indexes: map[int64][]*models.TransactionInventoryIndex{
				1: {{TransactionId: 1, InventoryRecordId: 10, Amount: 1}},
			},
			records: map[int64]*models.InventoryRecord{
				10: {InventoryRecordId: 10, ItemDefinitionId: 100, FieldValues: &models.ItemFieldValues{Values: map[string]any{"weight": 5.0}}},
			},
			definitions: map[int64]*models.ItemDefinition{
				100: {ItemDefinitionId: 100, Name: "钢板", FieldSchema: &models.ItemFieldSchema{Fields: []*models.ItemField{
					{Key: "weight", FieldType: models.ITEM_FIELD_TYPE_NUMBER, TrackInCalendar: false},
				}}},
			},
			want: map[int]*models.TrackedFieldDailySum{},
		},
		{
			name: "non-number field type → skipped",
			transactions: []*models.Transaction{
				{TransactionId: 1, TransactionTime: makeTxTime(1, 0)},
			},
			indexes: map[int64][]*models.TransactionInventoryIndex{
				1: {{TransactionId: 1, InventoryRecordId: 10, Amount: 1}},
			},
			records: map[int64]*models.InventoryRecord{
				10: {InventoryRecordId: 10, ItemDefinitionId: 100, FieldValues: &models.ItemFieldValues{Values: map[string]any{"note": "hello"}}},
			},
			definitions: map[int64]*models.ItemDefinition{
				100: {ItemDefinitionId: 100, Name: "钢板",
					FieldSchema: &models.ItemFieldSchema{Fields: []*models.ItemField{
						{Key: "note", FieldType: models.ITEM_FIELD_TYPE_TEXT, TrackInCalendar: true},
					}},
				},
			},
			want: map[int]*models.TrackedFieldDailySum{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeTrackedFieldDailySums(tt.transactions, tt.indexes, tt.records, tt.definitions)
			assert.Equal(t, tt.want, got)
		})
	}
}

// makeTxTime creates a TransactionTime for the given day of June 2026.
// TransactionTime = UnixTime * 1000.
func makeTxTime(day, hour int) int64 {
	daysSinceJune1 := int64(day - 1)
	hoursSeconds := int64(hour) * 3600
	// June 1, 2026 00:00:00 UTC = 20605 days since epoch = 1780272000
	june1Unix := int64(1780272000)
	unixTime := june1Unix + daysSinceJune1*86400 + hoursSeconds
	return unixTime * 1000
}
