package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func TestComputeRecordFields(t *testing.T) {
	tests := []struct {
		name       string
		schema     *models.ItemFieldSchema
		userValues map[string]any
		want       map[string]any
		wantErr    bool
	}{
		{
			name: "no computed fields → values pass through unchanged",
			schema: &models.ItemFieldSchema{
				Fields: []*models.ItemField{
					{Key: "length", FieldType: models.ITEM_FIELD_TYPE_NUMBER},
					{Key: "width", FieldType: models.ITEM_FIELD_TYPE_NUMBER},
				},
			},
			userValues: map[string]any{"length": 10.0, "width": 5.0},
			want:       map[string]any{"length": 10.0, "width": 5.0},
		},
		{
			name: "computed field overrides client value",
			schema: &models.ItemFieldSchema{
				Fields: []*models.ItemField{
					{Key: "price", FieldType: models.ITEM_FIELD_TYPE_NUMBER},
					{Key: "qty", FieldType: models.ITEM_FIELD_TYPE_NUMBER},
					{Key: "total", FieldType: models.ITEM_FIELD_TYPE_NUMBER, Expr: "price * qty"},
				},
			},
			userValues: map[string]any{"price": 10.0, "qty": 5.0, "total": 999.0},
			want:       map[string]any{"price": 10.0, "qty": 5.0, "total": 50.0},
		},
		{
			name: "chain computed fields",
			schema: &models.ItemFieldSchema{
				Fields: []*models.ItemField{
					{Key: "a", FieldType: models.ITEM_FIELD_TYPE_NUMBER},
					{Key: "b", FieldType: models.ITEM_FIELD_TYPE_NUMBER, Expr: "a * 2"},
					{Key: "c", FieldType: models.ITEM_FIELD_TYPE_NUMBER, Expr: "b + 10"},
				},
			},
			userValues: map[string]any{"a": 5.0},
			want:       map[string]any{"a": 5.0, "b": 10.0, "c": 20.0},
		},
		{
			name: "missing dependency → error",
			schema: &models.ItemFieldSchema{
				Fields: []*models.ItemField{
					{Key: "price", FieldType: models.ITEM_FIELD_TYPE_NUMBER},
					{Key: "total", FieldType: models.ITEM_FIELD_TYPE_NUMBER, Expr: "price * qty"},
				},
			},
			userValues: map[string]any{"price": 10.0},
			wantErr:    true,
		},
		{
			name: "nil schema → no-op",
			schema: nil,
			userValues: map[string]any{"x": 1.0},
			want:  map[string]any{"x": 1.0},
		},
		{
			name: "string values not convertible → computed field skipped",
			schema: &models.ItemFieldSchema{
				Fields: []*models.ItemField{
					{Key: "price", FieldType: models.ITEM_FIELD_TYPE_TEXT},
					{Key: "total", FieldType: models.ITEM_FIELD_TYPE_TEXT, Expr: "price * 2"},
				},
			},
			userValues: map[string]any{"price": "abc"},
			wantErr:    true,
		},
		{
			name: "JSON number strings from frontend → should be handled",
			schema: &models.ItemFieldSchema{
				Fields: []*models.ItemField{
					{Key: "length", FieldType: models.ITEM_FIELD_TYPE_NUMBER},
					{Key: "width", FieldType: models.ITEM_FIELD_TYPE_NUMBER},
					{Key: "area", FieldType: models.ITEM_FIELD_TYPE_NUMBER, Expr: "length * width"},
				},
			},
			userValues: map[string]any{"length": 3.0, "width": 2.0},
			want:       map[string]any{"length": 3.0, "width": 2.0, "area": 6.0},
		},
		{
			name: "string numbers from frontend (v-model without .number) → should be parseable",
			schema: &models.ItemFieldSchema{
				Fields: []*models.ItemField{
					{Key: "length", FieldType: models.ITEM_FIELD_TYPE_NUMBER},
					{Key: "width", FieldType: models.ITEM_FIELD_TYPE_NUMBER},
					{Key: "area", FieldType: models.ITEM_FIELD_TYPE_NUMBER, Expr: "length * width"},
				},
			},
			userValues: map[string]any{"length": "3", "width": "2"},
			want:       map[string]any{"length": "3", "width": "2", "area": 6.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := computeRecordFields(tt.schema, tt.userValues)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
