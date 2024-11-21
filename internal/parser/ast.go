package parser

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

const (
	ExprType_Undefined = iota
	ExprType_Block
	ExprType_Number
	ExprType_Add
	ExprType_Multiply
)

var errInvalidDecode = errors.New("cannot decode expression")

// Expr represents an evaluable expression.
type Expr interface {
	fmt.Stringer

	Type() int
	Decode() (any, error)
}

func (ExprBlock) Type() int    { return ExprType_Block }
func (ExprNumber) Type() int   { return ExprType_Number }
func (ExprAdd) Type() int      { return ExprType_Add }
func (ExprMultiply) Type() int { return ExprType_Multiply }

func (ExprBlock) String() string    { return "Block" }
func (ExprNumber) String() string   { return "Number" }
func (ExprAdd) String() string      { return "Add" }
func (ExprMultiply) String() string { return "Multiply" }

// ExprBlock represents a grouped expression.
type ExprBlock struct {
	Expr Expr
}

func (e ExprBlock) Decode() (result any, err error) {
	err = fmt.Errorf("%w: %s", errInvalidDecode, e)
	return
}

// ExprNumber represents a number literal.
type ExprNumber struct {
	Value decimal.Decimal
}

func (e ExprNumber) Decode() (result any, err error) {
	result, ok := e.Value.Float64()
	if !ok {
		err = fmt.Errorf("failed to decode as float64")
		return
	}

	return result, nil
}

// ExprAdd represents an operation that adds two expressions together.
type ExprAdd struct {
	Expr1 Expr
	Expr2 Expr
}

func (e ExprAdd) Decode() (result any, err error) {
	err = fmt.Errorf("%w: %s", errInvalidDecode, e)
	return
}

// ExprMultiply represents an operation that adds two expressions together.
type ExprMultiply struct {
	Expr1 Expr
	Expr2 Expr
}

func (e ExprMultiply) Decode() (result any, err error) {
	err = fmt.Errorf("%w: %s", errInvalidDecode, e)
	return
}
