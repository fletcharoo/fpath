package fpath

import (
	"fmt"

	"github.com/fletcharoo/fpath/internal/lexer"
	"github.com/fletcharoo/fpath/internal/parser"
	"github.com/fletcharoo/fpath/internal/runtime"
)

// Query represents a compiled fpath expression that can be evaluated multiple times
// with different input data. The Query type is opaque to external users.
type Query struct {
	expr parser.Expr
}

// Compile parses and validates an fpath query string, returning a Query that
// can be evaluated multiple times with different input data.
//
// The query string follows the fpath expression syntax, supporting:
// - Arithmetic operations: +, -, *, /, //, %, ^ (evaluated left-to-right)
// - Comparison operations: ==, !=, <, <=, >, >=
// - Logical operations: &&, ||
// - Ternary conditional: condition ? true_expr : false_expr
// - Indexing: list[index], map[key], string[index]
// - Slicing: list[start:end], string[start:end]
// - Functions: len(), filter(), contains(), abs(), min(), max(), round(), floor(), ceil()
// - Literals: numbers, strings, booleans, lists, maps
// - Input data reference: $
//
// Example:
//
//	query, err := Compile("$.items[0].price * 1.1")
//	if err != nil {
//		return err
//	}
//	result, err := query.Evaluate(inputData)
func Compile(query string) (*Query, error) {
	if query == "" {
		return nil, fmt.Errorf("empty query string")
	}

	// Create lexer and tokenize the input
	l := lexer.New(query)

	// Create parser and parse the tokens into an AST
	p := parser.New(l)
	expr, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to compile query: %w", err)
	}

	return &Query{
		expr: expr,
	}, nil
}

// Evaluate executes the compiled query against the provided input data and returns the result.
//
// The input data can be any Go value that the fpath expression can operate on:
// - Maps (map[string]any or struct types)
// - Slices and arrays
// - Primitive types (string, number, boolean)
// - Nested combinations of the above
//
// The $ symbol in the expression refers to the input data.
//
// Example:
//
//	query, _ := Compile("$.name")
//	result, err := query.Evaluate(map[string]any{"name": "Alice"})
//	// result == "Alice"
//
// expressionToGoValue recursively converts expression objects to native Go types
func expressionToGoValue(expr parser.Expr) (any, error) {
	if expr == nil {
		return nil, nil
	}

	// First decode the expression to get its basic value
	decoded, err := expr.Decode()
	if err != nil {
		return nil, err
	}

	// For simple types, return as-is
	switch expr.Type() {
	case parser.ExprType_Number, parser.ExprType_String, parser.ExprType_Boolean:
		return decoded, nil

	case parser.ExprType_List:
		// Convert ExprList to []any by recursively converting each element
		exprList, ok := expr.(parser.ExprList)
		if !ok {
			return nil, fmt.Errorf("failed to assert expression as list")
		}

		var result []any
		for _, elementExpr := range exprList.Values {
			elementValue, err := expressionToGoValue(elementExpr)
			if err != nil {
				return nil, fmt.Errorf("failed to convert list element: %w", err)
			}
			result = append(result, elementValue)
		}
		return result, nil

	case parser.ExprType_Map:
		// Convert ExprMap to map[string]any by recursively converting each key-value pair
		exprMap, ok := expr.(parser.ExprMap)
		if !ok {
			return nil, fmt.Errorf("failed to assert expression as map")
		}

		result := make(map[string]any)
		for _, pair := range exprMap.Pairs {
			// Convert key to string
			keyValue, err := expressionToGoValue(pair.Key)
			if err != nil {
				return nil, fmt.Errorf("failed to convert map key: %w", err)
			}

			keyStr, ok := keyValue.(string)
			if !ok {
				return nil, fmt.Errorf("map key must be string, got %T", keyValue)
			}

			// Convert value
			valueValue, err := expressionToGoValue(pair.Value)
			if err != nil {
				return nil, fmt.Errorf("failed to convert map value: %w", err)
			}

			result[keyStr] = valueValue
		}
		return result, nil

	default:
		// For other expression types (like ExprInput, ExprBlock, etc.),
		// try to decode them directly
		return decoded, nil
	}
}

func (q *Query) Evaluate(input any) (any, error) {
	if q == nil {
		return nil, fmt.Errorf("query is nil")
	}

	// Evaluate the compiled expression against the input data
	resultExpr, err := runtime.Eval(q.expr, input)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate query: %w", err)
	}

	// Convert the expression result to a native Go value
	result, err := expressionToGoValue(resultExpr)
	if err != nil {
		return nil, fmt.Errorf("failed to convert result to Go value: %w", err)
	}

	return result, nil
}
