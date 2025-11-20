package runtime

import (
	"fmt"

	"github.com/fletcharoo/fpath/internal/parser"
)

type evalFunc func(parser.Expr, any) (parser.Expr, error)

var evalMap map[int]evalFunc

func init() {
	evalMap = map[int]evalFunc{
		parser.ExprType_Undefined: evalUndefined,
		parser.ExprType_Block:     evalBlock,
		parser.ExprType_Number:    evalLiteral,
		parser.ExprType_String:    evalString,
		parser.ExprType_Add:       evalAdd,
		parser.ExprType_Subtract:  evalSubtract,
		parser.ExprType_Multiply:  evalMultiply,
		parser.ExprType_Divide:    evalDivide,
	}
}

// Eval accepts a parsed expression and the query's input data and returns the
// evaluated result
func Eval(expr parser.Expr, input any) (result parser.Expr, err error) {
	f, ok := evalMap[expr.Type()]
	if !ok {
		return evalUndefined(nil, nil)
	}

	return f(expr, input)
}

// evalUndefined returns an undefined error.
func evalUndefined(_ parser.Expr, _ any) (ret parser.Expr, err error) {
	err = fmt.Errorf("failed to eval undefined expression")
	return
}

// evalLiteral returns the expression passed into it.
func evalLiteral(expr parser.Expr, _ any) (ret parser.Expr, err error) {
	return expr, nil
}

// evalString returns the string expression passed into it.
func evalString(expr parser.Expr, _ any) (ret parser.Expr, err error) {
	return expr, nil
}

// evalLiteral evaluates the contained expression.
func evalBlock(expr parser.Expr, input any) (ret parser.Expr, err error) {
	exprBlock, ok := expr.(parser.ExprBlock)
	if !ok {
		err = fmt.Errorf("failed to assert expression as block")
		return
	}

	return Eval(exprBlock.Expr, input)
}

// evalAdd accepts a parser.ExprAdd expression and performs the operation.
func evalAdd(expr parser.Expr, input any) (ret parser.Expr, err error) {
	exprAdd, ok := expr.(parser.ExprAdd)
	if !ok {
		err = fmt.Errorf("failed to assert expression as add")
		return
	}

	expr1, err := Eval(exprAdd.Expr1, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate first expression: %w", err)
		return
	}

	expr2, err := Eval(exprAdd.Expr2, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate second expression: %w", err)
		return
	}

	expr1Type := expr1.Type()
	expr2Type := expr2.Type()
	if expr1Type != expr2Type {
		err = fmt.Errorf("incompatible types: %s and %s", expr1, expr2)
		return
	}

	switch expr1Type {
	case parser.ExprType_Number:
		return evalAddNumber(expr1, expr2)
	case parser.ExprType_String:
		return evalAddString(expr1, expr2)
	default:
		err = fmt.Errorf("invalid add type: %s", expr1)
		return
	}
}

// evalAddNumber accepts two paser.ExprNumber expressions and adds them
// together.
func evalAddNumber(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
	expr1Number, ok := expr1.(parser.ExprNumber)
	if !ok {
		err = fmt.Errorf("failed to assert first expression as number")
		return
	}

	expr2Number, ok := expr2.(parser.ExprNumber)
	if !ok {
		err = fmt.Errorf("failed to assert second expression as number")
		return
	}

	resultNumber := parser.ExprNumber{
		Value: expr1Number.Value.Add(expr2Number.Value),
	}

	return resultNumber, nil
}

// evalAddString accepts two parser.ExprString expressions and concatenates
// them together.
func evalAddString(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
	expr1String, ok := expr1.(parser.ExprString)
	if !ok {
		err = fmt.Errorf("failed to assert first expression as string")
		return
	}

	expr2String, ok := expr2.(parser.ExprString)
	if !ok {
		err = fmt.Errorf("failed to assert second expression as string")
		return
	}

	resultString := parser.ExprString{
		Value: expr1String.Value + expr2String.Value,
	}

	return resultString, nil
}

// evalSubtract accepts a parser.ExprSubtract expression and performs the operation.
func evalSubtract(expr parser.Expr, input any) (ret parser.Expr, err error) {
	exprSubtract, ok := expr.(parser.ExprSubtract)
	if !ok {
		err = fmt.Errorf("failed to assert expression as subtract")
		return
	}

	// Handle left-associativity for chained subtraction
	// If the second expression is also a subtract, we need to evaluate left-to-right
	if nestedSubtract, isNested := exprSubtract.Expr2.(parser.ExprSubtract); isNested {
		// Evaluate (a - (b - c)) as ((a - b) - c)
		// First evaluate a - b
		leftResult, err := evalSubtract(parser.ExprSubtract{
			Expr1: exprSubtract.Expr1,
			Expr2: nestedSubtract.Expr1,
		}, input)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate left part of chained subtraction: %w", err)
		}

		// Then evaluate (a - b) - c
		return evalSubtract(parser.ExprSubtract{
			Expr1: leftResult,
			Expr2: nestedSubtract.Expr2,
		}, input)
	}

	expr1, err := Eval(exprSubtract.Expr1, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate first expression: %w", err)
		return
	}

	expr2, err := Eval(exprSubtract.Expr2, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate second expression: %w", err)
		return
	}

	expr1Type := expr1.Type()
	expr2Type := expr2.Type()
	if expr1Type != expr2Type {
		err = fmt.Errorf("incompatible types: %s and %s", expr1, expr2)
		return
	}

	switch expr1Type {
	case parser.ExprType_Number:
		return evalSubtractNumber(expr1, expr2)
	default:
		err = fmt.Errorf("invalid subtract type: %s", expr1)
		return
	}
}

// evalSubtractNumber accepts two parser.ExprNumber expressions and subtracts
// the second from the first.
func evalSubtractNumber(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
	expr1Number, ok := expr1.(parser.ExprNumber)
	if !ok {
		err = fmt.Errorf("failed to assert first expression as number")
		return
	}

	expr2Number, ok := expr2.(parser.ExprNumber)
	if !ok {
		err = fmt.Errorf("failed to assert second expression as number")
		return
	}

	resultNumber := parser.ExprNumber{
		Value: expr1Number.Value.Sub(expr2Number.Value),
	}

	return resultNumber, nil
}

// evalMultoply accepts a parser.ExprMultiply expression and performs the operation.
func evalMultiply(expr parser.Expr, input any) (ret parser.Expr, err error) {
	exprAdd, ok := expr.(parser.ExprMultiply)
	if !ok {
		err = fmt.Errorf("failed to assert expression as multiply")
		return
	}

	expr1, err := Eval(exprAdd.Expr1, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate first expression: %w", err)
		return
	}

	expr2, err := Eval(exprAdd.Expr2, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate second expression: %w", err)
		return
	}

	expr1Type := expr1.Type()
	expr2Type := expr2.Type()
	if expr1Type != expr2Type {
		err = fmt.Errorf("incompatible types: %s and %s", expr1, expr2)
		return
	}

	switch expr1Type {
	case parser.ExprType_Number:
		return evalMultiplyNumber(expr1, expr2)
	default:
		err = fmt.Errorf("invalid add type: %s", expr1)
		return
	}
}

// evalMultiplyNumber accepts two paser.ExprNumber expressions and multiplies
// them together.
func evalMultiplyNumber(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
	expr1Number, ok := expr1.(parser.ExprNumber)
	if !ok {
		err = fmt.Errorf("failed to assert first expression as number")
		return
	}

	expr2Number, ok := expr2.(parser.ExprNumber)
	if !ok {
		err = fmt.Errorf("failed to assert second expression as number")
		return
	}

	resultNumber := parser.ExprNumber{
		Value: expr1Number.Value.Mul(expr2Number.Value),
	}

	return resultNumber, nil
}

// evalDivide accepts a parser.ExprDivide expression and performs the operation.
func evalDivide(expr parser.Expr, input any) (ret parser.Expr, err error) {
	exprDivide, ok := expr.(parser.ExprDivide)
	if !ok {
		err = fmt.Errorf("failed to assert expression as divide")
		return
	}

	// Handle left-associativity for chained division
	// If the second expression is also a divide, we need to evaluate left-to-right
	if nestedDivide, isNested := exprDivide.Expr2.(parser.ExprDivide); isNested {
		// Evaluate (a / (b / c)) as ((a / b) / c)
		// First evaluate a / b
		leftResult, err := evalDivide(parser.ExprDivide{
			Expr1: exprDivide.Expr1,
			Expr2: nestedDivide.Expr1,
		}, input)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate left part of chained division: %w", err)
		}

		// Then evaluate (a / b) / c
		return evalDivide(parser.ExprDivide{
			Expr1: leftResult,
			Expr2: nestedDivide.Expr2,
		}, input)
	}

	expr1, err := Eval(exprDivide.Expr1, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate first expression: %w", err)
		return
	}

	expr2, err := Eval(exprDivide.Expr2, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate second expression: %w", err)
		return
	}

	expr1Type := expr1.Type()
	expr2Type := expr2.Type()
	if expr1Type != expr2Type {
		err = fmt.Errorf("incompatible types: %s and %s", expr1, expr2)
		return
	}

	switch expr1Type {
	case parser.ExprType_Number:
		return evalDivideNumber(expr1, expr2)
	default:
		err = fmt.Errorf("invalid divide type: %s", expr1)
		return
	}
}

// evalDivideNumber accepts two parser.ExprNumber expressions and divides
// the first by the second.
func evalDivideNumber(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
	expr1Number, ok := expr1.(parser.ExprNumber)
	if !ok {
		err = fmt.Errorf("failed to assert first expression as number")
		return
	}

	expr2Number, ok := expr2.(parser.ExprNumber)
	if !ok {
		err = fmt.Errorf("failed to assert second expression as number")
		return
	}

	// Check for division by zero
	if expr2Number.Value.IsZero() {
		err = fmt.Errorf("division by zero")
		return
	}

	resultNumber := parser.ExprNumber{
		Value: expr1Number.Value.Div(expr2Number.Value),
	}

	return resultNumber, nil
}
