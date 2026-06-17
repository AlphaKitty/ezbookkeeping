package utils

import (
	"reflect"
	"testing"
)

func TestEvaluateExpression(t *testing.T) {
	tests := []struct {
		expr    string
		vars    map[string]float64
		want    float64
		wantErr bool
	}{
		{"", nil, 0, false},
		{"42", nil, 42, false},
		{"weight * unit_price", map[string]float64{"weight": 100, "unit_price": 12.5}, 1250, false},
		{"weight * unit_price * quantity", map[string]float64{"weight": 50, "unit_price": 8, "quantity": 3}, 1200, false},
		{"price + tax", map[string]float64{"price": 100, "tax": 13}, 113, false},
		{"revenue - cost", map[string]float64{"revenue": 500, "cost": 300}, 200, false},
		{"total / count", map[string]float64{"total": 100, "count": 4}, 25, false},
		{"(a + b) * c", map[string]float64{"a": 2, "b": 3, "c": 4}, 20, false},
		{"a + b * c", map[string]float64{"a": 2, "b": 3, "c": 4}, 14, false},
		{"-price", map[string]float64{"price": 10}, -10, false},
		{"2.5 * 4", nil, 10, false},
		{"x / 0", map[string]float64{"x": 1}, 0, true},
		{"undefined_var", nil, 0, true},
	}

	for _, tt := range tests {
		got, err := EvaluateExpression(tt.expr, tt.vars)
		if tt.wantErr {
			if err == nil {
				t.Errorf("EvaluateExpression(%q) expected error, got nil", tt.expr)
			}
			continue
		}
		if err != nil {
			t.Errorf("EvaluateExpression(%q) unexpected error: %v", tt.expr, err)
			continue
		}
		if got != tt.want {
			t.Errorf("EvaluateExpression(%q) = %v, want %v", tt.expr, got, tt.want)
		}
	}
}

func TestEvaluateFields(t *testing.T) {
	tests := []struct {
		name      string
		fields    []FieldExpr
		values    map[string]float64
		want      map[string]float64
		wantErr   bool
	}{
		{
			name:    "no computed fields",
			fields:  []FieldExpr{},
			values:  map[string]float64{"a": 1},
			want:    map[string]float64{},
			wantErr: false,
		},
		{
			name: "single computed field with all dependencies satisfied",
			fields: []FieldExpr{
				{Key: "total", Expr: "price * qty"},
			},
			values:  map[string]float64{"price": 10, "qty": 5},
			want:    map[string]float64{"total": 50},
			wantErr: false,
		},
		{
			name: "chain dependency A→B→C",
			fields: []FieldExpr{
				{Key: "b", Expr: "a * 2"},
				{Key: "c", Expr: "b + 10"},
			},
			values:  map[string]float64{"a": 5},
			want:    map[string]float64{"b": 10, "c": 20},
			wantErr: false,
		},
		{
			name: "missing dependency → field skipped",
			fields: []FieldExpr{
				{Key: "total", Expr: "price * qty"},
			},
			values:  map[string]float64{"price": 10},
			want:    map[string]float64{},
			wantErr: false,
		},
		{
			name: "cycle detection",
			fields: []FieldExpr{
				{Key: "a", Expr: "b + 1"},
				{Key: "b", Expr: "a + 1"},
			},
			values:  map[string]float64{},
			wantErr: true,
		},
		{
			name: "self-referencing cycle",
			fields: []FieldExpr{
				{Key: "a", Expr: "a + 1"},
			},
			values:  map[string]float64{},
			wantErr: true,
		},
		{
			name: "mixed manual and computed",
			fields: []FieldExpr{
				{Key: "total", Expr: "price * qty"},
			},
			values:  map[string]float64{"price": 10, "qty": 5, "other": 99},
			want:    map[string]float64{"total": 50},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvaluateFields(tt.fields, tt.values)
			if tt.wantErr {
				if err == nil {
					t.Errorf("EvaluateFields() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("EvaluateFields() unexpected error: %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EvaluateFields() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateFieldExpressions(t *testing.T) {
	tests := []struct {
		name    string
		fields  []FieldExpr
		wantErr bool
	}{
		{
			name: "valid expressions",
			fields: []FieldExpr{
				{Key: "price", Expr: ""},
				{Key: "qty", Expr: ""},
				{Key: "total", Expr: "price * qty"},
			},
			wantErr: false,
		},
		{
			name: "valid chain",
			fields: []FieldExpr{
				{Key: "a", Expr: ""},
				{Key: "b", Expr: "a * 2"},
				{Key: "c", Expr: "b + 10"},
			},
			wantErr: false,
		},
		{
			name: "undefined field key",
			fields: []FieldExpr{
				{Key: "price", Expr: ""},
				{Key: "total", Expr: "price * nonexistent"},
			},
			wantErr: true,
		},
		{
			name: "cycle",
			fields: []FieldExpr{
				{Key: "a", Expr: "b + 1"},
				{Key: "b", Expr: "a + 1"},
			},
			wantErr: true,
		},
		{
			name: "referencing itself only",
			fields: []FieldExpr{
				{Key: "a", Expr: "a + 1"},
			},
			wantErr: true,
		},
		{
			name:    "no computed fields",
			fields:  []FieldExpr{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFieldExpressions(tt.fields)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateFieldExpressions() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateFieldExpressions() unexpected error: %v", err)
			}
		})
	}
}
