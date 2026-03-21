package expr

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

var (
	ErrInvalidExpression = errors.New("invalid expression")
	ErrDivisionByZero    = errors.New("division by zero")
	ErrUnexpectedEnd     = errors.New("unexpected end of expression")
	ErrUnexpectedChar    = errors.New("unexpected character")
)

type Parser struct {
	expr string
	pos  int
}

func NewParser(expr string) *Parser {
	expr = strings.ReplaceAll(expr, " ", "")
	return &Parser{
		expr: expr,
		pos:  0,
	}
}

func Parse(expr string) (float64, error) {
	p := NewParser(expr)
	return p.Parse()
}

func (p *Parser) Parse() (float64, error) {
	if p.pos >= len(p.expr) {
		return 0, ErrUnexpectedEnd
	}

	result, err := p.parseExpression()
	if err != nil {
		return 0, err
	}

	if p.pos < len(p.expr) {
		return 0, fmt.Errorf("%w: unexpected character '%c' at position %d", ErrInvalidExpression, p.expr[p.pos], p.pos)
	}

	return result, nil
}

func (p *Parser) parseExpression() (float64, error) {
	return p.parseAddition()
}

func (p *Parser) parseAddition() (float64, error) {
	left, err := p.parseMultiplication()
	if err != nil {
		return 0, err
	}

	for p.pos < len(p.expr) {
		op := p.expr[p.pos]
		if op != '+' && op != '-' {
			break
		}
		p.pos++

		right, err := p.parseMultiplication()
		if err != nil {
			return 0, err
		}

		if op == '+' {
			left += right
		} else {
			left -= right
		}
	}

	return left, nil
}

func (p *Parser) parseMultiplication() (float64, error) {
	left, err := p.parseUnary()
	if err != nil {
		return 0, err
	}

	for p.pos < len(p.expr) {
		op := p.expr[p.pos]
		if op != '*' && op != '/' {
			break
		}
		p.pos++

		right, err := p.parseUnary()
		if err != nil {
			return 0, err
		}

		if op == '*' {
			left *= right
		} else {
			if right == 0 {
				return 0, ErrDivisionByZero
			}
			left /= right
		}
	}

	return left, nil
}

func (p *Parser) parseUnary() (float64, error) {
	if p.pos >= len(p.expr) {
		return 0, ErrUnexpectedEnd
	}

	if p.expr[p.pos] == '-' {
		p.pos++
		val, err := p.parseUnary()
		return -val, err
	}

	if p.expr[p.pos] == '+' {
		p.pos++
		return p.parseUnary()
	}

	return p.parsePrimary()
}

func (p *Parser) parsePrimary() (float64, error) {
	if p.pos >= len(p.expr) {
		return 0, ErrUnexpectedEnd
	}

	start := p.pos
	c := p.expr[p.pos]

	if unicode.IsDigit(rune(c)) || c == '.' {
		return p.parseNumber()
	}

	if c == '(' {
		p.pos++
		val, err := p.parseExpression()
		if err != nil {
			return 0, err
		}

		if p.pos >= len(p.expr) || p.expr[p.pos] != ')' {
			return 0, fmt.Errorf("%w: expected ')'", ErrInvalidExpression)
		}
		p.pos++
		return val, nil
	}

	return 0, fmt.Errorf("%w: '%c' at position %d", ErrUnexpectedChar, c, start)
}

func (p *Parser) parseNumber() (float64, error) {
	start := p.pos

	for p.pos < len(p.expr) && (unicode.IsDigit(rune(p.expr[p.pos])) || p.expr[p.pos] == '.') {
		p.pos++
	}

	if p.pos == start {
		return 0, fmt.Errorf("%w: no digits found", ErrInvalidExpression)
	}

	numStr := p.expr[start:p.pos]
	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: '%s'", ErrInvalidExpression, numStr)
	}

	return val, nil
}
