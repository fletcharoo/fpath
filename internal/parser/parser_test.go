package parser

import (
	"errors"
	"testing"

	"github.com/fletcharoo/fpath/internal/lexer"
	"github.com/shopspring/decimal"
)

func Test_New(t *testing.T) {
	input := "test input"
	lex := lexer.New(input)
	parser := New(lex)

	if parser.lexer != lex {
		t.Fatalf("Expected lexer to be set")
	}
}

func Test_Parser_Parse(t *testing.T) {
	testCases := map[string]struct {
		input    string
		validate func(Expr, error)
	}{
		"Number": {
			input: "123",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Number {
					t.Fatalf("Expected Number type, got %d", expr.Type())
				}
				num, ok := expr.(ExprNumber)
				if !ok {
					t.Fatalf("Expected ExprNumber, got %T", expr)
				}
				expected, _ := decimal.NewFromString("123")
				if !num.Value.Equal(expected) {
					t.Fatalf("Expected 123, got %s", num.Value.String())
				}
			},
		},
		"String": {
			input: `"hello"`,
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_String {
					t.Fatalf("Expected String type, got %d", expr.Type())
				}
				str, ok := expr.(ExprString)
				if !ok {
					t.Fatalf("Expected ExprString, got %T", expr)
				}
				if str.Value != "hello" {
					t.Fatalf("Expected hello, got %s", str.Value)
				}
			},
		},
		"Boolean true": {
			input: "true",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Boolean {
					t.Fatalf("Expected Boolean type, got %d", expr.Type())
				}
				boolean, ok := expr.(ExprBoolean)
				if !ok {
					t.Fatalf("Expected ExprBoolean, got %T", expr)
				}
				if !boolean.Value {
					t.Fatalf("Expected true, got false")
				}
			},
		},
		"Boolean false": {
			input: "false",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Boolean {
					t.Fatalf("Expected Boolean type, got %d", expr.Type())
				}
				boolean, ok := expr.(ExprBoolean)
				if !ok {
					t.Fatalf("Expected ExprBoolean, got %T", expr)
				}
				if boolean.Value {
					t.Fatalf("Expected false, got true")
				}
			},
		},
		"Input": {
			input: "$",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Input {
					t.Fatalf("Expected Input type, got %d", expr.Type())
				}
				_, ok := expr.(ExprInput)
				if !ok {
					t.Fatalf("Expected ExprInput, got %T", expr)
				}
			},
		},
		"Block": {
			input: "(123)",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Block {
					t.Fatalf("Expected Block type, got %d", expr.Type())
				}
				block, ok := expr.(ExprBlock)
				if !ok {
					t.Fatalf("Expected ExprBlock, got %T", expr)
				}
				if block.Expr.Type() != ExprType_Number {
					t.Fatalf("Expected Number inside block, got %d", block.Expr.Type())
				}
			},
		},
		"Add operation": {
			input: "123+456",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Add {
					t.Fatalf("Expected Add type, got %d", expr.Type())
				}
				add, ok := expr.(ExprAdd)
				if !ok {
					t.Fatalf("Expected ExprAdd, got %T", expr)
				}
				if add.Expr1.Type() != ExprType_Number {
					t.Fatalf("Expected Number as first operand, got %d", add.Expr1.Type())
				}
				if add.Expr2.Type() != ExprType_Number {
					t.Fatalf("Expected Number as second operand, got %d", add.Expr2.Type())
				}
			},
		},
		"Subtract operation": {
			input: "123-456",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Subtract {
					t.Fatalf("Expected Subtract type, got %d", expr.Type())
				}
				sub, ok := expr.(ExprSubtract)
				if !ok {
					t.Fatalf("Expected ExprSubtract, got %T", expr)
				}
				if sub.Expr1.Type() != ExprType_Number {
					t.Fatalf("Expected Number as first operand, got %d", sub.Expr1.Type())
				}
				if sub.Expr2.Type() != ExprType_Number {
					t.Fatalf("Expected Number as second operand, got %d", sub.Expr2.Type())
				}
			},
		},
		"Multiply operation": {
			input: "123*456",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Multiply {
					t.Fatalf("Expected Multiply type, got %d", expr.Type())
				}
				mul, ok := expr.(ExprMultiply)
				if !ok {
					t.Fatalf("Expected ExprMultiply, got %T", expr)
				}
				if mul.Expr1.Type() != ExprType_Number {
					t.Fatalf("Expected Number as first operand, got %d", mul.Expr1.Type())
				}
				if mul.Expr2.Type() != ExprType_Number {
					t.Fatalf("Expected Number as second operand, got %d", mul.Expr2.Type())
				}
			},
		},
		"Divide operation": {
			input: "123/456",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Divide {
					t.Fatalf("Expected Divide type, got %d", expr.Type())
				}
				div, ok := expr.(ExprDivide)
				if !ok {
					t.Fatalf("Expected ExprDivide, got %T", expr)
				}
				if div.Expr1.Type() != ExprType_Number {
					t.Fatalf("Expected Number as first operand, got %d", div.Expr1.Type())
				}
				if div.Expr2.Type() != ExprType_Number {
					t.Fatalf("Expected Number as second operand, got %d", div.Expr2.Type())
				}
			},
		},
		"Modulo operation": {
			input: "123%456",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Modulo {
					t.Fatalf("Expected Modulo type, got %d", expr.Type())
				}
				mod, ok := expr.(ExprModulo)
				if !ok {
					t.Fatalf("Expected ExprModulo, got %T", expr)
				}
				if mod.Expr1.Type() != ExprType_Number {
					t.Fatalf("Expected Number as first operand, got %d", mod.Expr1.Type())
				}
				if mod.Expr2.Type() != ExprType_Number {
					t.Fatalf("Expected Number as second operand, got %d", mod.Expr2.Type())
				}
			},
		},
		"IntegerDivision operation": {
			input: "123//456",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_IntegerDivision {
					t.Fatalf("Expected IntegerDivision type, got %d", expr.Type())
				}
				intDiv, ok := expr.(ExprIntegerDivision)
				if !ok {
					t.Fatalf("Expected ExprIntegerDivision, got %T", expr)
				}
				if intDiv.Expr1.Type() != ExprType_Number {
					t.Fatalf("Expected Number as first operand, got %d", intDiv.Expr1.Type())
				}
				if intDiv.Expr2.Type() != ExprType_Number {
					t.Fatalf("Expected Number as second operand, got %d", intDiv.Expr2.Type())
				}
			},
		},
		"Equals operation": {
			input: "123==456",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Equals {
					t.Fatalf("Expected Equals type, got %d", expr.Type())
				}
				eq, ok := expr.(ExprEquals)
				if !ok {
					t.Fatalf("Expected ExprEquals, got %T", expr)
				}
				if eq.Expr1.Type() != ExprType_Number {
					t.Fatalf("Expected Number as first operand, got %d", eq.Expr1.Type())
				}
				if eq.Expr2.Type() != ExprType_Number {
					t.Fatalf("Expected Number as second operand, got %d", eq.Expr2.Type())
				}
			},
		},
		"NotEquals operation": {
			input: "123!=456",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_NotEquals {
					t.Fatalf("Expected NotEquals type, got %d", expr.Type())
				}
				ne, ok := expr.(ExprNotEquals)
				if !ok {
					t.Fatalf("Expected ExprNotEquals, got %T", expr)
				}
				if ne.Expr1.Type() != ExprType_Number {
					t.Fatalf("Expected Number as first operand, got %d", ne.Expr1.Type())
				}
				if ne.Expr2.Type() != ExprType_Number {
					t.Fatalf("Expected Number as second operand, got %d", ne.Expr2.Type())
				}
			},
		},
		"GreaterThan operation": {
			input: "123>456",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_GreaterThan {
					t.Fatalf("Expected GreaterThan type, got %d", expr.Type())
				}
				gt, ok := expr.(ExprGreaterThan)
				if !ok {
					t.Fatalf("Expected ExprGreaterThan, got %T", expr)
				}
				if gt.Expr1.Type() != ExprType_Number {
					t.Fatalf("Expected Number as first operand, got %d", gt.Expr1.Type())
				}
				if gt.Expr2.Type() != ExprType_Number {
					t.Fatalf("Expected Number as second operand, got %d", gt.Expr2.Type())
				}
			},
		},
		"GreaterThanOrEqual operation": {
			input: "123>=456",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_GreaterThanOrEqual {
					t.Fatalf("Expected GreaterThanOrEqual type, got %d", expr.Type())
				}
				gte, ok := expr.(ExprGreaterThanOrEqual)
				if !ok {
					t.Fatalf("Expected ExprGreaterThanOrEqual, got %T", expr)
				}
				if gte.Expr1.Type() != ExprType_Number {
					t.Fatalf("Expected Number as first operand, got %d", gte.Expr1.Type())
				}
				if gte.Expr2.Type() != ExprType_Number {
					t.Fatalf("Expected Number as second operand, got %d", gte.Expr2.Type())
				}
			},
		},
		"LessThan operation": {
			input: "123<456",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_LessThan {
					t.Fatalf("Expected LessThan type, got %d", expr.Type())
				}
				lt, ok := expr.(ExprLessThan)
				if !ok {
					t.Fatalf("Expected ExprLessThan, got %T", expr)
				}
				if lt.Expr1.Type() != ExprType_Number {
					t.Fatalf("Expected Number as first operand, got %d", lt.Expr1.Type())
				}
				if lt.Expr2.Type() != ExprType_Number {
					t.Fatalf("Expected Number as second operand, got %d", lt.Expr2.Type())
				}
			},
		},
		"LessThanOrEqual operation": {
			input: "123<=456",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_LessThanOrEqual {
					t.Fatalf("Expected LessThanOrEqual type, got %d", expr.Type())
				}
				lte, ok := expr.(ExprLessThanOrEqual)
				if !ok {
					t.Fatalf("Expected ExprLessThanOrEqual, got %T", expr)
				}
				if lte.Expr1.Type() != ExprType_Number {
					t.Fatalf("Expected Number as first operand, got %d", lte.Expr1.Type())
				}
				if lte.Expr2.Type() != ExprType_Number {
					t.Fatalf("Expected Number as second operand, got %d", lte.Expr2.Type())
				}
			},
		},
		"And operation": {
			input: "true&&false",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_And {
					t.Fatalf("Expected And type, got %d", expr.Type())
				}
				and, ok := expr.(ExprAnd)
				if !ok {
					t.Fatalf("Expected ExprAnd, got %T", expr)
				}
				if and.Expr1.Type() != ExprType_Boolean {
					t.Fatalf("Expected Boolean as first operand, got %d", and.Expr1.Type())
				}
				if and.Expr2.Type() != ExprType_Boolean {
					t.Fatalf("Expected Boolean as second operand, got %d", and.Expr2.Type())
				}
			},
		},
		"And operation with parentheses": {
			input: "true&&false",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_And {
					t.Fatalf("Expected And type, got %d", expr.Type())
				}
				and, ok := expr.(ExprAnd)
				if !ok {
					t.Fatalf("Expected ExprAnd, got %T", expr)
				}
				if and.Expr1.Type() != ExprType_Boolean {
					t.Fatalf("Expected Boolean as first operand, got %d", and.Expr1.Type())
				}
				if and.Expr2.Type() != ExprType_Boolean {
					t.Fatalf("Expected Boolean as second operand, got %d", and.Expr2.Type())
				}
			},
		},
		"Empty list": {
			input: "[]",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_List {
					t.Fatalf("Expected List type, got %d", expr.Type())
				}
				list, ok := expr.(ExprList)
				if !ok {
					t.Fatalf("Expected ExprList, got %T", expr)
				}
				if len(list.Values) != 0 {
					t.Fatalf("Expected empty list, got %d elements", len(list.Values))
				}
			},
		},
		"List with numbers": {
			input: "[1, 2, 3]",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_List {
					t.Fatalf("Expected List type, got %d", expr.Type())
				}
				list, ok := expr.(ExprList)
				if !ok {
					t.Fatalf("Expected ExprList, got %T", expr)
				}
				if len(list.Values) != 3 {
					t.Fatalf("Expected 3 elements, got %d", len(list.Values))
				}
				for i, value := range list.Values {
					if value.Type() != ExprType_Number {
						t.Fatalf("Expected Number at index %d, got %d", i, value.Type())
					}
				}
			},
		},
		"List with mixed types": {
			input: "[1, true, \"hello\"]",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_List {
					t.Fatalf("Expected List type, got %d", expr.Type())
				}
				list, ok := expr.(ExprList)
				if !ok {
					t.Fatalf("Expected ExprList, got %T", expr)
				}
				if len(list.Values) != 3 {
					t.Fatalf("Expected 3 elements, got %d", len(list.Values))
				}
				if list.Values[0].Type() != ExprType_Number {
					t.Fatalf("Expected Number at index 0, got %d", list.Values[0].Type())
				}
				if list.Values[1].Type() != ExprType_Boolean {
					t.Fatalf("Expected Boolean at index 1, got %d", list.Values[1].Type())
				}
				if list.Values[2].Type() != ExprType_String {
					t.Fatalf("Expected String at index 2, got %d", list.Values[2].Type())
				}
			},
		},
		"List indexing": {
			input: "[1, 2, 3][0]",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_ListIndex {
					t.Fatalf("Expected ListIndex type, got %d", expr.Type())
				}
				index, ok := expr.(ExprListIndex)
				if !ok {
					t.Fatalf("Expected ExprListIndex, got %T", expr)
				}
				if index.List.Type() != ExprType_List {
					t.Fatalf("Expected List as list operand, got %d", index.List.Type())
				}
				if index.Index.Type() != ExprType_Number {
					t.Fatalf("Expected Number as index operand, got %d", index.Index.Type())
				}
			},
		},
		"Function call with no arguments": {
			input: "len()",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Function {
					t.Fatalf("Expected Function type, got %d", expr.Type())
				}
				function, ok := expr.(ExprFunction)
				if !ok {
					t.Fatalf("Expected ExprFunction, got %T", expr)
				}
				if function.Name != "len" {
					t.Fatalf("Expected function name 'len', got %s", function.Name)
				}
				if len(function.Args) != 0 {
					t.Fatalf("Expected 0 arguments, got %d", len(function.Args))
				}
			},
		},
		"Function call with one argument": {
			input: "len(\"hello\")",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Function {
					t.Fatalf("Expected Function type, got %d", expr.Type())
				}
				function, ok := expr.(ExprFunction)
				if !ok {
					t.Fatalf("Expected ExprFunction, got %T", expr)
				}
				if function.Name != "len" {
					t.Fatalf("Expected function name 'len', got %s", function.Name)
				}
				if len(function.Args) != 1 {
					t.Fatalf("Expected 1 argument, got %d", len(function.Args))
				}
				if function.Args[0].Type() != ExprType_String {
					t.Fatalf("Expected String argument, got %d", function.Args[0].Type())
				}
			},
		},
		"Function call with multiple arguments": {
			input: "len(\"hello\", \"world\")",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Function {
					t.Fatalf("Expected Function type, got %d", expr.Type())
				}
				function, ok := expr.(ExprFunction)
				if !ok {
					t.Fatalf("Expected ExprFunction, got %T", expr)
				}
				if function.Name != "len" {
					t.Fatalf("Expected function name 'len', got %s", function.Name)
				}
				if len(function.Args) != 2 {
					t.Fatalf("Expected 2 arguments, got %d", len(function.Args))
				}
				if function.Args[0].Type() != ExprType_String {
					t.Fatalf("Expected String argument at index 0, got %d", function.Args[0].Type())
				}
				if function.Args[1].Type() != ExprType_String {
					t.Fatalf("Expected String argument at index 1, got %d", function.Args[1].Type())
				}
			},
		},
		"Function call with complex argument": {
			input: "len([1, 2, 3])",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Function {
					t.Fatalf("Expected Function type, got %d", expr.Type())
				}
				function, ok := expr.(ExprFunction)
				if !ok {
					t.Fatalf("Expected ExprFunction, got %T", expr)
				}
				if function.Name != "len" {
					t.Fatalf("Expected function name 'len', got %s", function.Name)
				}
				if len(function.Args) != 1 {
					t.Fatalf("Expected 1 argument, got %d", len(function.Args))
				}
				if function.Args[0].Type() != ExprType_List {
					t.Fatalf("Expected List argument, got %d", function.Args[0].Type())
				}
			},
		},
		"Ternary true branch": {
			input: `true ? "yes" : "no"`,
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Ternary {
					t.Fatalf("Expected Ternary type, got %d", expr.Type())
				}
				ternary, ok := expr.(ExprTernary)
				if !ok {
					t.Fatalf("Expected ExprTernary, got %T", expr)
				}
				if ternary.Condition.Type() != ExprType_Boolean {
					t.Fatalf("Expected Boolean condition, got %d", ternary.Condition.Type())
				}
				if ternary.TrueExpr.Type() != ExprType_String {
					t.Fatalf("Expected String true expression, got %d", ternary.TrueExpr.Type())
				}
				if ternary.FalseExpr.Type() != ExprType_String {
					t.Fatalf("Expected String false expression, got %d", ternary.FalseExpr.Type())
				}
			},
		},
		"Ternary with comparison": {
			input: `5 > 3 ? "greater" : "less"`,
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Ternary {
					t.Fatalf("Expected Ternary type, got %d", expr.Type())
				}
				ternary, ok := expr.(ExprTernary)
				if !ok {
					t.Fatalf("Expected ExprTernary, got %T", expr)
				}
				if ternary.Condition.Type() != ExprType_GreaterThan {
					t.Fatalf("Expected GreaterThan condition, got %d", ternary.Condition.Type())
				}
				if ternary.TrueExpr.Type() != ExprType_String {
					t.Fatalf("Expected String true expression, got %d", ternary.TrueExpr.Type())
				}
				if ternary.FalseExpr.Type() != ExprType_String {
					t.Fatalf("Expected String false expression, got %d", ternary.FalseExpr.Type())
				}
			},
		},
		"Ternary false branch": {
			input: `false ? "yes" : "no"`,
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Ternary {
					t.Fatalf("Expected Ternary type, got %d", expr.Type())
				}
				ternary, ok := expr.(ExprTernary)
				if !ok {
					t.Fatalf("Expected ExprTernary, got %T", expr)
				}
				if ternary.Condition.Type() != ExprType_Boolean {
					t.Fatalf("Expected Boolean condition, got %d", ternary.Condition.Type())
				}
				if ternary.TrueExpr.Type() != ExprType_String {
					t.Fatalf("Expected String true expression, got %d", ternary.TrueExpr.Type())
				}
				if ternary.FalseExpr.Type() != ExprType_String {
					t.Fatalf("Expected String false expression, got %d", ternary.FalseExpr.Type())
				}
			},
		},
		"Nested ternary": {
			input: `true ? (false ? "a" : "b") : "c"`,
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Ternary {
					t.Fatalf("Expected Ternary type, got %d", expr.Type())
				}
				ternary, ok := expr.(ExprTernary)
				if !ok {
					t.Fatalf("Expected ExprTernary, got %T", expr)
				}
				if ternary.Condition.Type() != ExprType_Boolean {
					t.Fatalf("Expected Boolean condition, got %d", ternary.Condition.Type())
				}
				if ternary.TrueExpr.Type() != ExprType_Block {
					t.Fatalf("Expected Block true expression, got %d", ternary.TrueExpr.Type())
				}
				if ternary.FalseExpr.Type() != ExprType_String {
					t.Fatalf("Expected String false expression, got %d", ternary.FalseExpr.Type())
				}
			},
		},
		"Right-associative ternary": {
			input: `true ? "yes" : false ? "no" : "maybe"`,
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Ternary {
					t.Fatalf("Expected Ternary type, got %d", expr.Type())
				}
				ternary, ok := expr.(ExprTernary)
				if !ok {
					t.Fatalf("Expected ExprTernary, got %T", expr)
				}
				// Should be parsed as true ? "yes" : (false ? "no" : "maybe")
				if ternary.Condition.Type() != ExprType_Boolean {
					t.Fatalf("Expected Boolean condition, got %d", ternary.Condition.Type())
				}
				if ternary.TrueExpr.Type() != ExprType_String {
					t.Fatalf("Expected String true expression, got %d", ternary.TrueExpr.Type())
				}
				if ternary.FalseExpr.Type() != ExprType_Ternary {
					t.Fatalf("Expected Ternary false expression (right-associative), got %d", ternary.FalseExpr.Type())
				}
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.input)
			parser := New(lex)
			expr, err := parser.Parse()
			tc.validate(expr, err)
		})
	}
}

func Test_Parser_Parse_Function_Errors(t *testing.T) {
	testCases := map[string]struct {
		input    string
		validate func(Expr, error)
	}{
		"Missing right parenthesis": {
			input: "len(",
			validate: func(expr Expr, err error) {
				if err == nil {
					t.Fatalf("Expected error, but got none")
				}
				// The error could be ErrExpectedToken or could be wrapped in a parsing error
				if !errors.Is(err, ErrExpectedToken) && !errors.Is(err, ErrUndefinedToken) {
					t.Fatalf("Expected ErrExpectedToken or ErrUndefinedToken, got %s", err)
				}
			},
		},
		"Invalid argument syntax": {
			input: "len(,)",
			validate: func(expr Expr, err error) {
				if err == nil {
					t.Fatalf("Expected error, but got none")
				}
				// Should fail when trying to parse the first argument as an expression
				if err == nil {
					t.Fatalf("Expected parsing error")
				}
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.input)
			parser := New(lex)
			expr, err := parser.Parse()
			tc.validate(expr, err)
		})
	}
}

func Test_Parser_Parse_Map(t *testing.T) {
	testCases := map[string]struct {
		input    string
		validate func(Expr, error)
	}{
		"Empty map": {
			input: "{}",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Map {
					t.Fatalf("Expected Map type, got %d", expr.Type())
				}
				mapExpr, ok := expr.(ExprMap)
				if !ok {
					t.Fatalf("Expected ExprMap, got %T", expr)
				}
				if len(mapExpr.Pairs) != 0 {
					t.Fatalf("Expected 0 pairs, got %d", len(mapExpr.Pairs))
				}
			},
		},
		"Single pair": {
			input: `{"key": "value"}`,
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Map {
					t.Fatalf("Expected Map type, got %d", expr.Type())
				}
				mapExpr, ok := expr.(ExprMap)
				if !ok {
					t.Fatalf("Expected ExprMap, got %T", expr)
				}
				if len(mapExpr.Pairs) != 1 {
					t.Fatalf("Expected 1 pair, got %d", len(mapExpr.Pairs))
				}
				// Check key
				if mapExpr.Pairs[0].Key.Type() != ExprType_String {
					t.Fatalf("Expected String key, got %d", mapExpr.Pairs[0].Key.Type())
				}
				keyStr, ok := mapExpr.Pairs[0].Key.(ExprString)
				if !ok || keyStr.Value != "key" {
					t.Fatalf("Expected key 'key', got %v", mapExpr.Pairs[0].Key)
				}
				// Check value
				if mapExpr.Pairs[0].Value.Type() != ExprType_String {
					t.Fatalf("Expected String value, got %d", mapExpr.Pairs[0].Value.Type())
				}
				valueStr, ok := mapExpr.Pairs[0].Value.(ExprString)
				if !ok || valueStr.Value != "value" {
					t.Fatalf("Expected value 'value', got %v", mapExpr.Pairs[0].Value)
				}
			},
		},
		"Multiple pairs": {
			input: `{"name": "Andrew", "age": 30}`,
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Map {
					t.Fatalf("Expected Map type, got %d", expr.Type())
				}
				mapExpr, ok := expr.(ExprMap)
				if !ok {
					t.Fatalf("Expected ExprMap, got %T", expr)
				}
				if len(mapExpr.Pairs) != 2 {
					t.Fatalf("Expected 2 pairs, got %d", len(mapExpr.Pairs))
				}
			},
		},
		"Mixed types": {
			input: `{"name": "Andrew", 1: true, "active": false}`,
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Map {
					t.Fatalf("Expected Map type, got %d", expr.Type())
				}
				mapExpr, ok := expr.(ExprMap)
				if !ok {
					t.Fatalf("Expected ExprMap, got %T", expr)
				}
				if len(mapExpr.Pairs) != 3 {
					t.Fatalf("Expected 3 pairs, got %d", len(mapExpr.Pairs))
				}
			},
		},
		"Nested map": {
			input: `{"outer": {"inner": "value"}}`,
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_Map {
					t.Fatalf("Expected Map type, got %d", expr.Type())
				}
				mapExpr, ok := expr.(ExprMap)
				if !ok {
					t.Fatalf("Expected ExprMap, got %T", expr)
				}
				if len(mapExpr.Pairs) != 1 {
					t.Fatalf("Expected 1 pair, got %d", len(mapExpr.Pairs))
				}
				// Check that the value is a nested map
				if mapExpr.Pairs[0].Value.Type() != ExprType_Map {
					t.Fatalf("Expected nested Map value, got %d", mapExpr.Pairs[0].Value.Type())
				}
			},
		},
		"Invalid syntax - missing colon": {
			input: `{"key" "value"}`,
			validate: func(expr Expr, err error) {
				if err == nil {
					t.Fatalf("Expected error, but got none")
				}
				if !errors.Is(err, ErrExpectedToken) {
					t.Fatalf("Expected ErrExpectedToken, got %s", err)
				}
			},
		},
		"Invalid syntax - missing right brace": {
			input: `{"key": "value"`,
			validate: func(expr Expr, err error) {
				if err == nil {
					t.Fatalf("Expected error, but got none")
				}
				if !errors.Is(err, ErrExpectedToken) {
					t.Fatalf("Expected ErrExpectedToken, got %s", err)
				}
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.input)
			parser := New(lex)
			expr, err := parser.Parse()
			tc.validate(expr, err)
		})
	}
}

func Test_Parser_Parse_MapIndex(t *testing.T) {
	testCases := map[string]struct {
		input    string
		validate func(Expr, error)
	}{
		"String key indexing": {
			input: `{"name": "Andrew"}["name"]`,
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_MapIndex {
					t.Fatalf("Expected MapIndex type, got %d", expr.Type())
				}
				mapIndex, ok := expr.(ExprMapIndex)
				if !ok {
					t.Fatalf("Expected ExprMapIndex, got %T", expr)
				}
				// Check that the map expression is correct
				if mapIndex.Map.Type() != ExprType_Map {
					t.Fatalf("Expected Map expression, got %d", mapIndex.Map.Type())
				}
				// Check that the index expression is correct
				if mapIndex.Index.Type() != ExprType_String {
					t.Fatalf("Expected String index, got %d", mapIndex.Index.Type())
				}
			},
		},
		"Number key indexing": {
			input: `{1: "value"}[1]`,
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_MapIndex {
					t.Fatalf("Expected MapIndex type, got %d", expr.Type())
				}
				mapIndex, ok := expr.(ExprMapIndex)
				if !ok {
					t.Fatalf("Expected ExprMapIndex, got %T", expr)
				}
				// Check that the index expression is a number
				if mapIndex.Index.Type() != ExprType_Number {
					t.Fatalf("Expected Number index, got %d", mapIndex.Index.Type())
				}
			},
		},
		"Boolean key indexing": {
			input: `{true: "value"}[true]`,
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_MapIndex {
					t.Fatalf("Expected MapIndex type, got %d", expr.Type())
				}
				mapIndex, ok := expr.(ExprMapIndex)
				if !ok {
					t.Fatalf("Expected ExprMapIndex, got %T", expr)
				}
				// Check that the index expression is a boolean
				if mapIndex.Index.Type() != ExprType_Boolean {
					t.Fatalf("Expected Boolean index, got %d", mapIndex.Index.Type())
				}
			},
		},
		"Nested indexing": {
			input: `{"outer": {"inner": "value"}}["outer"]["inner"]`,
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_MapIndex {
					t.Fatalf("Expected MapIndex type, got %d", expr.Type())
				}
				mapIndex, ok := expr.(ExprMapIndex)
				if !ok {
					t.Fatalf("Expected ExprMapIndex, got %T", expr)
				}
				// Check that this is a nested indexing operation
				if mapIndex.Map.Type() != ExprType_MapIndex {
					t.Fatalf("Expected nested MapIndex, got %d", mapIndex.Map.Type())
				}
			},
		},
		"Invalid indexing - missing right bracket": {
			input: `{"name": "Andrew"}["name"`,
			validate: func(expr Expr, err error) {
				if err == nil {
					t.Fatalf("Expected error, but got none")
				}
				if !errors.Is(err, ErrExpectedToken) {
					t.Fatalf("Expected ErrExpectedToken, got %s", err)
				}
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.input)
			parser := New(lex)
			expr, err := parser.Parse()
			tc.validate(expr, err)
		})
	}
}

func Test_Parser_Parse_ListSlice(t *testing.T) {
	testCases := map[string]struct {
		input    string
		validate func(Expr, error)
	}{
		"List slice with start and end": {
			input: "[1, 2, 3, 4, 5][1:3]",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_ListSlice {
					t.Fatalf("Expected ListSlice type, got %d", expr.Type())
				}
				slice, ok := expr.(ExprListSlice)
				if !ok {
					t.Fatalf("Expected ExprListSlice, got %T", expr)
				}
				if slice.List.Type() != ExprType_List {
					t.Fatalf("Expected List as list operand, got %d", slice.List.Type())
				}
				if slice.Start.Type() != ExprType_Number {
					t.Fatalf("Expected Number as start operand, got %d", slice.Start.Type())
				}
				if slice.End.Type() != ExprType_Number {
					t.Fatalf("Expected Number as end operand, got %d", slice.End.Type())
				}
			},
		},
		"List slice with only start (omitted end)": {
			input: "[1, 2, 3][1:]",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_ListSlice {
					t.Fatalf("Expected ListSlice type, got %d", expr.Type())
				}
				slice, ok := expr.(ExprListSlice)
				if !ok {
					t.Fatalf("Expected ExprListSlice, got %T", expr)
				}
				if slice.List.Type() != ExprType_List {
					t.Fatalf("Expected List as list operand, got %d", slice.List.Type())
				}
				if slice.Start.Type() != ExprType_Number {
					t.Fatalf("Expected Number as start operand, got %d", slice.Start.Type())
				}
				// End should be nil (optional)
				if slice.End != nil {
					t.Fatalf("Expected nil end operand for slice [1:], got %T", slice.End)
				}
			},
		},
		"List slice with only end (omitted start)": {
			input: "[1, 2, 3][:2]",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_ListSlice {
					t.Fatalf("Expected ListSlice type, got %d", expr.Type())
				}
				slice, ok := expr.(ExprListSlice)
				if !ok {
					t.Fatalf("Expected ExprListSlice, got %T", expr)
				}
				if slice.List.Type() != ExprType_List {
					t.Fatalf("Expected List as list operand, got %d", slice.List.Type())
				}
				if slice.Start != nil {
					t.Fatalf("Expected nil start operand for slice [:2], got %T", slice.Start)
				}
				if slice.End.Type() != ExprType_Number {
					t.Fatalf("Expected Number as end operand, got %d", slice.End.Type())
				}
			},
		},
		"List slice with no bounds (full slice)": {
			input: "[1, 2, 3][:]",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_ListSlice {
					t.Fatalf("Expected ListSlice type, got %d", expr.Type())
				}
				slice, ok := expr.(ExprListSlice)
				if !ok {
					t.Fatalf("Expected ExprListSlice, got %T", expr)
				}
				if slice.List.Type() != ExprType_List {
					t.Fatalf("Expected List as list operand, got %d", slice.List.Type())
				}
				if slice.Start != nil {
					t.Fatalf("Expected nil start operand for slice [:], got %T", slice.Start)
				}
				if slice.End != nil {
					t.Fatalf("Expected nil end operand for slice [:], got %T", slice.End)
				}
			},
		},
		"Complex slice expressions": {
			input: "[1, 2, 3, 4, 5][1+1:4-1]",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_ListSlice {
					t.Fatalf("Expected ListSlice type, got %d", expr.Type())
				}
				slice, ok := expr.(ExprListSlice)
				if !ok {
					t.Fatalf("Expected ExprListSlice, got %T", expr)
				}
				if slice.List.Type() != ExprType_List {
					t.Fatalf("Expected List as list operand, got %d", slice.List.Type())
				}
				if slice.Start.Type() != ExprType_Add {
					t.Fatalf("Expected Add expression as start operand, got %d", slice.Start.Type())
				}
				if slice.End.Type() != ExprType_Subtract {
					t.Fatalf("Expected Subtract expression as end operand, got %d", slice.End.Type())
				}
			},
		},
		"Chained indexing with slice": {
			input: "[[1, 2], [3, 4]][0][1:]",
			validate: func(expr Expr, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if expr.Type() != ExprType_ListSlice {
					t.Fatalf("Expected ListSlice type, got %d", expr.Type())
				}
				slice, ok := expr.(ExprListSlice)
				if !ok {
					t.Fatalf("Expected ExprListSlice, got %T", expr)
				}
				// The list should be a ListIndex operation
				if slice.List.Type() != ExprType_ListIndex {
					t.Fatalf("Expected ListIndex as list operand, got %d", slice.List.Type())
				}
			},
		},
		"Map slice error": {
			input: "{\"a\": 1, \"b\": 2}[\"a\":",
			validate: func(expr Expr, err error) {
				if err == nil {
					t.Fatalf("Expected error for map slicing, but got none")
				}
				// Should return an error about maps not supporting slicing
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.input)
			parser := New(lex)
			expr, err := parser.Parse()
			tc.validate(expr, err)
		})
	}
}
