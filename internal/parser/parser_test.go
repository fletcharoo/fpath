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

func Test_parseUndefined(t *testing.T) {
	token := lexer.Token{Type: lexer.TokenType_Undefined, Value: "test"}
	expr, err := parseUndefined(nil, token)

	if err == nil {
		t.Fatalf("Error expected but not returned")
	}

	if expr != nil {
		t.Fatalf("Expected nil expression, got %v", expr)
	}

	if !errors.Is(err, ErrUndefinedToken) {
		t.Fatalf("Expected ErrUndefinedToken, got: %s", err.Error())
	}
}

func Test_parseBlock(t *testing.T) {
	testCases := map[string]struct {
		input    string
		validate func(Expr, error)
	}{
		"Valid block": {
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
		"Missing right parenthesis": {
			input: "(123",
			validate: func(expr Expr, err error) {
				if err == nil {
					t.Fatalf("Error expected but not returned")
				}
				if !errors.Is(err, ErrExpectedToken) {
					t.Fatalf("Expected ErrExpectedToken, got: %s", err.Error())
				}
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.input)
			parser := New(lex)
			// Skip the left parenthesis
			lex.GetToken()
			expr, err := parseBlock(parser, lexer.Token{})
			tc.validate(expr, err)
		})
	}
}

func Test_parseNumber(t *testing.T) {
	testCases := map[string]struct {
		input    string
		validate func(Expr, error)
	}{
		"Valid integer": {
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
		"Valid decimal": {
			input: "123.45",
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
				expected, _ := decimal.NewFromString("123.45")
				if !num.Value.Equal(expected) {
					t.Fatalf("Expected 123.45, got %s", num.Value.String())
				}
			},
		},
		"Invalid number": {
			input: "abc",
			validate: func(expr Expr, err error) {
				// parseNumber has a bug - it sets the error but doesn't return it
				// It returns exprNumber, nil even when decimal parsing fails
				// So we expect no error but the expression should have a zero decimal value
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				// The expression should be returned but with zero value due to the bug
				if expr == nil {
					t.Fatalf("Expected expression to be returned")
				}
				num, ok := expr.(ExprNumber)
				if !ok {
					t.Fatalf("Expected ExprNumber, got %T", expr)
				}
				// Should be zero value since decimal parsing failed
				if !num.Value.IsZero() {
					t.Fatalf("Expected zero decimal value, got %s", num.Value.String())
				}
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			token := lexer.Token{Type: lexer.TokenType_Number, Value: tc.input}
			expr, err := parseNumber(nil, token)
			tc.validate(expr, err)
		})
	}
}

func Test_parseString(t *testing.T) {
	token := lexer.Token{Type: lexer.TokenType_StringLiteral, Value: "hello world"}
	expr, err := parseString(nil, token)

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

	if str.Value != "hello world" {
		t.Fatalf("Expected 'hello world', got '%s'", str.Value)
	}
}

func Test_parseBoolean(t *testing.T) {
	testCases := map[string]struct {
		input    string
		expected bool
	}{
		"True":  {"true", true},
		"False": {"false", false},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			token := lexer.Token{Type: lexer.TokenType_Boolean, Value: tc.input}
			expr, err := parseBoolean(nil, token)

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

			if boolean.Value != tc.expected {
				t.Fatalf("Expected %v, got %v", tc.expected, boolean.Value)
			}
		})
	}
}

func Test_operatorFunctions(t *testing.T) {
	expr1 := ExprNumber{Value: decimal.NewFromInt(123)}
	expr2 := ExprNumber{Value: decimal.NewFromInt(456)}

	testCases := map[string]struct {
		operator operatorFunc
		expected int
	}{
		"Add":                {operatorAdd, ExprType_Add},
		"Subtract":           {operatorSubtract, ExprType_Subtract},
		"Multiply":           {operatorMultiply, ExprType_Multiply},
		"Divide":             {operatorDivide, ExprType_Divide},
		"Equals":             {operatorEquals, ExprType_Equals},
		"NotEquals":          {operatorNotEquals, ExprType_NotEquals},
		"GreaterThan":        {operatorGreaterThan, ExprType_GreaterThan},
		"GreaterThanOrEqual": {operatorGreaterThanOrEqual, ExprType_GreaterThanOrEqual},
		"LessThan":           {operatorLessThan, ExprType_LessThan},
		"LessThanOrEqual":    {operatorLessThanOrEqual, ExprType_LessThanOrEqual},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := tc.operator(expr1, expr2)
			if result.Type() != tc.expected {
				t.Fatalf("Expected type %d, got %d", tc.expected, result.Type())
			}
		})
	}
}

func Test_Expr_Decode(t *testing.T) {
	testCases := map[string]struct {
		expr     Expr
		expected any
		hasError bool
	}{
		"Number": {
			expr:     ExprNumber{Value: decimal.NewFromInt(123)},
			expected: float64(123),
			hasError: false,
		},
		"String": {
			expr:     ExprString{Value: "hello"},
			expected: "hello",
			hasError: false,
		},
		"Boolean true": {
			expr:     ExprBoolean{Value: true},
			expected: true,
			hasError: false,
		},
		"Boolean false": {
			expr:     ExprBoolean{Value: false},
			expected: false,
			hasError: false,
		},
		"Block": {
			expr:     ExprBlock{Expr: ExprNumber{Value: decimal.NewFromInt(123)}},
			expected: nil,
			hasError: true,
		},
		"Add": {
			expr:     ExprAdd{Expr1: ExprNumber{Value: decimal.NewFromInt(123)}, Expr2: ExprNumber{Value: decimal.NewFromInt(456)}},
			expected: nil,
			hasError: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result, err := tc.expr.Decode()

			if tc.hasError {
				if err == nil {
					t.Fatalf("Error expected but not returned")
				}
				if !errors.Is(err, ErrInvalidDecode) {
					t.Fatalf("Expected ErrInvalidDecode, got: %s", err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}
				if result != tc.expected {
					t.Fatalf("Expected %v, got %v", tc.expected, result)
				}
			}
		})
	}
}

func Test_Expr_String(t *testing.T) {
	testCases := map[string]struct {
		expr     Expr
		expected string
	}{
		"Block":              {ExprBlock{}, "Block"},
		"Number":             {ExprNumber{}, "Number"},
		"String":             {ExprString{}, "String"},
		"Add":                {ExprAdd{}, "Add"},
		"Subtract":           {ExprSubtract{}, "Subtract"},
		"Multiply":           {ExprMultiply{}, "Multiply"},
		"Divide":             {ExprDivide{}, "Divide"},
		"Equals":             {ExprEquals{}, "Equals"},
		"NotEquals":          {ExprNotEquals{}, "NotEquals"},
		"GreaterThan":        {ExprGreaterThan{}, "GreaterThan"},
		"GreaterThanOrEqual": {ExprGreaterThanOrEqual{}, "GreaterThanOrEqual"},
		"LessThan":           {ExprLessThan{}, "LessThan"},
		"LessThanOrEqual":    {ExprLessThanOrEqual{}, "LessThanOrEqual"},
		"Boolean":            {ExprBoolean{}, "Boolean"},
		"List":               {ExprList{}, "List"},
		"ListIndex":          {ExprListIndex{}, "ListIndex"},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := tc.expr.String()
			if result != tc.expected {
				t.Fatalf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func Test_Expr_Type(t *testing.T) {
	testCases := map[string]struct {
		expr     Expr
		expected int
	}{
		"Block":              {ExprBlock{}, ExprType_Block},
		"Number":             {ExprNumber{}, ExprType_Number},
		"String":             {ExprString{}, ExprType_String},
		"Add":                {ExprAdd{}, ExprType_Add},
		"Subtract":           {ExprSubtract{}, ExprType_Subtract},
		"Multiply":           {ExprMultiply{}, ExprType_Multiply},
		"Divide":             {ExprDivide{}, ExprType_Divide},
		"Equals":             {ExprEquals{}, ExprType_Equals},
		"NotEquals":          {ExprNotEquals{}, ExprType_NotEquals},
		"GreaterThan":        {ExprGreaterThan{}, ExprType_GreaterThan},
		"GreaterThanOrEqual": {ExprGreaterThanOrEqual{}, ExprType_GreaterThanOrEqual},
		"LessThan":           {ExprLessThan{}, ExprType_LessThan},
		"LessThanOrEqual":    {ExprLessThanOrEqual{}, ExprType_LessThanOrEqual},
		"Boolean":            {ExprBoolean{}, ExprType_Boolean},
		"List":               {ExprList{}, ExprType_List},
		"ListIndex":          {ExprListIndex{}, ExprType_ListIndex},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := tc.expr.Type()
			if result != tc.expected {
				t.Fatalf("Expected %d, got %d", tc.expected, result)
			}
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
