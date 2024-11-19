package parser

import (
	"fmt"

	"github.com/fletcharoo/fpath/internal/lexer"
	"github.com/shopspring/decimal"
)

type parseFunc func(lexer.Token) (Expr, error)

var parseMap map[int]parseFunc

func init() {
	parseMap = map[int]parseFunc{
		ExprType_Undefined: parseUndefined,
		ExprType_Number:    parseNumber,
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

	return f(tok)
}

// parseUndefined parses an undefined token.
// parseUndefined implements parseFunc.
func parseUndefined(tok lexer.Token) (expr Expr, err error) {
	return nil, fmt.Errorf("undefined token: %v", tok.Value)
}

// parseNumber parses a number token.
// parseNumber implements parseFunc.
func parseNumber(tok lexer.Token) (expr Expr, err error) {
	var exprNumber ExprNumber
	exprNumber.Value, err = decimal.NewFromString(tok.Value)
	if err != nil {
		err = fmt.Errorf("failed to parser token %q as number: %w", tok.Value, err)
	}

	return exprNumber, nil
}
