package utils

import "testing"

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
