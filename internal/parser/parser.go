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
		lexer.TokenType_Undefined:     parseUndefined,
		lexer.TokenType_LeftParan:     parseBlock,
		lexer.TokenType_Number:        parseNumber,
		lexer.TokenType_StringLiteral: parseString,
		lexer.TokenType_Boolean:       parseBoolean,
		lexer.TokenType_LeftBracket:   parseList,
		lexer.TokenType_Minus:         parseUnaryMinus,
	}

	operatorMap = map[int]operatorFunc{
		lexer.TokenType_Plus:               operatorAdd,
		lexer.TokenType_Minus:              operatorSubtract,
		lexer.TokenType_Asterisk:           operatorMultiply,
		lexer.TokenType_Slash:              operatorDivide,
		lexer.TokenType_Equals:             operatorEquals,
		lexer.TokenType_NotEquals:          operatorNotEquals,
		lexer.TokenType_GreaterThan:        operatorGreaterThan,
		lexer.TokenType_GreaterThanOrEqual: operatorGreaterThanOrEqual,
		lexer.TokenType_LessThan:           operatorLessThan,
		lexer.TokenType_LessThanOrEqual:    operatorLessThanOrEqual,
		lexer.TokenType_And:                operatorAnd,
		lexer.TokenType_Or:                 operatorOr,
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
		err = fmt.Errorf("unrecognizable token: %s (type: %d)", tok, tok.Type)
		return
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

	// Check for list indexing first (higher precedence)
	if tok.Type == lexer.TokenType_LeftBracket {
		return p.parseListIndex(expr)
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
	err = fmt.Errorf("%w: %v", ErrUndefinedToken, tok.Value)
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
		err = fmt.Errorf("%w RightParan, got %s", ErrExpectedToken, tok)
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

// parseString parses a string literal token.
// parseString implements parseFunc.
func parseString(_ *Parser, tok lexer.Token) (expr Expr, err error) {
	exprString := ExprString{
		Value: tok.Value,
	}

	return exprString, nil
}

// parseBoolean parses a boolean literal token.
// parseBoolean implements parseFunc.
func parseBoolean(_ *Parser, tok lexer.Token) (expr Expr, err error) {
	exprBoolean := ExprBoolean{
		Value: tok.Value == "true",
	}

	return exprBoolean, nil
}

// parseUnaryMinus parses a unary minus expression.
// parseUnaryMinus implements parseFunc.
func parseUnaryMinus(p *Parser, _ lexer.Token) (expr Expr, err error) {
	// Parse the operand after the minus
	operand, err := p.Parse()
	if err != nil {
		err = fmt.Errorf("failed to parse unary minus operand: %w", err)
		return
	}

	// Create a subtraction from zero: 0 - operand
	zero := ExprNumber{Value: decimal.NewFromInt(0)}
	return ExprSubtract{
		Expr1: zero,
		Expr2: operand,
	}, nil
}

// parseList parses a list literal token.
// parseList implements parseFunc.
func parseList(p *Parser, _ lexer.Token) (expr Expr, err error) {
	var values []Expr

	// Peek at the next token to see if it's a right bracket (empty list)
	nextTok, peekErr := p.lexer.PeekToken()
	if peekErr != nil && !errors.Is(peekErr, io.EOF) {
		err = fmt.Errorf("failed to peek token: %w", peekErr)
		return
	}

	if nextTok.Type == lexer.TokenType_RightBracket {
		// Empty list, consume the right bracket
		p.lexer.GetToken()
		return ExprList{
			Values: values,
		}, nil
	}

	// Parse the first expression
	firstExpr, parseErr := p.Parse()
	if parseErr != nil {
		err = fmt.Errorf("failed to parse first list element: %w", parseErr)
		return
	}
	values = append(values, firstExpr)

	// Check for comma-separated values
	for {
		nextTok, peekErr = p.lexer.PeekToken()
		if peekErr != nil {
			if errors.Is(peekErr, io.EOF) {
				err = fmt.Errorf("%w RightBracket, got EOF", ErrExpectedToken)
			} else {
				err = fmt.Errorf("failed to peek token: %w", peekErr)
			}
			return
		}

		if nextTok.Type == lexer.TokenType_RightBracket {
			// End of list, consume the right bracket
			p.lexer.GetToken()
			break
		}

		if nextTok.Type != lexer.TokenType_Comma {
			err = fmt.Errorf("%w comma or RightBracket in list, got %s", ErrExpectedToken, nextTok)
			return
		}

		// Consume the comma
		p.lexer.GetToken()

		// Parse the next expression
		nextExpr, parseErr := p.Parse()
		if parseErr != nil {
			err = fmt.Errorf("failed to parse list element: %w", parseErr)
			return
		}
		values = append(values, nextExpr)
	}

	return ExprList{
		Values: values,
	}, nil
}

// operatorAdd wraps two expressions in an add expression.
// operatorAdd implements operatorFunc.
func operatorAdd(expr1 Expr, expr2 Expr) (op Expr) {
	return ExprAdd{
		Expr1: expr1,
		Expr2: expr2,
	}
}

// operatorSubtract wraps two expressions in a subtract expression.
// operatorSubtract implements operatorFunc.
func operatorSubtract(expr1 Expr, expr2 Expr) (op Expr) {
	return ExprSubtract{
		Expr1: expr1,
		Expr2: expr2,
	}
}

// operatorMultiply wraps two expressions in an add expression.
// operatorMultiply implements operatorFunc.
func operatorMultiply(expr1 Expr, expr2 Expr) (op Expr) {
	return ExprMultiply{
		Expr1: expr1,
		Expr2: expr2,
	}
}

// operatorDivide wraps two expressions in a divide expression.
// operatorDivide implements operatorFunc.
func operatorDivide(expr1 Expr, expr2 Expr) (op Expr) {
	return ExprDivide{
		Expr1: expr1,
		Expr2: expr2,
	}
}

// operatorEquals wraps two expressions in an equals expression.
// operatorEquals implements operatorFunc.
func operatorEquals(expr1 Expr, expr2 Expr) (op Expr) {
	return ExprEquals{
		Expr1: expr1,
		Expr2: expr2,
	}
}

// operatorNotEquals wraps two expressions in a not equals expression.
// operatorNotEquals implements operatorFunc.
func operatorNotEquals(expr1 Expr, expr2 Expr) (op Expr) {
	return ExprNotEquals{
		Expr1: expr1,
		Expr2: expr2,
	}
}

// operatorGreaterThan wraps two expressions in a greater than expression.
// operatorGreaterThan implements operatorFunc.
func operatorGreaterThan(expr1 Expr, expr2 Expr) (op Expr) {
	return ExprGreaterThan{
		Expr1: expr1,
		Expr2: expr2,
	}
}

// operatorLessThan wraps two expressions in a less than expression.
// operatorLessThan implements operatorFunc.
func operatorLessThan(expr1 Expr, expr2 Expr) (op Expr) {
	return ExprLessThan{
		Expr1: expr1,
		Expr2: expr2,
	}
}

// operatorLessThanOrEqual wraps two expressions in a less than or equal expression.
// operatorLessThanOrEqual implements operatorFunc.
func operatorLessThanOrEqual(expr1 Expr, expr2 Expr) (op Expr) {
	return ExprLessThanOrEqual{
		Expr1: expr1,
		Expr2: expr2,
	}
}

// operatorGreaterThanOrEqual wraps two expressions in a greater than or equal expression.
// operatorGreaterThanOrEqual implements operatorFunc.
func operatorGreaterThanOrEqual(expr1 Expr, expr2 Expr) (op Expr) {
	return ExprGreaterThanOrEqual{
		Expr1: expr1,
		Expr2: expr2,
	}
}

// operatorAnd wraps two expressions in an AND expression.
// operatorAnd implements operatorFunc.
func operatorAnd(expr1 Expr, expr2 Expr) (op Expr) {
	return ExprAnd{
		Expr1: expr1,
		Expr2: expr2,
	}
}

// operatorOr wraps two expressions in an OR expression.
// operatorOr implements operatorFunc.
func operatorOr(expr1 Expr, expr2 Expr) (op Expr) {
	return ExprOr{
		Expr1: expr1,
		Expr2: expr2,
	}
}

// parseListIndex parses a list indexing operation.
func (p *Parser) parseListIndex(listExpr Expr) (expr Expr, err error) {
	if p == nil {
		err = fmt.Errorf("parser is nil")
		return
	}

	// Consume the left bracket
	_, err = p.lexer.GetToken()
	if err != nil {
		err = fmt.Errorf("failed to get left bracket token: %w", err)
		return
	}

	// Parse the index expression
	indexExpr, err := p.Parse()
	if err != nil {
		err = fmt.Errorf("failed to parse list index: %w", err)
		return
	}

	// Expect a right bracket
	tok, err := p.lexer.GetToken()
	if err != nil {
		err = fmt.Errorf("failed to get token: %w", err)
		return
	}

	if tok.Type != lexer.TokenType_RightBracket {
		err = fmt.Errorf("%w RightBracket, got %s", ErrExpectedToken, tok)
		return
	}

	// Check for chained indexing (e.g., [1,2,3][0][1])
	return p.wrapOperation(ExprListIndex{
		List:  listExpr,
		Index: indexExpr,
	})
}
