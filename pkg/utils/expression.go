package utils

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// EvaluateExpression evaluates a simple arithmetic expression with given variable values.
// Supports: +, -, *, /, parentheses, and variable names matching [a-zA-Z_][a-zA-Z0-9_]*
func EvaluateExpression(expr string, vars map[string]float64) (float64, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return 0, nil
	}

	tokens, err := tokenize(expr, vars)
	if err != nil {
		return 0, err
	}

	p := &parser{tokens: tokens}
	result, err := p.parseExpression()
	if err != nil {
		return 0, err
	}

	if p.pos < len(p.tokens) {
		return 0, fmt.Errorf("unexpected token after expression")
	}

	return result, nil
}

type tokenType int

const (
	tokNumber tokenType = iota
	tokPlus
	tokMinus
	tokStar
	tokSlash
	tokLParen
	tokRParen
)

type token struct {
	typ   tokenType
	value float64
}

func tokenize(expr string, vars map[string]float64) ([]token, error) {
	var tokens []token
	i := 0
	runes := []rune(expr)

	for i < len(runes) {
		ch := runes[i]

		switch {
		case unicode.IsSpace(ch):
			i++

		case ch == '+':
			tokens = append(tokens, token{typ: tokPlus})
			i++

		case ch == '-':
			tokens = append(tokens, token{typ: tokMinus})
			i++

		case ch == '*':
			tokens = append(tokens, token{typ: tokStar})
			i++

		case ch == '/':
			tokens = append(tokens, token{typ: tokSlash})
			i++

		case ch == '(':
			tokens = append(tokens, token{typ: tokLParen})
			i++

		case ch == ')':
			tokens = append(tokens, token{typ: tokRParen})
			i++

		case ch == '.' || unicode.IsDigit(ch):
			start := i
			for i < len(runes) && (unicode.IsDigit(runes[i]) || runes[i] == '.') {
				i++
			}
			val, err := strconv.ParseFloat(string(runes[start:i]), 64)
			if err != nil {
				return nil, fmt.Errorf("invalid number: %s", string(runes[start:i]))
			}
			tokens = append(tokens, token{typ: tokNumber, value: val})

		case unicode.IsLetter(ch) || ch == '_':
			start := i
			for i < len(runes) && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '_') {
				i++
			}
			name := string(runes[start:i])
			val, ok := vars[name]
			if !ok {
				return nil, fmt.Errorf("undefined variable: %s", name)
			}
			tokens = append(tokens, token{typ: tokNumber, value: val})

		default:
			return nil, fmt.Errorf("unexpected character: %c", ch)
		}
	}

	return tokens, nil
}

type parser struct {
	tokens []token
	pos    int
}

func (p *parser) peek() (*token, bool) {
	if p.pos >= len(p.tokens) {
		return nil, false
	}
	return &p.tokens[p.pos], true
}

func (p *parser) consume() {
	p.pos++
}

// parseExpression handles addition and subtraction
func (p *parser) parseExpression() (float64, error) {
	left, err := p.parseTerm()
	if err != nil {
		return 0, err
	}

	for {
		tok, ok := p.peek()
		if !ok {
			break
		}

		switch tok.typ {
		case tokPlus:
			p.consume()
			right, err := p.parseTerm()
			if err != nil {
				return 0, err
			}
			left += right
		case tokMinus:
			p.consume()
			right, err := p.parseTerm()
			if err != nil {
				return 0, err
			}
			left -= right
		default:
			return left, nil
		}
	}

	return left, nil
}

// parseTerm handles multiplication and division
func (p *parser) parseTerm() (float64, error) {
	left, err := p.parseFactor()
	if err != nil {
		return 0, err
	}

	for {
		tok, ok := p.peek()
		if !ok {
			break
		}

		switch tok.typ {
		case tokStar:
			p.consume()
			right, err := p.parseFactor()
			if err != nil {
				return 0, err
			}
			left *= right
		case tokSlash:
			p.consume()
			right, err := p.parseFactor()
			if err != nil {
				return 0, err
			}
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			left /= right
		default:
			return left, nil
		}
	}

	return left, nil
}

// parseFactor handles numbers, parenthesized expressions, and unary minus
func (p *parser) parseFactor() (float64, error) {
	tok, ok := p.peek()
	if !ok {
		return 0, fmt.Errorf("unexpected end of expression")
	}

	// Handle unary minus
	if tok.typ == tokMinus {
		p.consume()
		val, err := p.parseFactor()
		if err != nil {
			return 0, err
		}
		return -val, nil
	}

	if tok.typ == tokNumber {
		p.consume()
		return tok.value, nil
	}

	if tok.typ == tokLParen {
		p.consume()
		val, err := p.parseExpression()
		if err != nil {
			return 0, err
		}

		tok, ok = p.peek()
		if !ok || tok.typ != tokRParen {
			return 0, fmt.Errorf("missing closing parenthesis")
		}
		p.consume()
		return val, nil
	}

	return 0, fmt.Errorf("unexpected token")
}

// FieldExpr represents a field that may have a computed expression.
type FieldExpr struct {
	Key  string // field key, used as variable name in expressions
	Expr string // arithmetic expression; empty if manually entered
}

// EvaluateFields resolves computed fields in dependency order.
// It topologically sorts fields by their expression dependencies, then evaluates
// each in order. Computed fields whose dependencies are unsatisfied are silently
// omitted from the result (empty-on-incomplete). Returns an error if a cycle is detected.
func EvaluateFields(fields []FieldExpr, values map[string]float64) (map[string]float64, error) {
	if len(fields) == 0 {
		return map[string]float64{}, nil
	}

	result := make(map[string]float64)

	// Copy input values so we don't mutate the caller's map
	allValues := make(map[string]float64, len(values))
	for k, v := range values {
		allValues[k] = v
	}

	// Build dependency graph
	// indegree counts how many unresolved deps each field has
	indegree := make(map[string]int)
	dependents := make(map[string][]string) // field → fields that depend on it
	exprMap := make(map[string]string)

	for _, f := range fields {
		exprMap[f.Key] = f.Expr
	}

	for _, f := range fields {
		if f.Expr == "" {
			continue
		}
		deps := extractVariables(f.Expr, exprMap)
		indegree[f.Key] = len(deps)
		for _, dep := range deps {
			dependents[dep] = append(dependents[dep], f.Key)
		}
	}

	// Topological sort via Kahn's algorithm
	queue := make([]string, 0)
	for _, f := range fields {
		if f.Expr != "" && indegree[f.Key] == 0 {
			queue = append(queue, f.Key)
		}
	}

	evaluated := 0
	for len(queue) > 0 {
		key := queue[0]
		queue = queue[1:]

		expr := exprMap[key]
		val, err := EvaluateExpression(expr, allValues)
		if err != nil {
			// Dependency not satisfied — skip (empty-on-incomplete)
			continue
		}
		allValues[key] = val
		result[key] = val
		evaluated++

		for _, dep := range dependents[key] {
			indegree[dep]--
			if indegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	// Check for cycles: if some fields still have positive indegree
	for _, f := range fields {
		if f.Expr != "" && indegree[f.Key] > 0 {
			return nil, fmt.Errorf("cycle detected in field expressions involving %q", f.Key)
		}
	}

	return result, nil
}

// extractVariables extracts field key references from an expression string.
// It returns only keys that are also computed fields (appear in exprMap),
// since manually-entered fields are already in the values map.
func extractVariables(expr string, exprMap map[string]string) []string {
	var deps []string
	seen := make(map[string]bool)

	i := 0
	runes := []rune(expr)
	for i < len(runes) {
		ch := runes[i]
		if unicode.IsLetter(ch) || ch == '_' {
			start := i
			for i < len(runes) && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '_') {
				i++
			}
			name := string(runes[start:i])
			// Only treat as a dependency if it's another computed field
			if _, isComputed := exprMap[name]; isComputed && !seen[name] {
				deps = append(deps, name)
				seen[name] = true
			}
		} else {
			i++
		}
	}
	return deps
}

// extractAllVariables extracts all variable references from an expression string.
func extractAllVariables(expr string) []string {
	var vars []string
	seen := make(map[string]bool)

	i := 0
	runes := []rune(expr)
	for i < len(runes) {
		ch := runes[i]
		if unicode.IsLetter(ch) || ch == '_' {
			start := i
			for i < len(runes) && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '_') {
				i++
			}
			name := string(runes[start:i])
			if !seen[name] {
				vars = append(vars, name)
				seen[name] = true
			}
		} else {
			i++
		}
	}
	return vars
}

// ValidateFieldExpressions validates that all field expressions reference valid keys
// and that there are no cyclic dependencies.
func ValidateFieldExpressions(fields []FieldExpr) error {
	if len(fields) == 0 {
		return nil
	}

	// Build set of all known field keys
	knownKeys := make(map[string]bool)
	exprMap := make(map[string]string)
	for _, f := range fields {
		knownKeys[f.Key] = true
		if f.Expr != "" {
			exprMap[f.Key] = f.Expr
		}
	}

	// Validate all variable references exist as field keys
	for _, f := range fields {
		if f.Expr == "" {
			continue
		}
		for _, v := range extractAllVariables(f.Expr) {
			if !knownKeys[v] {
				return fmt.Errorf("expression for %q references undefined field %q", f.Key, v)
			}
		}
	}

	// Check for cycles using Kahn's algorithm on the dependency graph
	indegree := make(map[string]int)
	dependents := make(map[string][]string)

	for key, expr := range exprMap {
		for _, dep := range extractVariables(expr, exprMap) {
			indegree[key]++
			dependents[dep] = append(dependents[dep], key)
		}
	}

	queue := make([]string, 0)
	for key := range exprMap {
		if indegree[key] == 0 {
			queue = append(queue, key)
		}
	}

	processed := 0
	for len(queue) > 0 {
		key := queue[0]
		queue = queue[1:]
		processed++
		for _, dep := range dependents[key] {
			indegree[dep]--
			if indegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	if processed < len(exprMap) {
		for key := range exprMap {
			if indegree[key] > 0 {
				return fmt.Errorf("cycle detected in field expressions involving %q", key)
			}
		}
	}

	return nil
}
