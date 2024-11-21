package parser

import (
	"fmt"

	"github.com/shopspring/decimal"
)

const (
	ExprType_Undefined = iota
	ExprType_Number
)

// Expr represents an evaluable expression.
type Expr interface {
	fmt.Stringer

	Type() int
	Decode() (any, error)
}

func (ExprNumber) Type() int      { return ExprType_Number }
func (ExprNumber) String() string { return "Numer" }

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
