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
