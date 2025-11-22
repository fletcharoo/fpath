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
		lexer.TokenType_Dollar:        parseInput,
		lexer.TokenType_LeftBracket:   parseList,
		lexer.TokenType_LeftBrace:     parseMapLiteral,
		lexer.TokenType_Minus:         parseUnaryMinus,
		lexer.TokenType_Label:         parseLabelOrFunction,
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
// This now handles primary expressions and then calls wrapOperation to handle binary operations.
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

	// Check for ternary operator first (lowest precedence)
	if tok.Type == lexer.TokenType_Question {
		return p.parseTernary(expr)
	}

	// Check for indexing next (higher precedence)
	if tok.Type == lexer.TokenType_LeftBracket {
		// Consume the left bracket first
		p.lexer.GetToken() // Consume the LeftBracket from tok

		// Now peek at the next token to determine operation type
		nextTok, err := p.lexer.PeekToken()
		if err != nil {
			return nil, fmt.Errorf("failed to peek token: %w", err)
		}

		// Use map indexing if:
		// 1. Expression being indexed is a map or map index
		// 2. Expression being indexed is not a list AND index is a string literal
		// 3. Expression being indexed is a list AND index is a string literal (invalid map access)
		// 4. For ExprInput, decide based on index type (string -> map, number -> list)
		// Otherwise, use list indexing
		if expr.Type() == ExprType_Map || expr.Type() == ExprType_MapIndex || nextTok.Type == lexer.TokenType_StringLiteral {
			return p.parseMapIndex(expr)
		} else if expr.Type() == ExprType_Input {
			if nextTok.Type == lexer.TokenType_StringLiteral {
				return p.parseMapIndex(expr)
			} else {
				return p.parseListIndex(expr)
			}
		} else {
			return p.parseListIndex(expr)
		}
	}

	f, ok := operatorMap[tok.Type]
	if !ok {
		return expr, nil
	}

	// This skips the peeked token.
	p.lexer.GetToken()

	// For the specific requirement of left-associative arithmetic operations without precedence,
	// we need to distinguish between arithmetic operators and other operators.
	// Arithmetic operators (+, -, *, /) should all have the same precedence and be left-associative.
	// Other operators (==, !=, <, >, <=, >=, &&, ||) should maintain relative precedence.

	// Check if the current operator is an arithmetic operator
	isArithmeticOp := isArithmeticOperator(tok.Type)

	if isArithmeticOp {
		// Parse all arithmetic operations in a left-associative way
		return p.parseArithmeticLeftAssociative(expr, f)
	} else {
		// For non-arithmetic operators, use the original logic
		// but we still need to be careful about precedence
		expr2, err := p.Parse()
		if err != nil {
			err = fmt.Errorf("failed to parse the second expression: %w", err)
		}

		// Check if the second expression is a ternary and there's no more tokens
		// This handles cases like "5 > 3 ? "greater" : "less" where ternary
		// should have lower precedence than the binary operation
		if expr2.Type() == ExprType_Ternary {
			// Check if there are any more tokens after the ternary
			nextTok, peekErr := p.lexer.PeekToken()
			if peekErr != nil || (peekErr == nil && nextTok.Type == lexer.TokenType_Undefined) {
				// No more tokens, so this should be parsed as a ternary
				// with the binary operation as the condition
				ternaryExpr := expr2.(ExprTernary)
				binaryOp := f(expr, ternaryExpr.Condition)
				return ExprTernary{
					Condition: binaryOp,
					TrueExpr:  ternaryExpr.TrueExpr,
					FalseExpr: ternaryExpr.FalseExpr,
				}, nil
			}
		}

		result, err := p.wrapOperation(f(expr, expr2))
		if err != nil {
			return nil, err
		}

		// Check for ternary operator after complete expression (lowest precedence)
		nextTok, err := p.lexer.PeekToken()
		if err == nil && nextTok.Type == lexer.TokenType_Question {
			return p.parseTernary(result)
		}

		return result, nil
	}
}

// isArithmeticOperator checks if the token type is an arithmetic operator
func isArithmeticOperator(tokenType int) bool {
	return tokenType == lexer.TokenType_Plus ||
		   tokenType == lexer.TokenType_Minus ||
		   tokenType == lexer.TokenType_Asterisk ||
		   tokenType == lexer.TokenType_Slash
}

// parseArithmeticLeftAssociative handles arithmetic operations with equal precedence
// and left-associative evaluation
func (p *Parser) parseArithmeticLeftAssociative(left Expr, leftOp operatorFunc) (Expr, error) {
	// Start with the current left expression and operator
	result := leftOp(left, left) // We'll replace the second 'left' with the actual right operand

	// For left-associative arithmetic, we'll build the expression iteratively
	// Process the right operand of the first operation
	right, err := p.parseArithmeticOperand()
	if err != nil {
		return nil, fmt.Errorf("failed to parse arithmetic operand: %w", err)
	}

	// Apply the first operation
	result = leftOp(left, right)

	// Continue looking for more arithmetic operators at the same precedence level
	for {
		nextTok, err := p.lexer.PeekToken()
		if errors.Is(io.EOF, err) {
			return result, nil
		}
		if err != nil {
			return nil, fmt.Errorf("failed to peek token: %w", err)
		}

		// If it's not an arithmetic operator, we're done
		if !isArithmeticOperator(nextTok.Type) {
			break
		}

		// Get the next arithmetic operator
		nextOp := operatorMap[nextTok.Type]
		p.lexer.GetToken() // consume the operator

		// Get the right operand
		nextRight, err := p.parseArithmeticOperand()
		if err != nil {
			return nil, fmt.Errorf("failed to parse next arithmetic operand: %w", err)
		}

		// Apply the operator left-associatively: (result op nextRight)
		result = nextOp(result, nextRight)
	}

	// After processing all arithmetic operations at this level,
	// continue with the normal precedence handling
	return p.wrapOperation(result)
}

// parseArithmeticOperand parses a single operand for arithmetic operations
// This should parse the operand without allowing arithmetic operations at the same level
func (p *Parser) parseArithmeticOperand() (Expr, error) {
	// Parse the next primary expression (number, string, parenthesized, etc.)
	tok, err := p.lexer.GetToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get token for arithmetic operand: %w", err)
	}

	f, ok := parseMap[tok.Type]
	if !ok {
		return nil, fmt.Errorf("unrecognizable token for arithmetic operand: %s (type: %d)", tok, tok.Type)
	}

	expr, err := f(p, tok)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arithmetic operand: %w", err)
	}

	// After parsing the primary, we need to handle higher precedence operations
	// like indexing, but NOT arithmetic operations
	// So we'll call wrapOperation but in a way that only processes non-arithmetic operations
	return p.wrapOperationNonArithmetic(expr)
}

// wrapOperationNonArithmetic handles operations that have higher precedence than arithmetic
// like indexing and ternary (but not arithmetic operations)
func (p *Parser) wrapOperationNonArithmetic(expr Expr) (Expr, error) {
	for {
		tok, err := p.lexer.PeekToken()
		if errors.Is(io.EOF, err) {
			return expr, nil
		}
		if err != nil {
			return nil, fmt.Errorf("failed to peek token: %w", err)
		}

		// Skip arithmetic operators - we'll return to be processed at the arithmetic level
		if isArithmeticOperator(tok.Type) {
			return expr, nil
		}

		// Handle ternary operator (lowest precedence among those we handle here)
		if tok.Type == lexer.TokenType_Question {
			return p.parseTernary(expr)
		}

		// Handle indexing (higher precedence than binary ops)
		if tok.Type == lexer.TokenType_LeftBracket {
			// Consume left bracket
			p.lexer.GetToken()

			// Peek at the index expression
			nextTok, err := p.lexer.PeekToken()
			if err != nil {
				return nil, fmt.Errorf("failed to peek token: %w", err)
			}

			var indexedExpr Expr
			var indexErr error
			if expr.Type() == ExprType_Map || expr.Type() == ExprType_MapIndex || nextTok.Type == lexer.TokenType_StringLiteral {
				indexedExpr, indexErr = p.parseMapIndex(expr)
			} else if expr.Type() == ExprType_Input {
				if nextTok.Type == lexer.TokenType_StringLiteral {
					indexedExpr, indexErr = p.parseMapIndex(expr)
				} else {
					indexedExpr, indexErr = p.parseListIndex(expr)
				}
			} else {
				indexedExpr, indexErr = p.parseListIndex(expr)
			}

			if indexErr != nil {
				return nil, indexErr
			}

			// Continue to check for more non-arithmetic operations
			expr = indexedExpr
			continue
		}

		// If it's not an arithmetic operator or higher precedence operation, return
		return expr, nil
	}
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

// parseInput parses an input data variable token.
// parseInput implements parseFunc.
func parseInput(_ *Parser, _ lexer.Token) (expr Expr, err error) {
	exprInput := ExprInput{}
	return exprInput, nil
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

// parseMapLiteral parses a map literal token.
// parseMapLiteral implements parseFunc.
func parseMapLiteral(p *Parser, _ lexer.Token) (expr Expr, err error) {
	var pairs []ExprMapPair

	// Peek at the next token to see if it's a right brace (empty map)
	nextTok, peekErr := p.lexer.PeekToken()
	if peekErr != nil && !errors.Is(peekErr, io.EOF) {
		err = fmt.Errorf("failed to peek token: %w", peekErr)
		return
	}

	if nextTok.Type == lexer.TokenType_RightBrace {
		// Empty map, consume the right brace
		p.lexer.GetToken()
		return ExprMap{
			Pairs: pairs,
		}, nil
	}

	// Parse the first key-value pair
	keyExpr, parseErr := p.Parse()
	if parseErr != nil {
		err = fmt.Errorf("failed to parse first map key: %w", parseErr)
		return
	}

	// Expect a colon
	colonTok, colonErr := p.lexer.GetToken()
	if colonErr != nil {
		err = fmt.Errorf("failed to get token: %w", colonErr)
		return
	}
	if colonTok.Type != lexer.TokenType_Colon {
		err = fmt.Errorf("%w Colon after map key, got %s", ErrExpectedToken, colonTok)
		return
	}

	// Parse the value
	valueExpr, parseErr := p.Parse()
	if parseErr != nil {
		err = fmt.Errorf("failed to parse first map value: %w", parseErr)
		return
	}

	pairs = append(pairs, ExprMapPair{
		Key:   keyExpr,
		Value: valueExpr,
	})

	// Check for comma-separated pairs
	for {
		nextTok, peekErr = p.lexer.PeekToken()
		if peekErr != nil {
			if errors.Is(peekErr, io.EOF) {
				err = fmt.Errorf("%w RightBrace, got EOF", ErrExpectedToken)
			} else {
				err = fmt.Errorf("failed to peek token: %w", peekErr)
			}
			return
		}

		if nextTok.Type == lexer.TokenType_RightBrace {
			// End of map, consume the right brace
			p.lexer.GetToken()
			break
		}

		if nextTok.Type != lexer.TokenType_Comma {
			err = fmt.Errorf("%w comma or RightBrace in map, got %s", ErrExpectedToken, nextTok)
			return
		}

		// Consume the comma
		p.lexer.GetToken()

		// Parse the next key
		nextKeyExpr, parseErr := p.Parse()
		if parseErr != nil {
			err = fmt.Errorf("failed to parse map key: %w", parseErr)
			return
		}

		// Expect a colon
		nextColonTok, colonErr := p.lexer.GetToken()
		if colonErr != nil {
			err = fmt.Errorf("failed to get token: %w", colonErr)
			return
		}
		if nextColonTok.Type != lexer.TokenType_Colon {
			err = fmt.Errorf("%w Colon after map key, got %s", ErrExpectedToken, nextColonTok)
			return
		}

		// Parse the next value
		nextValueExpr, parseErr := p.Parse()
		if parseErr != nil {
			err = fmt.Errorf("failed to parse map value: %w", parseErr)
			return
		}

		pairs = append(pairs, ExprMapPair{
			Key:   nextKeyExpr,
			Value: nextValueExpr,
		})
	}

	return ExprMap{
		Pairs: pairs,
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

// parseMapIndex parses a map indexing operation.
func (p *Parser) parseMapIndex(mapExpr Expr) (expr Expr, err error) {
	if p == nil {
		err = fmt.Errorf("parser is nil")
		return
	}

	// Parse the index expression
	indexExpr, err := p.Parse()
	if err != nil {
		err = fmt.Errorf("failed to parse map index: %w", err)
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

	// Check for chained indexing (e.g., {"a": {"b": 1}}["a"]["b"])
	return p.wrapOperation(ExprMapIndex{
		Map:   mapExpr,
		Index: indexExpr,
	})
}

// parseLabelOrFunction parses a label token, checking if it's followed by a left parenthesis to determine if it's a function call.
// parseLabelOrFunction implements parseFunc.
func parseLabelOrFunction(p *Parser, tok lexer.Token) (expr Expr, err error) {
	// Check if this is the special underscore variable
	if tok.Value == "_" {
		// This is the special underscore variable used in filter expressions
		return ExprVariable{
			Name: "_",
		}, nil
	}

	// Peek at the next token to see if it's a left parenthesis
	nextTok, peekErr := p.lexer.PeekToken()
	if peekErr != nil && !errors.Is(peekErr, io.EOF) {
		err = fmt.Errorf("failed to peek token: %w", peekErr)
		return
	}

	if nextTok.Type == lexer.TokenType_LeftParan {
		// This is a function call
		return p.parseFunction(tok.Value)
	}

	// This is just a regular label (which should be an error in current grammar)
	err = fmt.Errorf("%w: %v", ErrUndefinedToken, tok.Value)
	return
}

// parseTernary parses a ternary conditional expression.
func (p *Parser) parseTernary(conditionExpr Expr) (expr Expr, err error) {
	// Consume the question mark (already peeked)
	p.lexer.GetToken()

	// Parse the true expression
	trueExpr, err := p.Parse()
	if err != nil {
		err = fmt.Errorf("failed to parse true expression in ternary: %w", err)
		return
	}

	// Expect a colon
	colonTok, err := p.lexer.GetToken()
	if err != nil {
		err = fmt.Errorf("failed to get token: %w", err)
		return
	}
	if colonTok.Type != lexer.TokenType_Colon {
		err = fmt.Errorf("%w Colon in ternary expression, got %s", ErrExpectedToken, colonTok)
		return
	}

	// Parse the false expression
	falseExpr, err := p.Parse()
	if err != nil {
		err = fmt.Errorf("failed to parse false expression in ternary: %w", err)
		return
	}

	return ExprTernary{
		Condition: conditionExpr,
		TrueExpr:  trueExpr,
		FalseExpr: falseExpr,
	}, nil
}

// parseFunction parses a function call with the given name.
func (p *Parser) parseFunction(functionName string) (expr Expr, err error) {
	// Consume the left parenthesis
	leftParanTok, err := p.lexer.GetToken()
	if err != nil {
		err = fmt.Errorf("failed to get token: %w", err)
		return
	}
	if leftParanTok.Type != lexer.TokenType_LeftParan {
		err = fmt.Errorf("%w LeftParan, got %s", ErrExpectedToken, leftParanTok)
		return
	}

	var args []Expr

	// Peek at the next token to see if it's a right parenthesis (empty argument list)
	nextTok, peekErr := p.lexer.PeekToken()
	if peekErr != nil && !errors.Is(peekErr, io.EOF) {
		err = fmt.Errorf("failed to peek token: %w", peekErr)
		return
	}

	if nextTok.Type == lexer.TokenType_RightParan {
		// Empty argument list, consume the right parenthesis
		p.lexer.GetToken()
		return ExprFunction{
			Name: functionName,
			Args: args,
		}, nil
	}

	// Parse the first argument
	firstArg, parseErr := p.Parse()
	if parseErr != nil {
		err = fmt.Errorf("failed to parse first function argument: %w", parseErr)
		return
	}
	args = append(args, firstArg)

	// Check for comma-separated arguments
	for {
		nextTok, peekErr = p.lexer.PeekToken()
		if peekErr != nil {
			if errors.Is(peekErr, io.EOF) {
				err = fmt.Errorf("%w RightParan, got EOF", ErrExpectedToken)
			} else {
				err = fmt.Errorf("failed to peek token: %w", peekErr)
			}
			return
		}

		if nextTok.Type == lexer.TokenType_RightParan {
			// End of argument list, consume the right parenthesis
			p.lexer.GetToken()
			break
		}

		if nextTok.Type != lexer.TokenType_Comma {
			err = fmt.Errorf("%w comma or RightParan in function arguments, got %s", ErrExpectedToken, nextTok)
			return
		}

		// Consume the comma
		p.lexer.GetToken()

		// Parse the next argument
		nextArg, parseErr := p.Parse()
		if parseErr != nil {
			err = fmt.Errorf("failed to parse function argument: %w", parseErr)
			return
		}
		args = append(args, nextArg)
	}

	return ExprFunction{
		Name: functionName,
		Args: args,
	}, nil
}
