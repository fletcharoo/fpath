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
	ExprType_String
	ExprType_Add
	ExprType_Subtract
	ExprType_Multiply
	ExprType_Divide
	ExprType_Equals
	ExprType_NotEquals
	ExprType_GreaterThan
	ExprType_GreaterThanOrEqual
	ExprType_LessThan
	ExprType_LessThanOrEqual
	ExprType_And
	ExprType_Or
	ExprType_Boolean
	ExprType_List
	ExprType_ListIndex
	ExprType_Map
	ExprType_MapIndex
)

var (
	ErrInvalidDecode  = errors.New("cannot decode expression")
	ErrUndefinedToken = errors.New("undefined token")
	ErrExpectedToken  = errors.New("expected token")
)

// Expr represents an evaluable expression.
type Expr interface {
	fmt.Stringer

	Type() int
	Decode() (any, error)
}

func (ExprBlock) Type() int              { return ExprType_Block }
func (ExprNumber) Type() int             { return ExprType_Number }
func (ExprString) Type() int             { return ExprType_String }
func (ExprAdd) Type() int                { return ExprType_Add }
func (ExprSubtract) Type() int           { return ExprType_Subtract }
func (ExprMultiply) Type() int           { return ExprType_Multiply }
func (ExprDivide) Type() int             { return ExprType_Divide }
func (ExprEquals) Type() int             { return ExprType_Equals }
func (ExprNotEquals) Type() int          { return ExprType_NotEquals }
func (ExprGreaterThan) Type() int        { return ExprType_GreaterThan }
func (ExprGreaterThanOrEqual) Type() int { return ExprType_GreaterThanOrEqual }
func (ExprLessThan) Type() int           { return ExprType_LessThan }
func (ExprLessThanOrEqual) Type() int    { return ExprType_LessThanOrEqual }
func (ExprAnd) Type() int                { return ExprType_And }
func (ExprOr) Type() int                 { return ExprType_Or }
func (ExprBoolean) Type() int            { return ExprType_Boolean }
func (ExprList) Type() int               { return ExprType_List }
func (ExprListIndex) Type() int          { return ExprType_ListIndex }
func (ExprMap) Type() int                { return ExprType_Map }
func (ExprMapIndex) Type() int           { return ExprType_MapIndex }

func (ExprBlock) String() string              { return "Block" }
func (ExprNumber) String() string             { return "Number" }
func (ExprString) String() string             { return "String" }
func (ExprAdd) String() string                { return "Add" }
func (ExprSubtract) String() string           { return "Subtract" }
func (ExprMultiply) String() string           { return "Multiply" }
func (ExprDivide) String() string             { return "Divide" }
func (ExprEquals) String() string             { return "Equals" }
func (ExprNotEquals) String() string          { return "NotEquals" }
func (ExprGreaterThan) String() string        { return "GreaterThan" }
func (ExprGreaterThanOrEqual) String() string { return "GreaterThanOrEqual" }
func (ExprLessThan) String() string           { return "LessThan" }
func (ExprLessThanOrEqual) String() string    { return "LessThanOrEqual" }
func (ExprAnd) String() string                { return "And" }
func (ExprOr) String() string                 { return "Or" }
func (ExprBoolean) String() string            { return "Boolean" }
func (ExprList) String() string               { return "List" }
func (ExprListIndex) String() string          { return "ListIndex" }
func (ExprMap) String() string                { return "Map" }
func (ExprMapIndex) String() string           { return "MapIndex" }

// ExprBlock represents a grouped expression.
type ExprBlock struct {
	Expr Expr
}

func (e ExprBlock) Decode() (result any, err error) {
	err = fmt.Errorf("%w: %s", ErrInvalidDecode, e)
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

// ExprString represents a string literal.
type ExprString struct {
	Value string
}

func (e ExprString) Decode() (result any, err error) {
	result = e.Value
	return result, nil
}

// ExprAdd represents an operation that adds two expressions together.
type ExprAdd struct {
	Expr1 Expr
	Expr2 Expr
}

func (e ExprAdd) Decode() (result any, err error) {
	err = fmt.Errorf("%w: %s", ErrInvalidDecode, e)
	return
}

// ExprSubtract represents an operation that subtracts two expressions.
type ExprSubtract struct {
	Expr1 Expr
	Expr2 Expr
}

func (e ExprSubtract) Decode() (result any, err error) {
	err = fmt.Errorf("%w: %s", ErrInvalidDecode, e)
	return
}

// ExprMultiply represents an operation that adds two expressions together.
type ExprMultiply struct {
	Expr1 Expr
	Expr2 Expr
}

func (e ExprMultiply) Decode() (result any, err error) {
	err = fmt.Errorf("%w: %s", ErrInvalidDecode, e)
	return
}

// ExprDivide represents an operation that divides two expressions.
type ExprDivide struct {
	Expr1 Expr
	Expr2 Expr
}

func (e ExprDivide) Decode() (result any, err error) {
	err = fmt.Errorf("%w: %s", ErrInvalidDecode, e)
	return
}

// ExprEquals represents an operation that checks equality between two expressions.
type ExprEquals struct {
	Expr1 Expr
	Expr2 Expr
}

func (e ExprEquals) Decode() (result any, err error) {
	err = fmt.Errorf("%w: %s", ErrInvalidDecode, e)
	return
}

// ExprNotEquals represents an operation that checks inequality between two expressions.
type ExprNotEquals struct {
	Expr1 Expr
	Expr2 Expr
}

func (e ExprNotEquals) Decode() (result any, err error) {
	err = fmt.Errorf("%w: %s", ErrInvalidDecode, e)
	return
}

// ExprGreaterThan represents an operation that checks if one expression is greater than another.
type ExprGreaterThan struct {
	Expr1 Expr
	Expr2 Expr
}

func (e ExprGreaterThan) Decode() (result any, err error) {
	err = fmt.Errorf("%w: %s", ErrInvalidDecode, e)
	return
}

// ExprGreaterThanOrEqual represents an operation that checks if one expression is greater than or equal to another.
type ExprGreaterThanOrEqual struct {
	Expr1 Expr
	Expr2 Expr
}

func (e ExprGreaterThanOrEqual) Decode() (result any, err error) {
	err = fmt.Errorf("%w: %s", ErrInvalidDecode, e)
	return
}

// ExprLessThan represents an operation that checks if one expression is less than another.
type ExprLessThan struct {
	Expr1 Expr
	Expr2 Expr
}

func (e ExprLessThan) Decode() (result any, err error) {
	err = fmt.Errorf("%w: %s", ErrInvalidDecode, e)
	return
}

// ExprLessThanOrEqual represents an operation that checks if one expression is less than or equal to another.
type ExprLessThanOrEqual struct {
	Expr1 Expr
	Expr2 Expr
}

func (e ExprLessThanOrEqual) Decode() (result any, err error) {
	err = fmt.Errorf("%w: %s", ErrInvalidDecode, e)
	return
}

// ExprAnd represents a logical AND operation between two expressions.
type ExprAnd struct {
	Expr1 Expr
	Expr2 Expr
}

func (e ExprAnd) Decode() (result any, err error) {
	err = fmt.Errorf("%w: %s", ErrInvalidDecode, e)
	return
}

// ExprOr represents a logical OR operation between two expressions.
type ExprOr struct {
	Expr1 Expr
	Expr2 Expr
}

func (e ExprOr) Decode() (result any, err error) {
	err = fmt.Errorf("%w: %s", ErrInvalidDecode, e)
	return
}

// ExprBoolean represents a boolean literal.
type ExprBoolean struct {
	Value bool
}

func (e ExprBoolean) Decode() (result any, err error) {
	result = e.Value
	return result, nil
}

// ExprList represents a list literal containing zero or more expressions.
type ExprList struct {
	Values []Expr
}

func (e ExprList) Decode() (result any, err error) {
	result = e
	return result, nil
}

// ExprListIndex represents an indexing operation into a list expression.
type ExprListIndex struct {
	List  Expr
	Index Expr
}

func (e ExprListIndex) Decode() (result any, err error) {
	err = fmt.Errorf("%w: %s", ErrInvalidDecode, e)
	return
}

// ExprMap represents a map literal containing zero or more key-value pairs.
type ExprMap struct {
	Pairs []ExprMapPair
}

func (e ExprMap) Decode() (result any, err error) {
	result = e
	return result, nil
}

// ExprMapPair represents a key-value pair in a map.
type ExprMapPair struct {
	Key   Expr
	Value Expr
}

// ExprMapIndex represents an indexing operation into a map expression.
type ExprMapIndex struct {
	Map   Expr
	Index Expr
}

func (e ExprMapIndex) Decode() (result any, err error) {
	err = fmt.Errorf("%w: %s", ErrInvalidDecode, e)
	return
}
