package runtime

import (
	"errors"
	"fmt"

	"github.com/fletcharoo/fpath/internal/parser"
	"github.com/shopspring/decimal"
)

var (
	ErrIncompatibleTypes    = errors.New("incompatible types")
	ErrDivisionByZero       = errors.New("division by zero")
	ErrBooleanOperation     = errors.New("boolean operation requires boolean expressions")
	ErrIndexOutOfBounds     = errors.New("list index out of bounds")
	ErrInvalidIndex         = errors.New("invalid list index")
	ErrKeyNotFound          = errors.New("map key not found")
	ErrInvalidMapIndex      = errors.New("invalid map index")
	ErrUndefinedFunction    = errors.New("undefined function")
	ErrInvalidArgumentCount = errors.New("invalid argument count")
	ErrInvalidArgumentType  = errors.New("invalid argument type")
)

type evalFunc func(parser.Expr, any) (parser.Expr, error)
type functionFunc func([]parser.Expr, any) (parser.Expr, error)

var evalFuncMap map[int]evalFunc
var functionRegistry map[string]functionFunc

func init() {
	evalFuncMap = map[int]evalFunc{
		parser.ExprType_Undefined:          evalUndefined,
		parser.ExprType_Block:              evalBlock,
		parser.ExprType_Number:             evalLiteral,
		parser.ExprType_String:             evalString,
		parser.ExprType_Input:              evalInput,
		parser.ExprType_Variable:           evalVariable,
		parser.ExprType_Boolean:            evalLiteral,
		parser.ExprType_Add:                evalAdd,
		parser.ExprType_Subtract:           evalSubtract,
		parser.ExprType_Multiply:           evalMultiply,
		parser.ExprType_Divide:             evalDivide,
		parser.ExprType_Modulo:             evalModulo,
		parser.ExprType_Equals:             evalEquals,
		parser.ExprType_NotEquals:          evalNotEquals,
		parser.ExprType_GreaterThan:        evalGreaterThan,
		parser.ExprType_GreaterThanOrEqual: evalGreaterThanOrEqual,
		parser.ExprType_LessThan:           evalLessThan,
		parser.ExprType_LessThanOrEqual:    evalLessThanOrEqual,
		parser.ExprType_And:                evalAnd,
		parser.ExprType_Or:                 evalOr,
		parser.ExprType_Ternary:            evalTernary,
		parser.ExprType_List:               evalList,
		parser.ExprType_ListIndex:          evalListIndex,
		parser.ExprType_Map:                evalMap,
		parser.ExprType_MapIndex:           evalMapIndex,
		parser.ExprType_Function:           evalFunction,
	}

	functionRegistry = map[string]functionFunc{
		"len": evalLenFunction,
		"filter": evalFilterFunction,
	}
}

// Eval accepts a parsed expression and the query's input data and returns the
// evaluated result
func Eval(expr parser.Expr, input any) (result parser.Expr, err error) {
	f, ok := evalFuncMap[expr.Type()]
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

// evalInput converts input data to appropriate expression types.
func evalInput(_ parser.Expr, input any) (ret parser.Expr, err error) {
	return convertInputToExpr(input)
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
		err = fmt.Errorf("%w: %s and %s", ErrIncompatibleTypes, expr1, expr2)
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
		err = fmt.Errorf("%w: %s and %s", ErrIncompatibleTypes, expr1, expr2)
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
		err = fmt.Errorf("%w: %s and %s", ErrIncompatibleTypes, expr1, expr2)
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
		err = fmt.Errorf("%w: %s and %s", ErrIncompatibleTypes, expr1, expr2)
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
		err = ErrDivisionByZero
		return
	}

	resultNumber := parser.ExprNumber{
		Value: expr1Number.Value.Div(expr2Number.Value),
	}

	return resultNumber, nil
}

// evalModulo accepts a parser.ExprModulo expression and performs the operation.
func evalModulo(expr parser.Expr, input any) (ret parser.Expr, err error) {
	exprModulo, ok := expr.(parser.ExprModulo)
	if !ok {
		err = fmt.Errorf("failed to assert expression as modulo")
		return
	}

	// Handle left-associativity for chained modulo operations
	// If the second expression is also a modulo, we need to evaluate left-to-right
	if nestedModulo, isNested := exprModulo.Expr2.(parser.ExprModulo); isNested {
		// Evaluate (a % (b % c)) as ((a % b) % c)
		// First evaluate a % b
		leftResult, err := evalModulo(parser.ExprModulo{
			Expr1: exprModulo.Expr1,
			Expr2: nestedModulo.Expr1,
		}, input)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate left part of chained modulo: %w", err)
		}

		// Then evaluate (a % b) % c
		return evalModulo(parser.ExprModulo{
			Expr1: leftResult,
			Expr2: nestedModulo.Expr2,
		}, input)
	}

	expr1, err := Eval(exprModulo.Expr1, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate first expression: %w", err)
		return
	}

	expr2, err := Eval(exprModulo.Expr2, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate second expression: %w", err)
		return
	}

	expr1Type := expr1.Type()
	expr2Type := expr2.Type()
	if expr1Type != expr2Type {
		err = fmt.Errorf("%w: %s and %s", ErrIncompatibleTypes, expr1, expr2)
		return
	}

	switch expr1Type {
	case parser.ExprType_Number:
		return evalModuloNumber(expr1, expr2)
	default:
		err = fmt.Errorf("invalid modulo type: %s", expr1)
		return
	}
}

// evalModuloNumber accepts two parser.ExprNumber expressions and performs the modulo operation.
func evalModuloNumber(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
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

	// Check for division by zero (modulo by zero)
	if expr2Number.Value.IsZero() {
		err = ErrDivisionByZero
		return
	}

	resultNumber := parser.ExprNumber{
		Value: expr1Number.Value.Mod(expr2Number.Value),
	}

	return resultNumber, nil
}

// evalEquals accepts a parser.ExprEquals expression and performs the equality comparison.
func evalEquals(expr parser.Expr, input any) (ret parser.Expr, err error) {
	exprEquals, ok := expr.(parser.ExprEquals)
	if !ok {
		err = fmt.Errorf("failed to assert expression as equals")
		return
	}

	expr1, err := Eval(exprEquals.Expr1, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate first expression: %w", err)
		return
	}

	expr2, err := Eval(exprEquals.Expr2, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate second expression: %w", err)
		return
	}

	expr1Type := expr1.Type()
	expr2Type := expr2.Type()
	if expr1Type != expr2Type {
		err = fmt.Errorf("%w: %s and %s", ErrIncompatibleTypes, expr1, expr2)
		return
	}

	switch expr1Type {
	case parser.ExprType_Number:
		return evalEqualsNumber(expr1, expr2)
	case parser.ExprType_String:
		return evalEqualsString(expr1, expr2)
	case parser.ExprType_Boolean:
		return evalEqualsBoolean(expr1, expr2)
	default:
		err = fmt.Errorf("invalid equals type: %s", expr1)
		return
	}
}

// evalEqualsNumber accepts two parser.ExprNumber expressions and compares them for equality.
func evalEqualsNumber(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
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

	isEqual := expr1Number.Value.Equal(expr2Number.Value)

	resultBoolean := parser.ExprBoolean{
		Value: isEqual,
	}

	return resultBoolean, nil
}

// evalEqualsString accepts two parser.ExprString expressions and compares them for equality.
func evalEqualsString(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
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

	isEqual := expr1String.Value == expr2String.Value

	resultBoolean := parser.ExprBoolean{
		Value: isEqual,
	}

	return resultBoolean, nil
}

// evalEqualsBoolean accepts two parser.ExprBoolean expressions and compares them for equality.
func evalEqualsBoolean(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
	expr1Boolean, ok := expr1.(parser.ExprBoolean)
	if !ok {
		err = fmt.Errorf("failed to assert first expression as boolean")
		return
	}

	expr2Boolean, ok := expr2.(parser.ExprBoolean)
	if !ok {
		err = fmt.Errorf("failed to assert second expression as boolean")
		return
	}

	isEqual := expr1Boolean.Value == expr2Boolean.Value

	resultBoolean := parser.ExprBoolean{
		Value: isEqual,
	}

	return resultBoolean, nil
}

// evalNotEquals accepts a parser.ExprNotEquals expression and performs the inequality comparison.
func evalNotEquals(expr parser.Expr, input any) (ret parser.Expr, err error) {
	exprNotEquals, ok := expr.(parser.ExprNotEquals)
	if !ok {
		err = fmt.Errorf("failed to assert expression as not equals")
		return
	}

	expr1, err := Eval(exprNotEquals.Expr1, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate first expression: %w", err)
		return
	}

	expr2, err := Eval(exprNotEquals.Expr2, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate second expression: %w", err)
		return
	}

	expr1Type := expr1.Type()
	expr2Type := expr2.Type()
	if expr1Type != expr2Type {
		err = fmt.Errorf("%w: %s and %s", ErrIncompatibleTypes, expr1, expr2)
		return
	}

	switch expr1Type {
	case parser.ExprType_Number:
		return evalNotEqualsNumber(expr1, expr2)
	case parser.ExprType_String:
		return evalNotEqualsString(expr1, expr2)
	case parser.ExprType_Boolean:
		return evalNotEqualsBoolean(expr1, expr2)
	default:
		err = fmt.Errorf("invalid not equals type: %s", expr1)
		return
	}
}

// evalNotEqualsNumber accepts two parser.ExprNumber expressions and compares them for inequality.
func evalNotEqualsNumber(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
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

	isNotEqual := !expr1Number.Value.Equal(expr2Number.Value)

	resultBoolean := parser.ExprBoolean{
		Value: isNotEqual,
	}

	return resultBoolean, nil
}

// evalNotEqualsString accepts two parser.ExprString expressions and compares them for inequality.
func evalNotEqualsString(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
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

	isNotEqual := expr1String.Value != expr2String.Value

	resultBoolean := parser.ExprBoolean{
		Value: isNotEqual,
	}

	return resultBoolean, nil
}

// evalNotEqualsBoolean accepts two parser.ExprBoolean expressions and compares them for inequality.
func evalNotEqualsBoolean(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
	expr1Boolean, ok := expr1.(parser.ExprBoolean)
	if !ok {
		err = fmt.Errorf("failed to assert first expression as boolean")
		return
	}

	expr2Boolean, ok := expr2.(parser.ExprBoolean)
	if !ok {
		err = fmt.Errorf("failed to assert second expression as boolean")
		return
	}

	isNotEqual := expr1Boolean.Value != expr2Boolean.Value

	resultBoolean := parser.ExprBoolean{
		Value: isNotEqual,
	}

	return resultBoolean, nil
}

// evalGreaterThan accepts a parser.ExprGreaterThan expression and performs the greater than comparison.
func evalGreaterThan(expr parser.Expr, input any) (ret parser.Expr, err error) {
	exprGreaterThan, ok := expr.(parser.ExprGreaterThan)
	if !ok {
		err = fmt.Errorf("failed to assert expression as greater than")
		return
	}

	expr1, err := Eval(exprGreaterThan.Expr1, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate first expression: %w", err)
		return
	}

	expr2, err := Eval(exprGreaterThan.Expr2, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate second expression: %w", err)
		return
	}

	expr1Type := expr1.Type()
	expr2Type := expr2.Type()
	if expr1Type != expr2Type {
		err = fmt.Errorf("%w: %s and %s", ErrIncompatibleTypes, expr1, expr2)
		return
	}

	switch expr1Type {
	case parser.ExprType_Number:
		return evalGreaterThanNumber(expr1, expr2)
	case parser.ExprType_String:
		return evalGreaterThanString(expr1, expr2)
	case parser.ExprType_Boolean:
		return evalGreaterThanBoolean(expr1, expr2)
	default:
		err = fmt.Errorf("invalid greater than type: %s", expr1)
		return
	}
}

// evalGreaterThanNumber accepts two parser.ExprNumber expressions and compares them for greater than.
func evalGreaterThanNumber(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
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

	isGreater := expr1Number.Value.GreaterThan(expr2Number.Value)

	resultBoolean := parser.ExprBoolean{
		Value: isGreater,
	}

	return resultBoolean, nil
}

// evalGreaterThanString accepts two parser.ExprString expressions and compares them for greater than.
func evalGreaterThanString(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
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

	isGreater := expr1String.Value > expr2String.Value

	resultBoolean := parser.ExprBoolean{
		Value: isGreater,
	}

	return resultBoolean, nil
}

// evalGreaterThanBoolean accepts two parser.ExprBoolean expressions and compares them for greater than.
func evalGreaterThanBoolean(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
	expr1Boolean, ok := expr1.(parser.ExprBoolean)
	if !ok {
		err = fmt.Errorf("failed to assert first expression as boolean")
		return
	}

	expr2Boolean, ok := expr2.(parser.ExprBoolean)
	if !ok {
		err = fmt.Errorf("failed to assert second expression as boolean")
		return
	}

	isGreater := expr1Boolean.Value && !expr2Boolean.Value

	resultBoolean := parser.ExprBoolean{
		Value: isGreater,
	}

	return resultBoolean, nil
}

// evalLessThan accepts a parser.ExprLessThan expression and performs the less than comparison.
func evalLessThan(expr parser.Expr, input any) (ret parser.Expr, err error) {
	exprLessThan, ok := expr.(parser.ExprLessThan)
	if !ok {
		err = fmt.Errorf("failed to assert expression as less than")
		return
	}

	expr1, err := Eval(exprLessThan.Expr1, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate first expression: %w", err)
		return
	}

	expr2, err := Eval(exprLessThan.Expr2, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate second expression: %w", err)
		return
	}

	expr1Type := expr1.Type()
	expr2Type := expr2.Type()
	if expr1Type != expr2Type {
		err = fmt.Errorf("%w: %s and %s", ErrIncompatibleTypes, expr1, expr2)
		return
	}

	switch expr1Type {
	case parser.ExprType_Number:
		return evalLessThanNumber(expr1, expr2)
	case parser.ExprType_String:
		return evalLessThanString(expr1, expr2)
	case parser.ExprType_Boolean:
		return evalLessThanBoolean(expr1, expr2)
	default:
		err = fmt.Errorf("invalid less than type: %s", expr1)
		return
	}
}

// evalLessThanNumber accepts two parser.ExprNumber expressions and compares them for less than.
func evalLessThanNumber(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
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

	isLess := expr1Number.Value.LessThan(expr2Number.Value)

	resultBoolean := parser.ExprBoolean{
		Value: isLess,
	}

	return resultBoolean, nil
}

// evalLessThanString accepts two parser.ExprString expressions and compares them for less than.
func evalLessThanString(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
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

	isLess := expr1String.Value < expr2String.Value

	resultBoolean := parser.ExprBoolean{
		Value: isLess,
	}

	return resultBoolean, nil
}

// evalLessThanBoolean accepts two parser.ExprBoolean expressions and compares them for less than.
func evalLessThanBoolean(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
	expr1Boolean, ok := expr1.(parser.ExprBoolean)
	if !ok {
		err = fmt.Errorf("failed to assert first expression as boolean")
		return
	}

	expr2Boolean, ok := expr2.(parser.ExprBoolean)
	if !ok {
		err = fmt.Errorf("failed to assert second expression as boolean")
		return
	}

	isLess := !expr1Boolean.Value && expr2Boolean.Value

	resultBoolean := parser.ExprBoolean{
		Value: isLess,
	}

	return resultBoolean, nil
}

// evalLessThanOrEqual accepts a parser.ExprLessThanOrEqual expression and performs the less than or equal comparison.
func evalLessThanOrEqual(expr parser.Expr, input any) (ret parser.Expr, err error) {
	exprLessThanOrEqual, ok := expr.(parser.ExprLessThanOrEqual)
	if !ok {
		err = fmt.Errorf("failed to assert expression as less than or equal")
		return
	}

	expr1, err := Eval(exprLessThanOrEqual.Expr1, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate first expression: %w", err)
		return
	}

	expr2, err := Eval(exprLessThanOrEqual.Expr2, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate second expression: %w", err)
		return
	}

	expr1Type := expr1.Type()
	expr2Type := expr2.Type()
	if expr1Type != expr2Type {
		err = fmt.Errorf("%w: %s and %s", ErrIncompatibleTypes, expr1, expr2)
		return
	}

	switch expr1Type {
	case parser.ExprType_Number:
		return evalLessThanOrEqualNumber(expr1, expr2)
	case parser.ExprType_String:
		return evalLessThanOrEqualString(expr1, expr2)
	case parser.ExprType_Boolean:
		return evalLessThanOrEqualBoolean(expr1, expr2)
	default:
		err = fmt.Errorf("invalid less than or equal type: %s", expr1)
		return
	}
}

// evalLessThanOrEqualNumber accepts two parser.ExprNumber expressions and compares them for less than or equal.
func evalLessThanOrEqualNumber(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
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

	isLessOrEqual := expr1Number.Value.LessThanOrEqual(expr2Number.Value)

	resultBoolean := parser.ExprBoolean{
		Value: isLessOrEqual,
	}

	return resultBoolean, nil
}

// evalLessThanOrEqualString accepts two parser.ExprString expressions and compares them for less than or equal.
func evalLessThanOrEqualString(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
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

	isLessOrEqual := expr1String.Value <= expr2String.Value

	resultBoolean := parser.ExprBoolean{
		Value: isLessOrEqual,
	}

	return resultBoolean, nil
}

// evalLessThanOrEqualBoolean accepts two parser.ExprBoolean expressions and compares them for less than or equal.
func evalLessThanOrEqualBoolean(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
	expr1Boolean, ok := expr1.(parser.ExprBoolean)
	if !ok {
		err = fmt.Errorf("failed to assert first expression as boolean")
		return
	}

	expr2Boolean, ok := expr2.(parser.ExprBoolean)
	if !ok {
		err = fmt.Errorf("failed to assert second expression as boolean")
		return
	}

	isLessOrEqual := !expr1Boolean.Value && expr2Boolean.Value || expr1Boolean.Value == expr2Boolean.Value

	resultBoolean := parser.ExprBoolean{
		Value: isLessOrEqual,
	}

	return resultBoolean, nil
}

// evalGreaterThanOrEqual accepts a parser.ExprGreaterThanOrEqual expression and performs the greater than or equal comparison.
func evalGreaterThanOrEqual(expr parser.Expr, input any) (ret parser.Expr, err error) {
	exprGreaterThanOrEqual, ok := expr.(parser.ExprGreaterThanOrEqual)
	if !ok {
		err = fmt.Errorf("failed to assert expression as greater than or equal")
		return
	}

	expr1, err := Eval(exprGreaterThanOrEqual.Expr1, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate first expression: %w", err)
		return
	}

	expr2, err := Eval(exprGreaterThanOrEqual.Expr2, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate second expression: %w", err)
		return
	}

	expr1Type := expr1.Type()
	expr2Type := expr2.Type()
	if expr1Type != expr2Type {
		err = fmt.Errorf("%w: %s and %s", ErrIncompatibleTypes, expr1, expr2)
		return
	}

	switch expr1Type {
	case parser.ExprType_Number:
		return evalGreaterThanOrEqualNumber(expr1, expr2)
	case parser.ExprType_String:
		return evalGreaterThanOrEqualString(expr1, expr2)
	case parser.ExprType_Boolean:
		return evalGreaterThanOrEqualBoolean(expr1, expr2)
	default:
		err = fmt.Errorf("invalid greater than or equal type: %s", expr1)
		return
	}
}

// evalGreaterThanOrEqualNumber accepts two parser.ExprNumber expressions and compares them for greater than or equal.
func evalGreaterThanOrEqualNumber(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
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

	isGreaterOrEqual := expr1Number.Value.GreaterThanOrEqual(expr2Number.Value)

	resultBoolean := parser.ExprBoolean{
		Value: isGreaterOrEqual,
	}

	return resultBoolean, nil
}

// evalGreaterThanOrEqualString accepts two parser.ExprString expressions and compares them for greater than or equal.
func evalGreaterThanOrEqualString(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
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

	isGreaterOrEqual := expr1String.Value >= expr2String.Value

	resultBoolean := parser.ExprBoolean{
		Value: isGreaterOrEqual,
	}

	return resultBoolean, nil
}

// evalGreaterThanOrEqualBoolean accepts two parser.ExprBoolean expressions and compares them for greater than or equal.
func evalGreaterThanOrEqualBoolean(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
	expr1Boolean, ok := expr1.(parser.ExprBoolean)
	if !ok {
		err = fmt.Errorf("failed to assert first expression as boolean")
		return
	}

	expr2Boolean, ok := expr2.(parser.ExprBoolean)
	if !ok {
		err = fmt.Errorf("failed to assert second expression as boolean")
		return
	}

	isGreaterOrEqual := expr1Boolean.Value && !expr2Boolean.Value || expr1Boolean.Value == expr2Boolean.Value

	resultBoolean := parser.ExprBoolean{
		Value: isGreaterOrEqual,
	}

	return resultBoolean, nil
}

// evalAnd accepts a parser.ExprAnd expression and performs the logical AND operation.
func evalAnd(expr parser.Expr, input any) (ret parser.Expr, err error) {
	exprAnd, ok := expr.(parser.ExprAnd)
	if !ok {
		err = fmt.Errorf("failed to assert expression as and")
		return
	}

	// Evaluate first expression
	expr1, err := Eval(exprAnd.Expr1, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate first expression: %w", err)
		return
	}

	// Short-circuit: if first expression is false, don't evaluate second
	if expr1.Type() == parser.ExprType_Boolean {
		expr1Boolean, ok := expr1.(parser.ExprBoolean)
		if ok && !expr1Boolean.Value {
			return parser.ExprBoolean{Value: false}, nil
		}
	}

	// Evaluate second expression
	expr2, err := Eval(exprAnd.Expr2, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate second expression: %w", err)
		return
	}

	// Both expressions must be boolean
	expr1Type := expr1.Type()
	expr2Type := expr2.Type()
	if expr1Type != parser.ExprType_Boolean || expr2Type != parser.ExprType_Boolean {
		err = fmt.Errorf("%w: got %s and %s", ErrBooleanOperation, expr1, expr2)
		return
	}

	return evalAndBoolean(expr1, expr2)
}

// evalAndBoolean accepts two parser.ExprBoolean expressions and performs logical AND.
func evalAndBoolean(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
	expr1Boolean, ok := expr1.(parser.ExprBoolean)
	if !ok {
		err = fmt.Errorf("failed to assert first expression as boolean")
		return
	}

	expr2Boolean, ok := expr2.(parser.ExprBoolean)
	if !ok {
		err = fmt.Errorf("failed to assert second expression as boolean")
		return
	}

	andResult := expr1Boolean.Value && expr2Boolean.Value

	resultBoolean := parser.ExprBoolean{
		Value: andResult,
	}

	return resultBoolean, nil
}

// evalOr accepts a parser.ExprOr expression and performs the logical OR operation.
func evalOr(expr parser.Expr, input any) (ret parser.Expr, err error) {
	exprOr, ok := expr.(parser.ExprOr)
	if !ok {
		err = fmt.Errorf("failed to assert expression as or")
		return
	}

	// Evaluate first expression
	expr1, err := Eval(exprOr.Expr1, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate first expression: %w", err)
		return
	}

	// Short-circuit: if first expression is true, don't evaluate second
	if expr1.Type() == parser.ExprType_Boolean {
		expr1Boolean, ok := expr1.(parser.ExprBoolean)
		if ok && expr1Boolean.Value {
			return parser.ExprBoolean{Value: true}, nil
		}
	}

	// Evaluate second expression
	expr2, err := Eval(exprOr.Expr2, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate second expression: %w", err)
		return
	}

	// Both expressions must be boolean
	expr1Type := expr1.Type()
	expr2Type := expr2.Type()
	if expr1Type != parser.ExprType_Boolean || expr2Type != parser.ExprType_Boolean {
		err = fmt.Errorf("%w: got %s and %s", ErrBooleanOperation, expr1, expr2)
		return
	}

	return evalOrBoolean(expr1, expr2)
}

// evalOrBoolean accepts two parser.ExprBoolean expressions and performs logical OR.
func evalOrBoolean(expr1, expr2 parser.Expr) (result parser.Expr, err error) {
	expr1Boolean, ok := expr1.(parser.ExprBoolean)
	if !ok {
		err = fmt.Errorf("failed to assert first expression as boolean")
		return
	}

	expr2Boolean, ok := expr2.(parser.ExprBoolean)
	if !ok {
		err = fmt.Errorf("failed to assert second expression as boolean")
		return
	}

	orResult := expr1Boolean.Value || expr2Boolean.Value

	resultBoolean := parser.ExprBoolean{
		Value: orResult,
	}

	return resultBoolean, nil
}

// evalTernary evaluates a ternary conditional expression with short-circuiting.
func evalTernary(expr parser.Expr, input any) (ret parser.Expr, err error) {
	exprTernary, ok := expr.(parser.ExprTernary)
	if !ok {
		err = fmt.Errorf("failed to assert expression as ternary")
		return
	}

	// Evaluate condition first
	conditionExpr, err := Eval(exprTernary.Condition, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate ternary condition: %w", err)
		return
	}

	// Condition must be boolean
	if conditionExpr.Type() != parser.ExprType_Boolean {
		err = fmt.Errorf("%w: ternary condition must be boolean, got %s", ErrBooleanOperation, conditionExpr)
		return
	}

	conditionBoolean, ok := conditionExpr.(parser.ExprBoolean)
	if !ok {
		err = fmt.Errorf("failed to assert condition as boolean")
		return
	}

	// Short-circuit: evaluate only the appropriate branch
	if conditionBoolean.Value {
		// Evaluate true expression
		trueExpr, err := Eval(exprTernary.TrueExpr, input)
		if err != nil {
			err = fmt.Errorf("failed to evaluate ternary true expression: %w", err)
			return nil, err
		}
		return trueExpr, nil
	} else {
		// Evaluate false expression
		falseExpr, err := Eval(exprTernary.FalseExpr, input)
		if err != nil {
			err = fmt.Errorf("failed to evaluate ternary false expression: %w", err)
			return nil, err
		}
		return falseExpr, nil
	}
}
func evalList(expr parser.Expr, input any) (ret parser.Expr, err error) {
	exprList, ok := expr.(parser.ExprList)
	if !ok {
		err = fmt.Errorf("failed to assert expression as list")
		return
	}

	var evaluatedValues []parser.Expr
	for _, valueExpr := range exprList.Values {
		evaluatedValue, err := Eval(valueExpr, input)
		if err != nil {
			err = fmt.Errorf("failed to evaluate list element: %w", err)
			return nil, err
		}
		evaluatedValues = append(evaluatedValues, evaluatedValue)
	}

	return parser.ExprList{
		Values: evaluatedValues,
	}, nil
}

// evalListIndex evaluates a list indexing operation.
func evalListIndex(expr parser.Expr, input any) (ret parser.Expr, err error) {
	exprListIndex, ok := expr.(parser.ExprListIndex)
	if !ok {
		err = fmt.Errorf("failed to assert expression as list index")
		return
	}

	// Evaluate the list expression
	listExpr, err := Eval(exprListIndex.List, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate list expression: %w", err)
		return
	}

	// Check if it's actually a list
	if listExpr.Type() != parser.ExprType_List {
		err = fmt.Errorf("%w: cannot index into non-list expression of type %d", ErrInvalidIndex, listExpr.Type())
		return
	}

	list, ok := listExpr.(parser.ExprList)
	if !ok {
		err = fmt.Errorf("failed to assert expression as list")
		return
	}

	// Evaluate the index expression
	indexExpr, err := Eval(exprListIndex.Index, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate index expression: %w", err)
		return
	}

	// Check if index is a number
	if indexExpr.Type() != parser.ExprType_Number {
		err = fmt.Errorf("%w: index must be a number, got %d", ErrInvalidIndex, indexExpr.Type())
		return
	}

	indexNumber, ok := indexExpr.(parser.ExprNumber)
	if !ok {
		err = fmt.Errorf("failed to assert index expression as number")
		return
	}

	// Convert index to int
	indexFloat, ok := indexNumber.Value.Float64()
	if !ok {
		err = fmt.Errorf("failed to convert index to float64")
		return
	}

	index := int(indexFloat)
	if indexFloat != float64(index) {
		err = fmt.Errorf("%w: index must be an integer, got %f", ErrInvalidIndex, indexFloat)
		return
	}

	// Check bounds
	if index < 0 || index >= len(list.Values) {
		err = fmt.Errorf("%w: index %d is out of bounds for list of length %d", ErrIndexOutOfBounds, index, len(list.Values))
		return
	}

	// Return the element at the index
	return list.Values[index], nil
}

// evalVariable evaluates a variable expression like `_` and returns its value from the input context.
func evalVariable(expr parser.Expr, input any) (ret parser.Expr, err error) {
	exprVariable, ok := expr.(parser.ExprVariable)
	if !ok {
		err = fmt.Errorf("failed to assert expression as variable")
		return
	}

	variableName := exprVariable.Name

	// Handle the special underscore variable used in filter operations
	if variableName == "_" {
		// Convert the input to an expression to return as the value of the variable
		return convertInputToExpr(input)
	}

	// For other variables (if any), return an error since they're not supported yet
	err = fmt.Errorf("undefined variable: %s", variableName)
	return
}

// evalMap evaluates a map expression by evaluating all its key-value pairs.
func evalMap(expr parser.Expr, input any) (ret parser.Expr, err error) {
	exprMap, ok := expr.(parser.ExprMap)
	if !ok {
		err = fmt.Errorf("failed to assert expression as map")
		return
	}

	var evaluatedPairs []parser.ExprMapPair
	for _, pair := range exprMap.Pairs {
		// Evaluate the key expression
		evaluatedKey, err := Eval(pair.Key, input)
		if err != nil {
			err = fmt.Errorf("failed to evaluate map key: %w", err)
			return nil, err
		}

		// Evaluate the value expression
		evaluatedValue, err := Eval(pair.Value, input)
		if err != nil {
			err = fmt.Errorf("failed to evaluate map value: %w", err)
			return nil, err
		}

		evaluatedPairs = append(evaluatedPairs, parser.ExprMapPair{
			Key:   evaluatedKey,
			Value: evaluatedValue,
		})
	}

	return parser.ExprMap{
		Pairs: evaluatedPairs,
	}, nil
}

// evalMapIndex evaluates a map indexing operation.
func evalMapIndex(expr parser.Expr, input any) (ret parser.Expr, err error) {
	exprMapIndex, ok := expr.(parser.ExprMapIndex)
	if !ok {
		err = fmt.Errorf("failed to assert expression as map index")
		return
	}

	// Evaluate the map expression
	mapExpr, err := Eval(exprMapIndex.Map, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate map expression: %w", err)
		return
	}

	// Check if it's actually a map
	if mapExpr.Type() != parser.ExprType_Map {
		err = fmt.Errorf("%w: cannot index into non-map expression of type %d", ErrInvalidMapIndex, mapExpr.Type())
		return
	}

	mapValue, ok := mapExpr.(parser.ExprMap)
	if !ok {
		err = fmt.Errorf("failed to assert expression as map")
		return
	}

	// Evaluate the index expression
	indexExpr, err := Eval(exprMapIndex.Index, input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate index expression: %w", err)
		return
	}

	// Search for the key in the map
	for _, pair := range mapValue.Pairs {
		// Compare keys for equality
		isEqual, err := areExpressionsEqual(pair.Key, indexExpr)
		if err != nil {
			return nil, fmt.Errorf("failed to compare map keys: %w", err)
		}

		if isEqual {
			return pair.Value, nil
		}
	}

	// Key not found
	err = fmt.Errorf("%w: key %v not found in map", ErrKeyNotFound, indexExpr)
	return
}

// areExpressionsEqual checks if two expressions are equal for map key comparison.
func areExpressionsEqual(expr1, expr2 parser.Expr) (bool, error) {
	// If both expressions are of the same type, compare directly
	if expr1.Type() == expr2.Type() {
		switch expr1.Type() {
		case parser.ExprType_String:
			str1, ok1 := expr1.(parser.ExprString)
			str2, ok2 := expr2.(parser.ExprString)
			if !ok1 || !ok2 {
				return false, fmt.Errorf("failed to assert expressions as strings")
			}
			return str1.Value == str2.Value, nil

		case parser.ExprType_Number:
			num1, ok1 := expr1.(parser.ExprNumber)
			num2, ok2 := expr2.(parser.ExprNumber)
			if !ok1 || !ok2 {
				return false, fmt.Errorf("failed to assert expressions as numbers")
			}
			return num1.Value.Equal(num2.Value), nil

		case parser.ExprType_Boolean:
			bool1, ok1 := expr1.(parser.ExprBoolean)
			bool2, ok2 := expr2.(parser.ExprBoolean)
			if !ok1 || !ok2 {
				return false, fmt.Errorf("failed to assert expressions as booleans")
			}
			return bool1.Value == bool2.Value, nil

		default:
			// For other types, we don't support them as map keys
			return false, fmt.Errorf("unsupported map key type: %s", expr1.String())
		}
	}

	// For cross-type comparisons, try to convert and compare
	// Support String <-> Number comparisons for map key flexibility
	if (expr1.Type() == parser.ExprType_String && expr2.Type() == parser.ExprType_Number) ||
		(expr1.Type() == parser.ExprType_Number && expr2.Type() == parser.ExprType_String) {

		// Convert both to string and compare
		str1, err1 := expr1.Decode()
		if err1 != nil {
			return false, fmt.Errorf("failed to decode expression 1: %w", err1)
		}

		str2, err2 := expr2.Decode()
		if err2 != nil {
			return false, fmt.Errorf("failed to decode expression 2: %w", err2)
		}

		// Convert to string and compare
		str1Str, ok1 := str1.(string)
		if !ok1 {
			return false, fmt.Errorf("expression 1 is not a string after decode")
		}

		str2Str, ok2 := str2.(string)
		if !ok2 {
			// If expr2 is a number, convert it to string
			if num2, ok2 := str2.(float64); ok2 {
				str2Str = fmt.Sprintf("%g", num2)
			} else {
				return false, fmt.Errorf("expression 2 is not a string or number after decode")
			}
		}

		return str1Str == str2Str, nil
	}

	// For other cross-type comparisons, they don't match
	return false, nil
}

// evalFunction evaluates a function call expression.
func evalFunction(expr parser.Expr, input any) (ret parser.Expr, err error) {
	exprFunction, ok := expr.(parser.ExprFunction)
	if !ok {
		err = fmt.Errorf("failed to assert expression as function")
		return
	}

	// Look up the function in the registry
	functionFunc, exists := functionRegistry[exprFunction.Name]
	if !exists {
		err = fmt.Errorf("%w: %s", ErrUndefinedFunction, exprFunction.Name)
		return
	}

	// Call the function with the evaluated arguments
	return functionFunc(exprFunction.Args, input)
}

// evalLenFunction implements the len() built-in function.
// Returns the length of strings, lists, and maps.
func evalLenFunction(args []parser.Expr, input any) (ret parser.Expr, err error) {
	if len(args) != 1 {
		err = fmt.Errorf("%w: len() expects exactly 1 argument, got %d", ErrInvalidArgumentCount, len(args))
		return
	}

	// Evaluate the argument
	argExpr, err := Eval(args[0], input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate len() argument: %w", err)
		return
	}

	switch argExpr.Type() {
	case parser.ExprType_String:
		exprString, ok := argExpr.(parser.ExprString)
		if !ok {
			err = fmt.Errorf("failed to assert expression as string")
			return
		}
		return parser.ExprNumber{Value: decimal.NewFromInt(int64(len(exprString.Value)))}, nil

	case parser.ExprType_List:
		exprList, ok := argExpr.(parser.ExprList)
		if !ok {
			err = fmt.Errorf("failed to assert expression as list")
			return
		}
		return parser.ExprNumber{Value: decimal.NewFromInt(int64(len(exprList.Values)))}, nil

	case parser.ExprType_Map:
		exprMap, ok := argExpr.(parser.ExprMap)
		if !ok {
			err = fmt.Errorf("failed to assert expression as map")
			return
		}
		return parser.ExprNumber{Value: decimal.NewFromInt(int64(len(exprMap.Pairs)))}, nil

	case parser.ExprType_Number:
		// For numbers, return error as per ticket specification
		err = fmt.Errorf("%w: len() cannot be applied to numbers", ErrInvalidArgumentType)
		return

	case parser.ExprType_Boolean:
		// For booleans, return error as per ticket specification
		err = fmt.Errorf("%w: len() cannot be applied to booleans", ErrInvalidArgumentType)
		return

	default:
		err = fmt.Errorf("%w: len() cannot be applied to type %s", ErrInvalidArgumentType, argExpr.String())
		return
	}
}

// evalFilterFunction implements the filter() built-in function.
// Filters a list based on a boolean expression using `_` as the element placeholder.
func evalFilterFunction(args []parser.Expr, input any) (ret parser.Expr, err error) {
	if len(args) != 2 {
		err = fmt.Errorf("%w: filter() expects exactly 2 arguments, got %d", ErrInvalidArgumentCount, len(args))
		return
	}

	// Evaluate the first argument (the list to filter)
	listArg, err := Eval(args[0], input)
	if err != nil {
		err = fmt.Errorf("failed to evaluate filter() list argument: %w", err)
		return
	}

	// Check that the first argument is a list
	if listArg.Type() != parser.ExprType_List {
		err = fmt.Errorf("%w: filter() first argument must be a list, got %s", ErrInvalidArgumentType, listArg.String())
		return
	}

	exprList, ok := listArg.(parser.ExprList)
	if !ok {
		err = fmt.Errorf("failed to assert expression as list")
		return
	}

	// The second argument is the filter expression with `_` as placeholder
	filterExpr := args[1]

	// Create a result list to store filtered elements
	var filteredValues []parser.Expr

	// Iterate through each element in the input list
	for _, element := range exprList.Values {
		// Create a custom evaluation context where `_` is treated as the current element
		// We need to create a mechanism to substitute `_` during evaluation
		// In this case, we'll need to have a special eval function that can handle `_` as an identifier
		// For now, let me check if there's an identifier type

		// For the purpose of this implementation, we'll need to create a custom evaluation
		// where the underscore `_` is treated as a special variable containing the current element
		// This could be done by creating a context or by implementing a custom eval that understands `_`

		// For now, let me implement a simple approach by wrapping the element in a map
		// where we could potentially use `$._` for the underscore, but that's not what the requirement says.
		// The requirement is to replace `_` with the element value during evaluation.

		// The cleanest approach would be to implement an expression visitor that can substitute `_` with the element
		// However, for now, let's implement a simplified approach by creating a custom evaluation context
		// based on the input data, where the underscore is treated specially.

		// We can implement this by modifying Eval to recognize a special context where `_` is the current element
		// For this, I need to first identify what `_` would be parsed as.
		// Let me implement a basic version that passes the element as input and handles `_` specially.

		// For this implementation, we need to understand how `_` should be parsed.
		// Since `_` is typically an identifier, I'll need to create a custom evaluation context.

		// The approach here is to evaluate the filter expression in a context where `_` refers to the current element
		// Let me create a custom evaluation function that handles this case
		result, evalErr := evalFilterExpression(filterExpr, element)
		if evalErr != nil {
			err = fmt.Errorf("failed to evaluate filter expression: %w", evalErr)
			return nil, err
		}

		// Check that the result is a boolean
		if result.Type() != parser.ExprType_Boolean {
			err = fmt.Errorf("%w: filter expression must evaluate to a boolean, got %s", ErrInvalidArgumentType, result.String())
			return
		}

		// Add element to result if condition is true
		resultBool, ok := result.(parser.ExprBoolean)
		if !ok {
			err = fmt.Errorf("failed to assert result as boolean")
			return
		}

		if resultBool.Value {
			filteredValues = append(filteredValues, element)
		}
	}

	// Return the filtered list
	return parser.ExprList{Values: filteredValues}, nil
}

// evalFilterExpression evaluates the filter expression with the given element as the value for `_`.
// This function evaluates the expression by using the element as the input context, so that
// when the variable `_` is encountered during evaluation, it returns the element.
func evalFilterExpression(expr parser.Expr, element parser.Expr) (parser.Expr, error) {
	// Convert the element back to its original data type for use as input
	elementData, err := element.Decode()
	if err != nil {
		// If we can't decode the element, use the element expression directly
		elementData = element
	}

	// Evaluate the filter expression with the element as input context
	// This allows the variable `_` to resolve to the current element during evaluation
	return Eval(expr, elementData)
}

// convertInputToExpr converts input data to appropriate expression types.
func convertInputToExpr(input any) (parser.Expr, error) {
	if input == nil {
		return nil, fmt.Errorf("%w: input data cannot be nil", ErrIncompatibleTypes)
	}

	switch v := input.(type) {
	case string:
		return parser.ExprString{Value: v}, nil
	case int:
		return parser.ExprNumber{Value: decimal.NewFromInt(int64(v))}, nil
	case int8:
		return parser.ExprNumber{Value: decimal.NewFromInt(int64(v))}, nil
	case int16:
		return parser.ExprNumber{Value: decimal.NewFromInt(int64(v))}, nil
	case int32:
		return parser.ExprNumber{Value: decimal.NewFromInt(int64(v))}, nil
	case int64:
		return parser.ExprNumber{Value: decimal.NewFromInt(v)}, nil
	case uint:
		return parser.ExprNumber{Value: decimal.NewFromInt(int64(v))}, nil
	case uint8:
		return parser.ExprNumber{Value: decimal.NewFromInt(int64(v))}, nil
	case uint16:
		return parser.ExprNumber{Value: decimal.NewFromInt(int64(v))}, nil
	case uint32:
		return parser.ExprNumber{Value: decimal.NewFromInt(int64(v))}, nil
	case uint64:
		return parser.ExprNumber{Value: decimal.NewFromInt(int64(v))}, nil
	case float32:
		return parser.ExprNumber{Value: decimal.NewFromFloat32(v)}, nil
	case float64:
		return parser.ExprNumber{Value: decimal.NewFromFloat(v)}, nil
	case bool:
		return parser.ExprBoolean{Value: v}, nil
	case []any:
		var values []parser.Expr
		for _, item := range v {
			expr, err := convertInputToExpr(item)
			if err != nil {
				return nil, fmt.Errorf("failed to convert list item: %w", err)
			}
			values = append(values, expr)
		}
		return parser.ExprList{Values: values}, nil
	case []string:
		var values []parser.Expr
		for _, item := range v {
			values = append(values, parser.ExprString{Value: item})
		}
		return parser.ExprList{Values: values}, nil
	case []int:
		var values []parser.Expr
		for _, item := range v {
			values = append(values, parser.ExprNumber{Value: decimal.NewFromInt(int64(item))})
		}
		return parser.ExprList{Values: values}, nil
	case []int64:
		var values []parser.Expr
		for _, item := range v {
			values = append(values, parser.ExprNumber{Value: decimal.NewFromInt(item)})
		}
		return parser.ExprList{Values: values}, nil
	case []float64:
		var values []parser.Expr
		for _, item := range v {
			values = append(values, parser.ExprNumber{Value: decimal.NewFromFloat(item)})
		}
		return parser.ExprList{Values: values}, nil
	case map[string]any:
		var pairs []parser.ExprMapPair
		for key, value := range v {
			valueExpr, err := convertInputToExpr(value)
			if err != nil {
				return nil, fmt.Errorf("failed to convert map value for key %q: %w", key, err)
			}
			pairs = append(pairs, parser.ExprMapPair{
				Key:   parser.ExprString{Value: key},
				Value: valueExpr,
			})
		}
		return parser.ExprMap{Pairs: pairs}, nil
	case map[any]any:
		var pairs []parser.ExprMapPair
		for key, value := range v {
			// Convert key to string
			var keyExpr parser.Expr
			switch k := key.(type) {
			case string:
				keyExpr = parser.ExprString{Value: k}
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
				keyStr := fmt.Sprintf("%v", k)
				keyExpr = parser.ExprString{Value: keyStr}
			default:
				return nil, fmt.Errorf("unsupported map key type: %T", key)
			}

			valueExpr, err := convertInputToExpr(value)
			if err != nil {
				return nil, fmt.Errorf("failed to convert map value for key %v: %w", key, err)
			}
			pairs = append(pairs, parser.ExprMapPair{
				Key:   keyExpr,
				Value: valueExpr,
			})
		}
		return parser.ExprMap{Pairs: pairs}, nil
	default:
		return nil, fmt.Errorf("%w: unsupported input type: %T", ErrIncompatibleTypes, input)
	}
}
