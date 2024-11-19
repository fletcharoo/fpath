package parser

import (
	"github.com/shopspring/decimal"
)

const (
	ExprType_Undefined = iota
	ExprType_Number
)

var ExprTypeString = map[int]string{
	ExprType_Undefined: "Undefined",
	ExprType_Number:    "Number",
}

// Expr represents an evaluable expression.
type Expr interface {
	Type() int
}

func (ExprNumber) Type() int { return ExprType_Number }

// ExprNumber represents a number literal.
type ExprNumber struct {
	Value decimal.Decimal
}
