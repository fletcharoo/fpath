package parser

import (
	"errors"
	"fmt"
	"io"

	"github.com/fletcharoo/fpath/internal/lexer"
	"github.com/shopspring/decimal"
)

type (
	parseFunc    func(*Parser, lexer.Token) (Expr, error)
	operatorFunc func(Expr, Expr) Expr
)

var parseMap map[int]parseFunc
var operatorMap map[int]operatorFunc

func init() {
	parseMap = map[int]parseFunc{
		lexer.TokenType_Undefined: parseUndefined,
		lexer.TokenType_LeftParan: parseBlock,
		lexer.TokenType_Number:    parseNumber,
	}

	operatorMap = map[int]operatorFunc{
		lexer.TokenType_Plus:     operatorAdd,
		lexer.TokenType_Asterisk: operatorMultiply,
	}
}

func New(lexer *lexer.Lexer) *Parser {
	return &Parser{
		lexer: lexer,
	}
}

// Parser parses a tokenized string into an executable AST.
type Parser struct {
	lexer *lexer.Lexer
}

// Parse parses the next expression in the query.
func (p *Parser) Parse() (expr Expr, err error) {
	tok, err := p.lexer.GetToken()
	if err != nil {
		err = fmt.Errorf("failed to get token: %w", err)
		return
	}

	f, ok := parseMap[tok.Type]

	if !ok {
		err = fmt.Errorf("unrecognizable token: %s", tok)
	}

	expr, err = f(p, tok)
	if err != nil {
		err = fmt.Errorf("failed to parse: %w", err)
		return
	}

	return p.wrapOperation(expr)
}

// wrapOperation checks if the given expression is part of an operation and
// wraps it if so.
func (p *Parser) wrapOperation(expr Expr) (op Expr, err error) {
	tok, err := p.lexer.PeekToken()

	if errors.Is(io.EOF, err) {
		return expr, nil
	}

	if err != nil {
		err = fmt.Errorf("failed to peek token: %w", err)
		return
	}

	f, ok := operatorMap[tok.Type]
	if !ok {
		return expr, nil
	}

	// This skips the peeked token.
	p.lexer.GetToken()

	expr2, err := p.Parse()
	if err != nil {
		err = fmt.Errorf("failed to parse the second expression: %w", err)
	}

	return f(expr, expr2), nil
}

// parseUndefined parses an undefined token.
// parseUndefined implements parseFunc.
func parseUndefined(_ *Parser, tok lexer.Token) (expr Expr, err error) {
	err = fmt.Errorf("undefined token: %v", tok.Value)
	return
}

// parseBlock parses a blocked expression (an expression within parantheses).
// parseBlock implement parseFunc.
func parseBlock(p *Parser, _ lexer.Token) (expr Expr, err error) {
	expr, err = p.Parse()
	if err != nil {
		err = fmt.Errorf("parseBlock: %w", err)
		return
	}

	tok, err := p.lexer.GetToken()
	if err != nil {
		err = fmt.Errorf("failed to get token: %w", err)
		return
	}

	if tok.Type != lexer.TokenType_RightParan {
		err = fmt.Errorf("expected RightParan, got %s", tok)
		return
	}

	return ExprBlock{
		Expr: expr,
	}, nil
}

// parseNumber parses a number token.
// parseNumber implements parseFunc.
func parseNumber(_ *Parser, tok lexer.Token) (expr Expr, err error) {
	var exprNumber ExprNumber
	exprNumber.Value, err = decimal.NewFromString(tok.Value)
	if err != nil {
		err = fmt.Errorf("failed to parser token %q as number: %w", tok.Value, err)
	}

	return exprNumber, nil
}

// operatorAdd wraps two expressions in an add expression.
// operatorAdd implements operatorFunc.
func operatorAdd(expr1 Expr, expr2 Expr) (op Expr) {
	return ExprAdd{
		Expr1: expr1,
		Expr2: expr2,
	}
}

// operatorAdd wraps two expressions in an add expression.
// operatorAdd implements operatorFunc.
func operatorMultiply(expr1 Expr, expr2 Expr) (op Expr) {
	return ExprMultiply{
		Expr1: expr1,
		Expr2: expr2,
	}
}
