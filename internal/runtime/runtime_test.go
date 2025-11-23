package runtime_test

import (
	"sort"
	"testing"

	"github.com/fletcharoo/fpath/internal/lexer"
	"github.com/fletcharoo/fpath/internal/parser"
	"github.com/fletcharoo/fpath/internal/runtime"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func Test_Eval(t *testing.T) {
	testCases := map[string]struct {
		query    string
		input    any
		expected float64
	}{
		"number": {
			query:    "2",
			expected: 2.0,
		},
		"add": {
			query:    "2 + 3",
			expected: 5.0,
		},
		"multiply": {
			query:    "2 * 3",
			expected: 6.0,
		},
		"block": {
			query:    "2 + ((5 * 10) * 2) + 2",
			expected: 104.0,
		},
		"subtract": {
			query:    "5 - 2",
			expected: 3.0,
		},
		"subtract left-associative": {
			query:    "10 - 3 - 2",
			expected: 5.0,
		},
		"subtract with addition": {
			query:    "(5 + 3) - 2",
			expected: 6.0,
		},
		"subtract nested": {
			query:    "10 - (2 + 3)",
			expected: 5.0,
		},
		"subtract with multiplication": {
			query:    "10 - 2 * 3",
			expected: 24.0, // Left-associative: (10 - 2) * 3 = 8 * 3 = 24
		},
		"subtract negative result": {
			query:    "2 - 5",
			expected: -3.0,
		},
		"subtract zero": {
			query:    "5 - 5",
			expected: 0.0,
		},
		"divide": {
			query:    "10 / 2",
			expected: 5.0,
		},
		"divide decimal result": {
			query:    "7 / 2",
			expected: 3.5,
		},
		"divide with addition": {
			query:    "(10 + 5) / 3",
			expected: 5.0,
		},
		"divide with multiplication": {
			query:    "20 / (2 * 2)",
			expected: 5.0,
		},
		"divide with subtraction": {
			query:    "15 / (10 - 5)",
			expected: 3.0,
		},
		"divide complex expression": {
			query:    "(10 + 5) / (2 + 1)",
			expected: 5.0,
		},

		"divide left-associative": {
			query:    "100 / 10 / 2",
			expected: 5.0,
		},
		"divide by one": {
			query:    "7 / 1",
			expected: 7.0,
		},
		"divide same numbers": {
			query:    "8 / 8",
			expected: 1.0,
		},
		"modulo": {
			query:    "10 % 3",
			expected: 1.0,
		},
		"modulo decimal result": {
			query:    "10.5 % 3",
			expected: 1.5,
		},
		"modulo with addition": {
			query:    "(10 + 5) % 4",
			expected: 3.0,
		},
		"modulo with multiplication": {
			query:    "20 % (3 * 2)",
			expected: 2.0,
		},
		"modulo with division": {
			query:    "17 % (10 / 2)",
			expected: 2.0,
		},
		"modulo complex expression": {
			query:    "(10 + 5) % (2 * 3)",
			expected: 3.0,
		},
		"modulo left-associative": {
			query:    "100 % 17 % 7",
			expected: 1.0, // (100 % 17) % 7 = 15 % 7 = 1
		},
		"modulo by one": {
			query:    "7 % 1",
			expected: 0.0,
		},
		"modulo same numbers": {
			query:    "8 % 8",
			expected: 0.0,
		},
		"modulo zero": {
			query:    "0 % 7",
			expected: 0.0,
		},
		"modulo larger divisor": {
			query:    "3 % 7",
			expected: 3.0,
		},
		"exponent": {
			query:    "2 ^ 3",
			expected: 8.0,
		},
		"exponent left-associative": {
			query:    "2 ^ 3 ^ 2", // (2 ^ 3) ^ 2 = 8 ^ 2 = 64
			expected: 64.0,
		},
		"exponent with decimal base": {
			query:    "2.5 ^ 2",
			expected: 6.25,
		},
		"exponent with decimal exponent": {
			query:    "4 ^ 1.5",
			expected: 8.0, // 4^1.5 = 4^(3/2) = sqrt(4^3) = sqrt(64) = 8
		},
		"exponent with decimal base and exponent": {
			query:    "2.0 ^ 3.0",
			expected: 8.0,
		},
		"exponent of zero": {
			query:    "5 ^ 0",
			expected: 1.0,
		},
		"zero exponent": {
			query:    "0 ^ 5",
			expected: 0.0,
		},
		"exponent one": {
			query:    "7 ^ 1",
			expected: 7.0,
		},
		"one exponent": {
			query:    "1 ^ 100",
			expected: 1.0,
		},
		"negative base with even exponent": {
			query:    "(-2) ^ 2",
			expected: 4.0,
		},
		"negative base with odd exponent": {
			query:    "(-2) ^ 3",
			expected: -8.0,
		},
		"exponent with addition": {
			query:    "(2 + 3) ^ 2",
			expected: 25.0,
		},
		"exponent with complex expression": {
			query:    "2 ^ (3 * 2)",
			expected: 64.0, // 2^6 = 64
		},
		"exponent with multiplication in chain": {
			query:    "2 ^ 3 * 2", // Left-associative: (2 ^ 3) * 2 = 8 * 2 = 16
			expected: 16.0,
		},
		"exponent with division in chain": {
			query:    "2 ^ 4 / 2", // Left-associative: (2 ^ 4) / 2 = 16 / 2 = 8
			expected: 8.0,
		},
		"exponent with subtraction in chain": {
			query:    "3 ^ 2 - 1", // Left-associative: (3 ^ 2) - 1 = 9 - 1 = 8
			expected: 8.0,
		},
		"integer division simple": {
			query:    "7 // 2",
			expected: 3.0,
		},
		"integer division no remainder": {
			query:    "10 // 3",
			expected: 3.0,
		},
		"integer division same numbers": {
			query:    "5 // 5",
			expected: 1.0,
		},
		"integer division one": {
			query:    "1 // 2",
			expected: 0.0,
		},
		"integer division negative result": {
			query:    "-7 // 2",
			expected: -3.0,
		},
		"integer division with decimals": {
			query:    "7.9 // 2",
			expected: 3.0,
		},
		"integer division with addition": {
			query:    "10 + 6 // 4 * 2", // Left-associative: (10 + 6) // 4 * 2 = 16 // 4 * 2 = 4 * 2 = 8
			expected: 8.0,
		},
		"integer division left-associative": {
			query:    "20 // 3 // 2", // (20 // 3) // 2 = 6 // 2 = 3
			expected: 3.0,
		},
		"integer division with decimals result": {
			query:    "15.7 // 3.2", // 15.7 / 3.2 = 4.906..., truncated to 4
			expected: 4.0,
		},
		"integer division negative divided by positive": {
			query:    "-7 // 3", // -2.333..., truncated to -2
			expected: -2.0,
		},
		"integer division positive divided by negative": {
			query:    "7 // -3", // -2.333..., truncated to -2
			expected: -2.0,
		},
		"integer division negative divided by negative": {
			query:    "-7 // -3", // 2.333..., truncated to 2
			expected: 2.0,
		},
		"integer division zero by positive": {
			query:    "0 // 5",
			expected: 0.0,
		},
		"integer division zero by negative": {
			query:    "0 // -5",
			expected: 0.0,
		},
		"integer division larger divisor": {
			query:    "2 // 5",
			expected: 0.0,
		},
		"integer division with complex expression": {
			query:    "(10 + 5) // (2 * 3)", // 15 // 6 = 2
			expected: 2.0,
		},
		"integer division with parentheses": {
			query:    "(20 // 3) // 2", // 6 // 2 = 3
			expected: 3.0,
		},
		"integer division with block": {
			query:    "((8 + 2) // (2 + 1)) * 3", // (10 // 3) * 3 = 3 * 3 = 9
			expected: 9.0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, nil)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

// normalizeExprMap returns a normalized ExprMap with pairs sorted by key for deterministic comparison
func normalizeExprMap(exprMap parser.ExprMap) parser.ExprMap {
	// Create a copy of pairs to avoid modifying the original
	pairs := make([]parser.ExprMapPair, len(exprMap.Pairs))
	copy(pairs, exprMap.Pairs)

	// Sort pairs by key
	sort.Slice(pairs, func(i, j int) bool {
		keyI := pairs[i].Key.(parser.ExprString).Value
		keyJ := pairs[j].Key.(parser.ExprString).Value
		return keyI < keyJ
	})

	// Recursively normalize nested maps
	for i := range pairs {
		if nestedMap, ok := pairs[i].Value.(parser.ExprMap); ok {
			pairs[i].Value = normalizeExprMap(nestedMap)
		}
	}

	return parser.ExprMap{Pairs: pairs}
}

func Test_Eval_Input(t *testing.T) {
	testCases := map[string]struct {
		query    string
		input    any
		expected any
	}{
		"input number": {
			query:    "$",
			input:    42,
			expected: 42.0,
		},
		"input string": {
			query:    "$",
			input:    "hello",
			expected: "hello",
		},
		"input boolean true": {
			query:    "$",
			input:    true,
			expected: true,
		},
		"input boolean false": {
			query:    "$",
			input:    false,
			expected: false,
		},
		"input float": {
			query:    "$",
			input:    3.0,
			expected: 3.0,
		},
		"input int8": {
			query:    "$",
			input:    int8(8),
			expected: 8.0,
		},
		"input int16": {
			query:    "$",
			input:    int16(16),
			expected: 16.0,
		},
		"input int32": {
			query:    "$",
			input:    int32(32),
			expected: 32.0,
		},
		"input int64": {
			query:    "$",
			input:    int64(64),
			expected: 64.0,
		},
		"input uint": {
			query:    "$",
			input:    uint(10),
			expected: 10.0,
		},
		"input uint8": {
			query:    "$",
			input:    uint8(8),
			expected: 8.0,
		},
		"input uint16": {
			query:    "$",
			input:    uint16(16),
			expected: 16.0,
		},
		"input uint32": {
			query:    "$",
			input:    uint32(32),
			expected: 32.0,
		},
		"input uint64": {
			query:    "$",
			input:    uint64(64),
			expected: 64.0,
		},
		"input float32": {
			query:    "$",
			input:    float32(3.0),
			expected: 3.0,
		},

		"input with addition": {
			query:    "$ + 3",
			input:    1,
			expected: 4.0,
		},
		"input with subtraction": {
			query:    "$ - 2",
			input:    10,
			expected: 8.0,
		},
		"input with multiplication": {
			query:    "$ * 3",
			input:    5,
			expected: 15.0,
		},
		"input with division": {
			query:    "$ / 2",
			input:    8,
			expected: 4.0,
		},
		"input string concatenation": {
			query:    `$ + " world"`,
			input:    "hello",
			expected: "hello world",
		},
		"input equality check": {
			query:    "$ == 4",
			input:    4,
			expected: true,
		},
		"input inequality check": {
			query:    "$ != 5",
			input:    4,
			expected: true,
		},
		"input greater than": {
			query:    "$ > 3",
			input:    4,
			expected: true,
		},
		"input less than": {
			query:    "$ < 5",
			input:    4,
			expected: true,
		},
		"input greater than or equal": {
			query:    "$ >= 4",
			input:    4,
			expected: true,
		},
		"input less than or equal": {
			query:    "$ <= 4",
			input:    4,
			expected: true,
		},
		"input list": {
			query: "$",
			input: []any{1, "two", true},
			expected: func() parser.ExprList {
				return parser.ExprList{
					Values: []parser.Expr{
						parser.ExprNumber{Value: decimal.NewFromInt(1)},
						parser.ExprString{Value: "two"},
						parser.ExprBoolean{Value: true},
					},
				}
			}(),
		},
		"input string list": {
			query: "$",
			input: []string{"a", "b", "c"},
			expected: func() parser.ExprList {
				return parser.ExprList{
					Values: []parser.Expr{
						parser.ExprString{Value: "a"},
						parser.ExprString{Value: "b"},
						parser.ExprString{Value: "c"},
					},
				}
			}(),
		},
		"input int list": {
			query: "$",
			input: []int{1, 2, 3},
			expected: func() parser.ExprList {
				return parser.ExprList{
					Values: []parser.Expr{
						parser.ExprNumber{Value: decimal.NewFromInt(1)},
						parser.ExprNumber{Value: decimal.NewFromInt(2)},
						parser.ExprNumber{Value: decimal.NewFromInt(3)},
					},
				}
			}(),
		},
		"input int64 list": {
			query: "$",
			input: []int64{10, 20, 30},
			expected: func() parser.ExprList {
				return parser.ExprList{
					Values: []parser.Expr{
						parser.ExprNumber{Value: decimal.NewFromInt(10)},
						parser.ExprNumber{Value: decimal.NewFromInt(20)},
						parser.ExprNumber{Value: decimal.NewFromInt(30)},
					},
				}
			}(),
		},
		"input float64 list": {
			query: "$",
			input: []float64{1.0, 2.0, 3.0},
			expected: func() parser.ExprList {
				return parser.ExprList{
					Values: []parser.Expr{
						parser.ExprNumber{Value: decimal.NewFromFloat(1.0)},
						parser.ExprNumber{Value: decimal.NewFromFloat(2.0)},
						parser.ExprNumber{Value: decimal.NewFromFloat(3.0)},
					},
				}
			}(),
		},
		"input map": {
			query: "$",
			input: map[string]any{"name": "Andrew", "age": 30},
			expected: func() parser.ExprMap {
				// Test by evaluating the expression and checking structure
				result, _ := runtime.Eval(parser.ExprInput{}, map[string]any{"name": "Andrew", "age": 30})
				mapExpr := result.(parser.ExprMap)
				return mapExpr
			}(),
		},

		"input nested map": {
			query: "$",
			input: map[string]any{"user": map[string]any{"name": "John", "age": 25}},
			expected: func() parser.ExprMap {
				return parser.ExprMap{
					Pairs: []parser.ExprMapPair{
						{Key: parser.ExprString{Value: "user"}, Value: parser.ExprMap{
							Pairs: []parser.ExprMapPair{
								{Key: parser.ExprString{Value: "name"}, Value: parser.ExprString{Value: "John"}},
								{Key: parser.ExprString{Value: "age"}, Value: parser.ExprNumber{Value: decimal.NewFromInt(25)}},
							},
						}},
					},
				}
			}(),
		},
		"input nested list": {
			query: "$",
			input: []any{[]any{1, 2}, []any{3, 4}},
			expected: func() parser.ExprList {
				return parser.ExprList{
					Values: []parser.Expr{
						parser.ExprList{
							Values: []parser.Expr{
								parser.ExprNumber{Value: decimal.NewFromInt(1)},
								parser.ExprNumber{Value: decimal.NewFromInt(2)},
							},
						},
						parser.ExprList{
							Values: []parser.Expr{
								parser.ExprNumber{Value: decimal.NewFromInt(3)},
								parser.ExprNumber{Value: decimal.NewFromInt(4)},
							},
						},
					},
				}
			}(),
		},
		"input complex nested": {
			query: "$",
			input: map[string]any{"items": []any{map[string]any{"id": 1}, map[string]any{"id": 2}}},
			expected: func() parser.ExprMap {
				return parser.ExprMap{
					Pairs: []parser.ExprMapPair{
						{Key: parser.ExprString{Value: "items"}, Value: parser.ExprList{
							Values: []parser.Expr{
								parser.ExprMap{
									Pairs: []parser.ExprMapPair{
										{Key: parser.ExprString{Value: "id"}, Value: parser.ExprNumber{Value: decimal.NewFromInt(1)}},
									},
								},
								parser.ExprMap{
									Pairs: []parser.ExprMapPair{
										{Key: parser.ExprString{Value: "id"}, Value: parser.ExprNumber{Value: decimal.NewFromInt(2)}},
									},
								},
							},
						}},
					},
				}
			}(),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, tc.input)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			// Normalize both expected and actual ExprMaps for deterministic comparison
			if expectedMap, ok := tc.expected.(parser.ExprMap); ok {
				if actualMap, ok := resultDecoded.(parser.ExprMap); ok {
					require.Equal(t, normalizeExprMap(expectedMap), normalizeExprMap(actualMap), "Result does not match expected value")
					return
				}
			}

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_Input_Indexing(t *testing.T) {
	testCases := map[string]struct {
		query    string
		input    any
		expected any
	}{
		"input list index": {
			query:    "$[0]",
			input:    []string{"first", "second"},
			expected: "first",
		},
		"input list index number": {
			query:    "$[1]",
			input:    []int{10, 20, 30},
			expected: 20.0,
		},
		"input map index string": {
			query:    `$["name"]`,
			input:    map[string]any{"name": "Andrew", "age": 30},
			expected: "Andrew",
		},
		"input map index number as string": {
			query:    `$["1"]`,
			input:    map[string]any{"1": "one", "2": "two"},
			expected: "one",
		},
		"input nested indexing": {
			query:    `$["user"]["name"]`,
			input:    map[string]any{"user": map[string]any{"name": "John"}},
			expected: "John",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, tc.input)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_Input_Error_Cases(t *testing.T) {
	testCases := map[string]struct {
		query     string
		input     any
		expectErr error
	}{
		"nil input": {
			query:     "$",
			input:     nil,
			expectErr: runtime.ErrIncompatibleTypes,
		},
		"unsupported type": {
			query:     "$",
			input:     func() {},
			expectErr: runtime.ErrIncompatibleTypes,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			_, err = runtime.Eval(expr, tc.input)
			require.Error(t, err, "Expected runtime error")
			require.ErrorIs(t, err, tc.expectErr, "Error type mismatch")
		})
	}
}

func Test_Eval_String(t *testing.T) {
	testCases := map[string]struct {
		query    string
		expected string
	}{
		"string literal": {
			query:    `"hello"`,
			expected: "hello",
		},
		"empty string": {
			query:    `""`,
			expected: "",
		},
		"string with spaces": {
			query:    `"hello world"`,
			expected: "hello world",
		},
		"string concatenation": {
			query:    `"hello" + " world"`,
			expected: "hello world",
		},
		"string concatenation multiple": {
			query:    `"a" + "b" + "c"`,
			expected: "abc",
		},
		"string concatenation with empty": {
			query:    `"hello" + "" + "world"`,
			expected: "helloworld",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, nil)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_String_TypeErrors(t *testing.T) {
	testCases := map[string]struct {
		query string
	}{
		"string + number": {
			query: `"hello" + 5`,
		},
		"number + string": {
			query: `5 + "hello"`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			_, err = runtime.Eval(expr, nil)
			require.Error(t, err, "Expected runtime error for type mismatch")
			require.ErrorIs(t, err, runtime.ErrIncompatibleTypes, "Error should be ErrIncompatibleTypes")
		})
	}
}

func Test_Eval_DivideByZero(t *testing.T) {
	testCases := map[string]struct {
		query string
	}{
		"divide by zero": {
			query: "10 / 0",
		},
		"divide by zero in complex expression": {
			query: "(10 + 5) / (5 - 5)",
		},
		"divide by zero with multiplication": {
			query: "10 / (2 * 0)",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			_, err = runtime.Eval(expr, nil)
			require.Error(t, err, "Expected runtime error for division by zero")
			require.ErrorIs(t, err, runtime.ErrDivisionByZero, "Error should be ErrDivisionByZero")
		})
	}
}

func Test_Eval_IntegerDivideByZero(t *testing.T) {
	testCases := map[string]struct {
		query string
	}{
		"integer divide by zero": {
			query: "10 // 0",
		},
		"integer divide by zero in complex expression": {
			query: "(10 + 5) // (5 - 5)",
		},
		"integer divide by zero with multiplication": {
			query: "10 // (2 * 0)",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			_, err = runtime.Eval(expr, nil)
			require.Error(t, err, "Expected runtime error for integer division by zero")
			require.ErrorIs(t, err, runtime.ErrDivisionByZero, "Error should be ErrDivisionByZero")
		})
	}
}

func Test_Eval_ModuloByZero(t *testing.T) {
	testCases := map[string]struct {
		query string
	}{
		"modulo by zero": {
			query: "10 % 0",
		},
		"modulo by zero in complex expression": {
			query: "(10 + 5) % (5 - 5)",
		},
		"modulo by zero with multiplication": {
			query: "10 % (2 * 0)",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			_, err = runtime.Eval(expr, nil)
			require.Error(t, err, "Expected runtime error for modulo by zero")
			require.ErrorIs(t, err, runtime.ErrDivisionByZero, "Error should be ErrDivisionByZero")
		})
	}
}

func Test_Eval_Modulo_TypeErrors(t *testing.T) {
	testCases := map[string]struct {
		query string
	}{
		"number % string": {
			query: `5 % "hello"`,
		},
		"string % number": {
			query: `"hello" % 5`,
		},
		"number % boolean": {
			query: `5 % true`,
		},
		"boolean % number": {
			query: `true % 5`,
		},
		"string % boolean": {
			query: `"hello" % true`,
		},
		"boolean % string": {
			query: `true % "hello"`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			_, err = runtime.Eval(expr, nil)
			require.Error(t, err, "Expected runtime error for type mismatch")
			require.ErrorIs(t, err, runtime.ErrIncompatibleTypes, "Error should be ErrIncompatibleTypes")
		})
	}
}

func Test_Eval_IntegerDivision_TypeErrors(t *testing.T) {
	testCases := map[string]struct {
		query string
	}{
		"number // string": {
			query: `5 // "hello"`,
		},
		"string // number": {
			query: `"hello" // 5`,
		},
		"number // boolean": {
			query: `5 // true`,
		},
		"boolean // number": {
			query: `true // 5`,
		},
		"string // boolean": {
			query: `"hello" // true`,
		},
		"boolean // string": {
			query: `true // "hello"`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			_, err = runtime.Eval(expr, nil)
			require.Error(t, err, "Expected runtime error for type mismatch")
			require.ErrorIs(t, err, runtime.ErrIncompatibleTypes, "Error should be ErrIncompatibleTypes")
		})
	}
}

func Test_Eval_Equals(t *testing.T) {
	testCases := map[string]struct {
		query    string
		expected bool
	}{
		"number equals true": {
			query:    "5 == 5",
			expected: true,
		},
		"number equals false": {
			query:    "5 == 3",
			expected: false,
		},
		"string equals true": {
			query:    `"hello" == "hello"`,
			expected: true,
		},
		"string equals false": {
			query:    `"hello" == "world"`,
			expected: false,
		},
		"empty string equals true": {
			query:    `"" == ""`,
			expected: true,
		},
		"complex expression equals true": {
			query:    "(2 + 3) == 5",
			expected: true,
		},
		"complex expression equals false": {
			query:    "(2 + 3) == 6",
			expected: false,
		},
		"equals with multiplication": {
			query:    "10 == 2 * 5",
			expected: true,
		},
		"equals with division": {
			query:    "5 == 10 / 2",
			expected: true,
		},
		"equals with subtraction": {
			query:    "3 == 5 - 2",
			expected: true,
		},
		"equals with parentheses": {
			query:    "(5 == 5) == true",
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, nil)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_Equals_TypeErrors(t *testing.T) {
	testCases := map[string]struct {
		query string
	}{
		"number == string": {
			query: `5 == "5"`,
		},
		"string == number": {
			query: `"5" == 5`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			_, err = runtime.Eval(expr, nil)
			require.Error(t, err, "Expected runtime error for type mismatch")
			require.ErrorIs(t, err, runtime.ErrIncompatibleTypes, "Error should be ErrIncompatibleTypes")
		})
	}
}

func Test_Eval_NotEquals(t *testing.T) {
	testCases := map[string]struct {
		query    string
		expected bool
	}{
		"number not equals true": {
			query:    "5 != 3",
			expected: true,
		},
		"number not equals false": {
			query:    "5 != 5",
			expected: false,
		},
		"string not equals true": {
			query:    `"hello" != "world"`,
			expected: true,
		},
		"string not equals false": {
			query:    `"hello" != "hello"`,
			expected: false,
		},
		"empty string not equals false": {
			query:    `"" != ""`,
			expected: false,
		},
		"complex expression not equals true": {
			query:    "(2 + 3) != 6",
			expected: true,
		},
		"complex expression not equals false": {
			query:    "(2 + 3) != 5",
			expected: false,
		},
		"not equals with multiplication": {
			query:    "10 != 2 * 6",
			expected: true,
		},
		"not equals with division": {
			query:    "5 != 10 / 3",
			expected: true,
		},
		"not equals with subtraction": {
			query:    "3 != 5 - 2",
			expected: false,
		},
		"not equals with parentheses": {
			query:    "(5 != 5) != true",
			expected: true,
		},
		"boolean not equals true": {
			query:    "true != false",
			expected: true,
		},
		"boolean not equals false": {
			query:    "true != true",
			expected: false,
		},
		"large number not equals true": {
			query:    "314 != 271",
			expected: true,
		},
		"large number not equals false": {
			query:    "314 != 314",
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, nil)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_NotEquals_TypeErrors(t *testing.T) {
	testCases := map[string]struct {
		query string
	}{
		"number != string": {
			query: `5 != "5"`,
		},
		"string != number": {
			query: `"5" != 5`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			_, err = runtime.Eval(expr, nil)
			require.Error(t, err, "Expected runtime error for type mismatch")
			require.ErrorIs(t, err, runtime.ErrIncompatibleTypes, "Error should be ErrIncompatibleTypes")
		})
	}
}

func Test_Eval_GreaterThan(t *testing.T) {
	testCases := map[string]struct {
		query    string
		expected bool
	}{
		"number greater than true": {
			query:    "5 > 3",
			expected: true,
		},
		"number greater than false": {
			query:    "3 > 5",
			expected: false,
		},
		"number greater than equal": {
			query:    "5 > 5",
			expected: false,
		},
		"decimal greater than true": {
			query:    "5.5 > 5.3",
			expected: true,
		},
		"decimal greater than false": {
			query:    "5.3 > 5.5",
			expected: false,
		},
		"string greater than true": {
			query:    `"hello" > "world"`,
			expected: false,
		},
		"string greater than lexicographic true": {
			query:    `"world" > "hello"`,
			expected: true,
		},
		"string greater than false": {
			query:    `"hello" > "hello"`,
			expected: false,
		},
		"empty string greater than false": {
			query:    `"" > ""`,
			expected: false,
		},
		"empty string greater than non-empty false": {
			query:    `"" > "a"`,
			expected: false,
		},
		"non-empty string greater than empty true": {
			query:    `"a" > ""`,
			expected: true,
		},
		"boolean greater than true": {
			query:    "true > false",
			expected: true,
		},
		"boolean greater than false": {
			query:    "false > true",
			expected: false,
		},
		"boolean greater than equal": {
			query:    "true > true",
			expected: false,
		},
		"boolean greater than equal false": {
			query:    "false > false",
			expected: false,
		},
		"complex expression greater than true": {
			query:    "(2 + 3) > 4",
			expected: true,
		},
		"complex expression greater than false": {
			query:    "(2 + 3) > 5",
			expected: false,
		},
		"greater than with multiplication": {
			query:    "11 > 2 * 5",
			expected: true,
		},
		"greater than with division": {
			query:    "6 > 10 / 2",
			expected: true,
		},
		"greater than with subtraction": {
			query:    "4 > 5 - 2",
			expected: true,
		},
		"greater than with parentheses": {
			query:    "(5 > 3) == true",
			expected: true,
		},
		"large number greater than true": {
			query:    "1000 > 500",
			expected: true,
		},
		"large number greater than false": {
			query:    "500 > 1000",
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, nil)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_GreaterThan_TypeErrors(t *testing.T) {
	testCases := map[string]struct {
		query string
	}{
		"number > string": {
			query: `5 > "5"`,
		},
		"string > number": {
			query: `"5" > 5`,
		},
		"boolean > number": {
			query: `true > 5`,
		},
		"number > boolean": {
			query: `5 > false`,
		},
		"boolean > string": {
			query: `true > "hello"`,
		},
		"string > boolean": {
			query: `"hello" > false`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			_, err = runtime.Eval(expr, nil)
			require.Error(t, err, "Expected runtime error for type mismatch")
			require.ErrorIs(t, err, runtime.ErrIncompatibleTypes, "Error should be ErrIncompatibleTypes")
		})
	}
}

func Test_Eval_LessThan(t *testing.T) {
	testCases := map[string]struct {
		query    string
		expected bool
	}{
		"number less than true": {
			query:    "3 < 5",
			expected: true,
		},
		"number less than false": {
			query:    "5 < 3",
			expected: false,
		},
		"number less than equal": {
			query:    "5 < 5",
			expected: false,
		},
		"decimal less than true": {
			query:    "5.3 < 5.5",
			expected: true,
		},
		"decimal less than false": {
			query:    "5.5 < 5.3",
			expected: false,
		},
		"string less than true": {
			query:    `"hello" < "world"`,
			expected: true,
		},
		"string less than lexicographic false": {
			query:    `"world" < "hello"`,
			expected: false,
		},
		"string less than false": {
			query:    `"hello" < "hello"`,
			expected: false,
		},
		"empty string less than false": {
			query:    `"" < ""`,
			expected: false,
		},
		"empty string less than non-empty true": {
			query:    `"" < "a"`,
			expected: true,
		},
		"non-empty string less than empty false": {
			query:    `"a" < ""`,
			expected: false,
		},
		"boolean less than true": {
			query:    "false < true",
			expected: true,
		},
		"boolean less than false": {
			query:    "true < false",
			expected: false,
		},
		"boolean less than equal": {
			query:    "true < true",
			expected: false,
		},
		"boolean less than equal false": {
			query:    "false < false",
			expected: false,
		},
		"complex expression less than true": {
			query:    "(2 + 3) < 6",
			expected: true,
		},
		"complex expression less than false": {
			query:    "(2 + 3) < 5",
			expected: false,
		},
		"less than with multiplication": {
			query:    "9 < 2 * 5",
			expected: true,
		},
		"less than with division": {
			query:    "4 < 10 / 2",
			expected: true,
		},
		"less than with subtraction": {
			query:    "2 < 5 - 2",
			expected: true,
		},
		"less than with parentheses": {
			query:    "(3 < 5) == true",
			expected: true,
		},
		"large number less than true": {
			query:    "500 < 1000",
			expected: true,
		},
		"large number less than false": {
			query:    "1000 < 500",
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, nil)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_LessThan_TypeErrors(t *testing.T) {
	testCases := map[string]struct {
		query string
	}{
		"number < string": {
			query: `5 < "5"`,
		},
		"string < number": {
			query: `"5" < 5`,
		},
		"boolean < number": {
			query: `true < 5`,
		},
		"number < boolean": {
			query: `5 < false`,
		},
		"boolean < string": {
			query: `true < "hello"`,
		},
		"string < boolean": {
			query: `"hello" < false`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			_, err = runtime.Eval(expr, nil)
			require.Error(t, err, "Expected runtime error for type mismatch")
			require.ErrorIs(t, err, runtime.ErrIncompatibleTypes, "Error should be ErrIncompatibleTypes")
		})
	}
}

func Test_Eval_LessThan_ComplexExpressions(t *testing.T) {
	testCases := map[string]struct {
		query    string
		expected bool
	}{
		"less than with addition": {
			query:    "5 < 3 + 3",
			expected: true,
		},
		"less than with subtraction": {
			query:    "5 < 10 - 3",
			expected: true,
		},
		"less than with multiplication": {
			query:    "5 < 2 * 3",
			expected: true,
		},
		"less than with division": {
			query:    "5 < 12 / 2",
			expected: true,
		},
		"less than with nested operations": {
			query:    "5 < (2 + 3) * 2",
			expected: true,
		},
		"less than with complex parentheses": {
			query:    "(5 + 3) < (10 - 1)",
			expected: true,
		},
		"less than with multiple operations": {
			query:    "5 < 2 + 3 + 1",
			expected: true,
		},
		"less than with mixed operations": {
			query:    "3 < 10 / (2 + 1)",
			expected: true,
		},
		"less than with equals": {
			query:    "(5 < 10) == true",
			expected: true,
		},
		"less than with not equals": {
			query:    "(5 < 10) != false",
			expected: true,
		},
		"less than with greater than": {
			query:    "(5 < 10) == (1 > 0)",
			expected: true,
		},
		"less than string concatenation": {
			query:    `"a" < "a" + "b"`,
			expected: true,
		},
		"less than with string operations": {
			query:    `"hello" < "hello" + " world"`,
			expected: true,
		},
		"less than boolean expression": {
			query:    "false < (5 < 10)",
			expected: true,
		},
		"less than with boolean result": {
			query:    "(5 < 10) == true",
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, nil)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_GreaterThanOrEqual(t *testing.T) {
	testCases := map[string]struct {
		query    string
		expected bool
	}{
		"number greater than or equal true": {
			query:    "5 >= 3",
			expected: true,
		},
		"number greater than or equal equal": {
			query:    "5 >= 5",
			expected: true,
		},
		"number greater than or equal false": {
			query:    "3 >= 5",
			expected: false,
		},
		"decimal greater than or equal true": {
			query:    "5.5 >= 5.3",
			expected: true,
		},
		"decimal greater than or equal equal": {
			query:    "5.5 >= 5.5",
			expected: true,
		},
		"decimal greater than or equal false": {
			query:    "5.3 >= 5.5",
			expected: false,
		},
		"string greater than or equal true": {
			query:    `"world" >= "hello"`,
			expected: true,
		},
		"string greater than or equal equal": {
			query:    `"hello" >= "hello"`,
			expected: true,
		},
		"string greater than or equal false": {
			query:    `"hello" >= "world"`,
			expected: false,
		},
		"empty string greater than or equal equal": {
			query:    `"" >= ""`,
			expected: true,
		},
		"empty string greater than or equal false": {
			query:    `"" >= "a"`,
			expected: false,
		},
		"non-empty string greater than or equal true": {
			query:    `"a" >= ""`,
			expected: true,
		},
		"boolean greater than or equal true": {
			query:    "true >= false",
			expected: true,
		},
		"boolean greater than or equal equal true": {
			query:    "true >= true",
			expected: true,
		},
		"boolean greater than or equal equal false": {
			query:    "false >= false",
			expected: true,
		},
		"boolean greater than or equal false": {
			query:    "false >= true",
			expected: false,
		},
		"complex expression greater than or equal true": {
			query:    "(2 + 3) >= 4",
			expected: true,
		},
		"complex expression greater than or equal equal": {
			query:    "(2 + 3) >= 5",
			expected: true,
		},
		"complex expression greater than or equal false": {
			query:    "(2 + 3) >= 6",
			expected: false,
		},
		"greater than or equal with multiplication": {
			query:    "10 >= 2 * 5",
			expected: true,
		},
		"greater than or equal with division": {
			query:    "5 >= 10 / 2",
			expected: true,
		},
		"greater than or equal with subtraction": {
			query:    "3 >= 5 - 2",
			expected: true,
		},
		"greater than or equal with parentheses": {
			query:    "(5 >= 3) == true",
			expected: true,
		},
		"large number greater than or equal true": {
			query:    "1000 >= 500",
			expected: true,
		},
		"large number greater than or equal equal": {
			query:    "1000 >= 1000",
			expected: true,
		},
		"large number greater than or equal false": {
			query:    "500 >= 1000",
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, nil)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_GreaterThanOrEqual_TypeErrors(t *testing.T) {
	testCases := map[string]struct {
		query string
	}{
		"number >= string": {
			query: `5 >= "5"`,
		},
		"string >= number": {
			query: `"5" >= 5`,
		},
		"boolean >= number": {
			query: `true >= 5`,
		},
		"number >= boolean": {
			query: `5 >= false`,
		},
		"boolean >= string": {
			query: `true >= "hello"`,
		},
		"string >= boolean": {
			query: `"hello" >= false`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			_, err = runtime.Eval(expr, nil)
			require.Error(t, err, "Expected runtime error for type mismatch")
			require.ErrorIs(t, err, runtime.ErrIncompatibleTypes, "Error should be ErrIncompatibleTypes")
		})
	}
}

func Test_Eval_GreaterThanOrEqual_ComplexExpressions(t *testing.T) {
	testCases := map[string]struct {
		query    string
		expected bool
	}{
		"greater than or equal with addition": {
			query:    "5 >= 3 + 2",
			expected: true,
		},
		"greater than or equal with subtraction": {
			query:    "5 >= 10 - 5",
			expected: true,
		},
		"greater than or equal with multiplication": {
			query:    "10 >= 2 * 5",
			expected: true,
		},
		"greater than or equal with division": {
			query:    "5 >= 10 / 2",
			expected: true,
		},
		"greater than or equal with nested operations": {
			query:    "10 >= (2 + 3) * 2",
			expected: true,
		},
		"greater than or equal with complex parentheses": {
			query:    "(5 + 3) >= (10 - 2)",
			expected: true,
		},
		"greater than or equal with multiple operations": {
			query:    "5 >= 2 + 3",
			expected: true,
		},
		"greater than or equal with mixed operations": {
			query:    "4 >= 10 / (2 + 1)",
			expected: true,
		},
		"greater than or equal with equals": {
			query:    "(5 >= 3) == true",
			expected: true,
		},
		"greater than or equal with not equals": {
			query:    "(5 >= 10) != true",
			expected: true,
		},
		"greater than or equal with greater than": {
			query:    "(5 >= 3) == (1 > 0)",
			expected: true,
		},
		"greater than or equal string concatenation": {
			query:    `"ab" >= "a" + "b"`,
			expected: true,
		},
		"greater than or equal with string operations": {
			query:    `"hello world" >= "hello" + " world"`,
			expected: true,
		},
		"greater than or equal boolean expression": {
			query:    "true >= (5 < 10)",
			expected: true,
		},
		"greater than or equal with boolean result": {
			query:    "(5 >= 3) == true",
			expected: true,
		},
		"greater than or equal with less than": {
			query:    "(10 >= 5) == (3 < 10)",
			expected: true,
		},

		"greater than or equal with decimal arithmetic": {
			query:    "5.5 >= 2.5 + 3.0",
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, nil)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_LessThanOrEqual(t *testing.T) {
	testCases := map[string]struct {
		query    string
		expected bool
	}{
		"number less than or equal true": {
			query:    "3 <= 5",
			expected: true,
		},
		"number less than or equal equal": {
			query:    "5 <= 5",
			expected: true,
		},
		"number less than or equal false": {
			query:    "5 <= 3",
			expected: false,
		},
		"decimal less than or equal true": {
			query:    "5.3 <= 5.5",
			expected: true,
		},
		"decimal less than or equal equal": {
			query:    "5.5 <= 5.5",
			expected: true,
		},
		"decimal less than or equal false": {
			query:    "5.5 <= 5.3",
			expected: false,
		},
		"string less than or equal true": {
			query:    `"hello" <= "world"`,
			expected: true,
		},
		"string less than or equal equal": {
			query:    `"hello" <= "hello"`,
			expected: true,
		},
		"string less than or equal false": {
			query:    `"world" <= "hello"`,
			expected: false,
		},
		"empty string less than or equal equal": {
			query:    `"" <= ""`,
			expected: true,
		},
		"empty string less than or equal true": {
			query:    `"" <= "a"`,
			expected: true,
		},
		"non-empty string less than or equal false": {
			query:    `"a" <= ""`,
			expected: false,
		},
		"boolean less than or equal true": {
			query:    "false <= true",
			expected: true,
		},
		"boolean less than or equal equal true": {
			query:    "true <= true",
			expected: true,
		},
		"boolean less than or equal equal false": {
			query:    "false <= false",
			expected: true,
		},
		"boolean less than or equal false": {
			query:    "true <= false",
			expected: false,
		},
		"complex expression less than or equal true": {
			query:    "(2 + 3) <= 6",
			expected: true,
		},
		"complex expression less than or equal equal": {
			query:    "(2 + 3) <= 5",
			expected: true,
		},
		"complex expression less than or equal false": {
			query:    "(2 + 3) <= 4",
			expected: false,
		},
		"less than or equal with multiplication": {
			query:    "9 <= 2 * 5",
			expected: true,
		},
		"less than or equal with division": {
			query:    "5 <= 10 / 2",
			expected: true,
		},
		"less than or equal with subtraction": {
			query:    "3 <= 5 - 2",
			expected: true,
		},
		"less than or equal with parentheses": {
			query:    "(3 <= 5) == true",
			expected: true,
		},
		"large number less than or equal true": {
			query:    "500 <= 1000",
			expected: true,
		},
		"large number less than or equal equal": {
			query:    "1000 <= 1000",
			expected: true,
		},
		"large number less than or equal false": {
			query:    "1000 <= 500",
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, nil)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_LessThanOrEqual_TypeErrors(t *testing.T) {
	testCases := map[string]struct {
		query string
	}{
		"number <= string": {
			query: `5 <= "5"`,
		},
		"string <= number": {
			query: `"5" <= 5`,
		},
		"boolean <= number": {
			query: `true <= 5`,
		},
		"number <= boolean": {
			query: `5 <= false`,
		},
		"boolean <= string": {
			query: `true <= "hello"`,
		},
		"string <= boolean": {
			query: `"hello" <= false`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			_, err = runtime.Eval(expr, nil)
			require.Error(t, err, "Expected runtime error for type mismatch")
			require.ErrorIs(t, err, runtime.ErrIncompatibleTypes, "Error should be ErrIncompatibleTypes")
		})
	}
}

func Test_Eval_LessThanOrEqual_ComplexExpressions(t *testing.T) {
	testCases := map[string]struct {
		query    string
		expected bool
	}{
		"less than or equal with addition": {
			query:    "5 <= 3 + 3",
			expected: true,
		},
		"less than or equal with subtraction": {
			query:    "5 <= 10 - 3",
			expected: true,
		},
		"less than or equal with multiplication": {
			query:    "10 <= 2 * 5",
			expected: true,
		},
		"less than or equal with division": {
			query:    "5 <= 12 / 2",
			expected: true,
		},
		"less than or equal with nested operations": {
			query:    "10 <= (2 + 3) * 2",
			expected: true,
		},
		"less than or equal with complex parentheses": {
			query:    "(5 + 3) <= (10 - 1)",
			expected: true,
		},
		"less than or equal with multiple operations": {
			query:    "5 <= 2 + 3",
			expected: true,
		},
		"less than or equal with mixed operations": {
			query:    "3 <= 10 / (2 + 1)",
			expected: true,
		},
		"less than or equal with equals": {
			query:    "(5 <= 10) == true",
			expected: true,
		},
		"less than or equal with not equals": {
			query:    "(5 <= 10) != false",
			expected: true,
		},
		"less than or equal with greater than": {
			query:    "(5 <= 10) == (1 > 0)",
			expected: true,
		},
		"less than or equal string concatenation": {
			query:    `"a" <= "a" + "b"`,
			expected: true,
		},
		"less than or equal with string operations": {
			query:    `"hello" <= "hello" + " world"`,
			expected: true,
		},
		"less than or equal boolean expression": {
			query:    "false <= (5 < 10)",
			expected: true,
		},
		"less than or equal with boolean result": {
			query:    "(5 <= 10) == true",
			expected: true,
		},
		"less than or equal with less than": {
			query:    "(5 <= 10) == (3 < 10)",
			expected: true,
		},
		"less than or equal with decimal arithmetic": {
			query:    "5.0 <= 2.5 + 3.0",
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, nil)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_And(t *testing.T) {
	testCases := map[string]struct {
		query    string
		expected bool
	}{
		"true && true": {
			query:    "true && true",
			expected: true,
		},
		"true && false": {
			query:    "true && false",
			expected: false,
		},
		"false && true": {
			query:    "false && true",
			expected: false,
		},
		"false && false": {
			query:    "false && false",
			expected: false,
		},
		"and with equals true": {
			query:    "(true == true) && (false == false)",
			expected: true,
		},
		"and with equals false": {
			query:    "(true == false) && (true == true)",
			expected: false,
		},
		"and with not equals true": {
			query:    "(true != false) && (false != true)",
			expected: true,
		},
		"and with not equals false": {
			query:    "(true != true) && (false != false)",
			expected: false,
		},
		"and with greater than true": {
			query:    "(5 > 3) && (10 > 5)",
			expected: true,
		},
		"and with greater than false": {
			query:    "(5 > 10) && (10 > 5)",
			expected: false,
		},
		"and with less than true": {
			query:    "(3 < 5) && (5 < 10)",
			expected: true,
		},
		"and with less than false": {
			query:    "(10 < 5) && (5 < 10)",
			expected: false,
		},
		"and with greater than or equal true": {
			query:    "(5 >= 3) && (10 >= 10)",
			expected: true,
		},
		"and with greater than or equal false": {
			query:    "(5 >= 10) && (10 >= 5)",
			expected: false,
		},
		"and with less than or equal true": {
			query:    "(3 <= 5) && (10 <= 10)",
			expected: true,
		},
		"and with less than or equal false": {
			query:    "(10 <= 5) && (5 <= 10)",
			expected: false,
		},
		"complex and expression true": {
			query:    "(5 > 3) && (10 < 20) && (15 >= 10)",
			expected: true,
		},
		"complex and expression false": {
			query:    "(5 > 3) && (10 < 5) && (15 >= 10)",
			expected: false,
		},
		"and with arithmetic true": {
			query:    "((2 + 3) == 5) && ((10 - 5) == 5)",
			expected: true,
		},
		"and with arithmetic false": {
			query:    "((2 + 3) == 6) && ((10 - 5) == 5)",
			expected: false,
		},
		"nested and with parentheses true": {
			query:    "((true == true) && (false == false)) && ((5 > 3) && (10 < 20))",
			expected: true,
		},
		"nested and with parentheses false": {
			query:    "((true == true) && (false == true)) && ((5 > 3) && (10 < 20))",
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, nil)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_And_TypeErrors(t *testing.T) {
	testCases := map[string]struct {
		query string
	}{
		"boolean && number": {
			query: "true && 5",
		},
		"number && boolean": {
			query: "5 && true",
		},
		"number && number": {
			query: "5 && 10",
		},
		"boolean && string": {
			query: `true && "hello"`,
		},
		"string && boolean": {
			query: `"hello" && true`,
		},
		"string && string": {
			query: `"hello" && "world"`,
		},
		"and with mixed types": {
			query: "(5 == 5) && \"hello\"",
		},
		"and with arithmetic result": {
			query: "true && (2 + 3)",
		},
		"and with string concatenation": {
			query: "true && (\"hello\" + \"world\")",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			_, err = runtime.Eval(expr, nil)
			require.Error(t, err, "Expected runtime error for type mismatch")
			require.ErrorIs(t, err, runtime.ErrBooleanOperation, "Error should be ErrBooleanOperation")
		})
	}
}

func Test_Eval_And_ShortCircuit(t *testing.T) {
	testCases := map[string]struct {
		query    string
		expected bool
	}{
		"short circuit false && true": {
			query:    "false && true",
			expected: false,
		},
		"short circuit false && false": {
			query:    "false && false",
			expected: false,
		},
		"short circuit false && complex": {
			query:    "false && (5 == 5)",
			expected: false,
		},
		"short circuit false && invalid": {
			query:    "false && (5 == \"hello\")", // This would fail type checking if evaluated
			expected: false,
		},
		"short circuit with comparison": {
			query:    "(5 > 10) && (10 / 0 == 1)", // Division by zero would error if evaluated
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, nil)
			require.NoError(t, err, "Unexpected runtime error (short-circuit should prevent errors)")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_Or(t *testing.T) {
	testCases := map[string]struct {
		query    string
		expected bool
	}{
		"true || true": {
			query:    "true || true",
			expected: true,
		},
		"true || false": {
			query:    "true || false",
			expected: true,
		},
		"false || true": {
			query:    "false || true",
			expected: true,
		},
		"false || false": {
			query:    "false || false",
			expected: false,
		},
		"or with equals true": {
			query:    "(true == true) || (false == false)",
			expected: true,
		},
		"or with equals false": {
			query:    "(true == false) || (false == true)",
			expected: false,
		},
		"or with not equals true": {
			query:    "(true != false) || (false != true)",
			expected: true,
		},
		"or with not equals mixed": {
			query:    "(true != true) || (false != false)",
			expected: false,
		},
		"or with greater than true": {
			query:    "(5 > 3) || (10 > 5)",
			expected: true,
		},
		"or with greater than mixed": {
			query:    "(5 > 10) || (10 > 5)",
			expected: true,
		},
		"or with greater than false": {
			query:    "(5 > 10) || (10 > 20)",
			expected: false,
		},
		"or with less than true": {
			query:    "(3 < 5) || (5 < 10)",
			expected: true,
		},
		"or with less than mixed": {
			query:    "(10 < 5) || (5 < 10)",
			expected: true,
		},
		"or with less than false": {
			query:    "(10 < 5) || (5 < 1)",
			expected: false,
		},
		"or with greater than or equal true": {
			query:    "(5 >= 3) || (10 >= 10)",
			expected: true,
		},
		"or with greater than or equal mixed": {
			query:    "(5 >= 10) || (10 >= 5)",
			expected: true,
		},
		"or with greater than or equal false": {
			query:    "(5 >= 10) || (10 >= 20)",
			expected: false,
		},
		"or with less than or equal true": {
			query:    "(3 <= 5) || (10 <= 10)",
			expected: true,
		},
		"or with less than or equal mixed": {
			query:    "(10 <= 5) || (5 <= 10)",
			expected: true,
		},
		"or with less than or equal false": {
			query:    "(10 <= 5) || (5 <= 1)",
			expected: false,
		},
		"complex or expression true": {
			query:    "(5 > 3) || (10 < 5) || (15 >= 20)",
			expected: true,
		},
		"complex or expression false": {
			query:    "(5 > 10) || (10 < 5) || (15 >= 20)",
			expected: false,
		},
		"or with arithmetic true": {
			query:    "((2 + 3) == 5) || ((10 - 5) == 3)",
			expected: true,
		},
		"or with arithmetic mixed": {
			query:    "((2 + 3) == 6) || ((10 - 5) == 5)",
			expected: true,
		},
		"or with arithmetic false": {
			query:    "((2 + 3) == 6) || ((10 - 5) == 3)",
			expected: false,
		},
		"nested or with parentheses true": {
			query:    "((true == true) || (false == true)) || ((5 > 3) || (10 < 5))",
			expected: true,
		},
		"nested or with parentheses mixed": {
			query:    "((true == false) || (true == true)) || ((5 > 10) || (10 < 5))",
			expected: true,
		},
		"nested or with parentheses false": {
			query:    "((true == false) || (false == true)) || ((5 > 10) || (10 < 5))",
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, nil)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_Or_TypeErrors(t *testing.T) {
	testCases := map[string]struct {
		query string
	}{
		"boolean || number": {
			query: "false || 5",
		},
		"number || boolean": {
			query: "5 || false",
		},
		"number || number": {
			query: "5 || 10",
		},
		"boolean || string": {
			query: `false || "hello"`,
		},
		"string || boolean": {
			query: `"hello" || false`,
		},
		"string || string": {
			query: `"hello" || "world"`,
		},
		"or with mixed types": {
			query: "(5 != 5) || \"hello\"",
		},
		"or with arithmetic result": {
			query: "false || (2 + 3)",
		},
		"or with string concatenation": {
			query: "false || (\"hello\" + \"world\")",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			_, err = runtime.Eval(expr, nil)
			require.Error(t, err, "Expected runtime error for type mismatch")
			require.ErrorIs(t, err, runtime.ErrBooleanOperation, "Error should be ErrBooleanOperation")
		})
	}
}

func Test_Eval_Or_ShortCircuit(t *testing.T) {
	testCases := map[string]struct {
		query    string
		expected bool
	}{
		"short circuit true || false": {
			query:    "true || false",
			expected: true,
		},
		"short circuit true || true": {
			query:    "true || true",
			expected: true,
		},
		"short circuit true || complex": {
			query:    "true || (5 == 5)",
			expected: true,
		},
		"short circuit true || invalid": {
			query:    "true || (5 == \"hello\")", // This would fail type checking if evaluated
			expected: true,
		},
		"short circuit with comparison": {
			query:    "(5 < 10) || (10 / 0 == 1)", // Division by zero would error if evaluated
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, nil)
			require.NoError(t, err, "Unexpected runtime error (short-circuit should prevent errors)")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_Ternary(t *testing.T) {
	testCases := map[string]struct {
		query    string
		expected any
	}{
		"true branch": {
			query:    `true ? "yes" : "no"`,
			expected: "yes",
		},
		"false branch": {
			query:    `false ? "yes" : "no"`,
			expected: "no",
		},
		"complex condition true": {
			query:    `5 > 3 ? "greater" : "less"`,
			expected: "greater",
		},
		"complex condition false": {
			query:    `2 > 3 ? "greater" : "less"`,
			expected: "less",
		},
		"nested ternary true": {
			query:    `true ? (false ? "a" : "b") : "c"`,
			expected: "b",
		},
		"nested ternary false": {
			query:    `false ? (true ? "a" : "b") : "c"`,
			expected: "c",
		},
		"right-associative ternary": {
			query:    `true ? "yes" : false ? "no" : "maybe"`,
			expected: "yes",
		},
		"right-associative ternary false": {
			query:    `false ? "yes" : true ? "no" : "maybe"`,
			expected: "no",
		},
		"number expressions": {
			query:    `true ? 1 : 0`,
			expected: 1.0,
		},
		"boolean expressions": {
			query:    `true ? true : false`,
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			parser := parser.New(lex)
			expr, err := parser.Parse()
			require.NoError(t, err, "Failed to parse query")

			result, err := runtime.Eval(expr, nil)
			require.NoError(t, err, "Failed to evaluate expression")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_Ternary_ShortCircuit(t *testing.T) {
	// Test that only the appropriate branch is evaluated
	// This is hard to test directly without side effects, but we can test
	// that errors in the unchosen branch are not triggered

	testCases := map[string]struct {
		query    string
		expected any
	}{
		"true branch - false branch has error": {
			query:    `true ? "yes" : (1 / 0)`,
			expected: "yes",
		},
		"false branch - true branch has error": {
			query:    `false ? (1 / 0) : "no"`,
			expected: "no",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			parser := parser.New(lex)
			expr, err := parser.Parse()
			require.NoError(t, err, "Failed to parse query")

			result, err := runtime.Eval(expr, nil)
			require.NoError(t, err, "Failed to evaluate expression")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_Ternary_TypeErrors(t *testing.T) {
	testCases := map[string]struct {
		query         string
		expectedError error
	}{
		"non-boolean condition": {
			query:         `"string" ? "yes" : "no"`,
			expectedError: runtime.ErrBooleanOperation,
		},
		"number condition": {
			query:         `5 ? "yes" : "no"`,
			expectedError: runtime.ErrBooleanOperation,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			parser := parser.New(lex)
			expr, err := parser.Parse()
			require.NoError(t, err, "Failed to parse query")

			_, err = runtime.Eval(expr, nil)
			require.Error(t, err, "Expected error when evaluating ternary")
			require.ErrorIs(t, err, tc.expectedError, "Error should be of expected type")
		})
	}
}

func Test_Eval_List(t *testing.T) {
	testCases := map[string]struct {
		query    string
		expected []any
	}{
		"empty list": {
			query:    "[]",
			expected: []any{},
		},
		"list with numbers": {
			query:    "[1, 2, 3]",
			expected: []any{1.0, 2.0, 3.0},
		},
		"list with mixed types": {
			query:    "[1, true, \"hello\"]",
			expected: []any{1.0, true, "hello"},
		},
		"list with expressions": {
			query:    "[1+2, 3*4, 5-1]",
			expected: []any{3.0, 12.0, 4.0},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, nil)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			// For lists, we need to check the decoded values
			resultList, ok := resultDecoded.(parser.ExprList)
			require.True(t, ok, "Expected ExprList result")

			require.Equal(t, len(tc.expected), len(resultList.Values), "List length mismatch")

			for i, expectedValue := range tc.expected {
				valueDecoded, err := resultList.Values[i].Decode()
				require.NoError(t, err, "Failed to decode list element %d", i)
				require.Equal(t, expectedValue, valueDecoded, "Element %d does not match expected value", i)
			}
		})
	}
}

func Test_Eval_ListIndex(t *testing.T) {
	testCases := map[string]struct {
		query    string
		expected any
	}{
		"index into number list": {
			query:    "[1, 2, 3][0]",
			expected: 1.0,
		},
		"index into mixed list": {
			query:    "[1, true, \"hello\"][2]",
			expected: "hello",
		},
		"index with expression": {
			query:    "[1, 2, 3][1+1]",
			expected: 3.0,
		},
		"chained indexing": {
			query:    "[[1, 2], [3, 4]][0][1]",
			expected: 2.0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, nil)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_ListIndex_Errors(t *testing.T) {
	testCases := map[string]struct {
		query             string
		expectedErrorType error
	}{
		"index out of bounds - negative": {
			query:             "[1, 2, 3][-1]",
			expectedErrorType: runtime.ErrIndexOutOfBounds,
		},
		"index out of bounds - too large": {
			query:             "[1, 2, 3][5]",
			expectedErrorType: runtime.ErrIndexOutOfBounds,
		},
		"index into empty list": {
			query:             "[][0]",
			expectedErrorType: runtime.ErrIndexOutOfBounds,
		},
		"non-integer index": {
			query:             "[1, 2, 3][1.5]",
			expectedErrorType: runtime.ErrInvalidIndex,
		},
		"index into non-list": {
			query:             "123[0]",
			expectedErrorType: runtime.ErrInvalidIndex,
		},
		"non-number index": {
			query:             `[1, 2, 3]["hello"]`,
			expectedErrorType: runtime.ErrInvalidMapIndex,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			_, err = runtime.Eval(expr, nil)
			require.Error(t, err, "Expected runtime error")
			require.ErrorIs(t, err, tc.expectedErrorType, "Error should be of expected type")
		})
	}
}

func Test_Eval_StringIndex(t *testing.T) {
	testCases := map[string]struct {
		query    string
		input    any
		expected any
	}{
		"index into string first char": {
			query:    `"hello"[0]`,
			input:    nil,
			expected: "h",
		},
		"index into string middle char": {
			query:    `"hello"[1]`,
			input:    nil,
			expected: "e",
		},
		"index into string last char": {
			query:    `"hello"[4]`,
			input:    nil,
			expected: "o",
		},
		"index with expression": {
			query:    `"hello"[1+1]`,
			input:    nil,
			expected: "l",
		},
		"chained indexing with mixed types": {
			query:    `[[1, 2], "hello"][1][0]`,
			input:    nil,
			expected: "h",
		},
		"string indexing from input": {
			query:    `$[1]`,
			input:    "world",
			expected: "o",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, tc.input)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_StringIndex_Errors(t *testing.T) {
	testCases := map[string]struct {
		query             string
		expectedErrorType error
	}{
		"index out of bounds - negative": {
			query:             `"hello"[-1]`,
			expectedErrorType: runtime.ErrIndexOutOfBounds,
		},
		"index out of bounds - too large": {
			query:             `"hello"[5]`,
			expectedErrorType: runtime.ErrIndexOutOfBounds,
		},
		"index into empty string": {
			query:             `""[0]`,
			expectedErrorType: runtime.ErrIndexOutOfBounds,
		},
		"non-integer index": {
			query:             `"hello"[1.5]`,
			expectedErrorType: runtime.ErrInvalidIndex,
		},
		"non-number index": {
			query:             `"hello"["hello"]`,
			expectedErrorType: runtime.ErrInvalidMapIndex,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			_, err = runtime.Eval(expr, nil)
			require.Error(t, err, "Expected runtime error")
			require.ErrorIs(t, err, tc.expectedErrorType, "Error should be of expected type")
		})
	}
}

func Test_Eval_Map(t *testing.T) {
	testCases := map[string]struct {
		query    string
		validate func(any, error)
	}{
		"empty map": {
			query: "{}",
			validate: func(result any, err error) {
				require.NoError(t, err, "Unexpected runtime error")

				mapExpr, ok := result.(parser.ExprMap)
				require.True(t, ok, "Expected ExprMap result")
				require.Equal(t, 0, len(mapExpr.Pairs), "Expected 0 pairs")
			},
		},
		"single pair": {
			query: `{"key": "value"}`,
			validate: func(result any, err error) {
				require.NoError(t, err, "Unexpected runtime error")

				mapExpr, ok := result.(parser.ExprMap)
				require.True(t, ok, "Expected ExprMap result")
				require.Equal(t, 1, len(mapExpr.Pairs), "Expected 1 pair")

				// Check key
				keyExpr := mapExpr.Pairs[0].Key
				require.Equal(t, parser.ExprType_String, keyExpr.Type(), "Expected String key")
				keyStr, ok := keyExpr.(parser.ExprString)
				require.True(t, ok, "Expected ExprString key")
				require.Equal(t, "key", keyStr.Value, "Expected key 'key'")

				// Check value
				valueExpr := mapExpr.Pairs[0].Value
				require.Equal(t, parser.ExprType_String, valueExpr.Type(), "Expected String value")
				valueStr, ok := valueExpr.(parser.ExprString)
				require.True(t, ok, "Expected ExprString value")
				require.Equal(t, "value", valueStr.Value, "Expected value 'value'")
			},
		},
		"multiple pairs": {
			query: `{"name": "Andrew", "age": 30}`,
			validate: func(result any, err error) {
				require.NoError(t, err, "Unexpected runtime error")

				mapExpr, ok := result.(parser.ExprMap)
				require.True(t, ok, "Expected ExprMap result")
				require.Equal(t, 2, len(mapExpr.Pairs), "Expected 2 pairs")
			},
		},
		"mixed types": {
			query: `{"name": "Andrew", 1: true, "active": false}`,
			validate: func(result any, err error) {
				require.NoError(t, err, "Unexpected runtime error")

				mapExpr, ok := result.(parser.ExprMap)
				require.True(t, ok, "Expected ExprMap result")
				require.Equal(t, 3, len(mapExpr.Pairs), "Expected 3 pairs")

				// Check first pair (string key, string value)
				require.Equal(t, parser.ExprType_String, mapExpr.Pairs[0].Key.Type(), "Expected String key")
				require.Equal(t, parser.ExprType_String, mapExpr.Pairs[0].Value.Type(), "Expected String value")

				// Check second pair (number key, boolean value)
				require.Equal(t, parser.ExprType_Number, mapExpr.Pairs[1].Key.Type(), "Expected Number key")
				require.Equal(t, parser.ExprType_Boolean, mapExpr.Pairs[1].Value.Type(), "Expected Boolean value")

				// Check third pair (string key, boolean value)
				require.Equal(t, parser.ExprType_String, mapExpr.Pairs[2].Key.Type(), "Expected String key")
				require.Equal(t, parser.ExprType_Boolean, mapExpr.Pairs[2].Value.Type(), "Expected Boolean value")
			},
		},
		"nested map": {
			query: `{"outer": {"inner": "value"}}`,
			validate: func(result any, err error) {
				require.NoError(t, err, "Unexpected runtime error")

				mapExpr, ok := result.(parser.ExprMap)
				require.True(t, ok, "Expected ExprMap result")
				require.Equal(t, 1, len(mapExpr.Pairs), "Expected 1 pair")

				// Check that the value is a nested map
				valueExpr := mapExpr.Pairs[0].Value
				require.Equal(t, parser.ExprType_Map, valueExpr.Type(), "Expected nested Map value")

				nestedMap, ok := valueExpr.(parser.ExprMap)
				require.True(t, ok, "Expected ExprMap value")
				require.Equal(t, 1, len(nestedMap.Pairs), "Expected 1 pair in nested map")
			},
		},
		"map with expressions": {
			query: `{"sum": 1+2, "concat": "a" + "b"}`,
			validate: func(result any, err error) {
				require.NoError(t, err, "Unexpected runtime error")

				mapExpr, ok := result.(parser.ExprMap)
				require.True(t, ok, "Expected ExprMap result")
				require.Equal(t, 2, len(mapExpr.Pairs), "Expected 2 pairs")

				// Check that expressions were evaluated
				// First pair: 1+2 should be evaluated to 3
				valueExpr1 := mapExpr.Pairs[0].Value
				require.Equal(t, parser.ExprType_Number, valueExpr1.Type(), "Expected evaluated Number value")
				num1, ok := valueExpr1.(parser.ExprNumber)
				require.True(t, ok, "Expected ExprNumber")
				expectedNum1, _ := decimal.NewFromString("3")
				require.True(t, num1.Value.Equal(expectedNum1), "Expected 1+2 = 3")

				// Second pair: "a"+"b" should be evaluated to "ab"
				valueExpr2 := mapExpr.Pairs[1].Value
				require.Equal(t, parser.ExprType_String, valueExpr2.Type(), "Expected evaluated String value")
				str2, ok := valueExpr2.(parser.ExprString)
				require.True(t, ok, "Expected ExprString")
				require.Equal(t, "ab", str2.Value, "Expected 'a'+'b' = 'ab'")
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, nil)
			tc.validate(result, err)
		})
	}
}

func Test_Eval_MapIndex(t *testing.T) {
	testCases := map[string]struct {
		query    string
		expected any
	}{
		"string key indexing": {
			query:    `{"name": "Andrew"}["name"]`,
			expected: "Andrew",
		},
		"number key indexing": {
			query:    `{1: "value"}[1]`,
			expected: "value",
		},
		"boolean key indexing": {
			query:    `{true: "value"}[true]`,
			expected: "value",
		},
		"nested indexing": {
			query:    `{"outer": {"inner": "value"}}["outer"]["inner"]`,
			expected: "value",
		},
		"indexing with evaluated expressions": {
			query:    `{"1": "first", "2": "second"}[1+1]`,
			expected: "second",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, nil)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_MapIndex_Errors(t *testing.T) {
	testCases := map[string]struct {
		query             string
		expectedErrorType error
	}{
		"key not found": {
			query:             `{"name": "Andrew"}["age"]`,
			expectedErrorType: runtime.ErrKeyNotFound,
		},
		"indexing non-map": {
			query:             `123["key"]`,
			expectedErrorType: runtime.ErrInvalidMapIndex,
		},
		"indexing list as map": {
			query:             `[1, 2, 3]["key"]`,
			expectedErrorType: runtime.ErrInvalidMapIndex,
		},
		"indexing string as map": {
			query:             `"hello"["key"]`,
			expectedErrorType: runtime.ErrInvalidMapIndex,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			_, err = runtime.Eval(expr, nil)
			require.Error(t, err, "Expected runtime error")
			require.ErrorIs(t, err, tc.expectedErrorType, "Error should be of expected type")
		})
	}
}

func Test_Eval_Function(t *testing.T) {
	testCases := map[string]struct {
		query    string
		input    any
		expected any
	}{
		"len with string literal": {
			query:    `len("hello")`,
			expected: 5.0,
		},
		"len with empty string": {
			query:    `len("")`,
			expected: 0.0,
		},
		"len with list literal": {
			query:    `len([1, 2, 3, 4, 5])`,
			expected: 5.0,
		},
		"len with empty list": {
			query:    `len([])`,
			expected: 0.0,
		},
		"len with map literal": {
			query:    `len({"a": 1, "b": 2})`,
			expected: 2.0,
		},
		"len with empty map": {
			query:    `len({})`,
			expected: 0.0,
		},
		"len with input string": {
			query:    `len($)`,
			input:    "hello world",
			expected: 11.0,
		},
		"len with input list": {
			query:    `len($)`,
			input:    []any{1, 2, 3},
			expected: 3.0,
		},
		"len with input map": {
			query:    `len($)`,
			input:    map[string]any{"a": 1, "b": 2, "c": 3},
			expected: 3.0,
		},
		"len in complex expression": {
			query:    `len("test") + 2`,
			expected: 6.0,
		},
		"len in comparison": {
			query:    `len("hello") == 5`,
			expected: true,
		},
		"abs with positive integer": {
			query:    `abs(5)`,
			expected: 5.0,
		},
		"abs with negative integer": {
			query:    `abs(-5)`,
			expected: 5.0,
		},
		"abs with zero": {
			query:    `abs(0)`,
			expected: 0.0,
		},
		"abs with positive decimal": {
			query:    `abs(3.0)`,
			expected: 3.0,
		},
		"abs with negative decimal": {
			query:    `abs(-3.0)`,
			expected: 3.0,
		},
		"abs with expression": {
			query:    `abs(0 - 2 + 3)`,
			expected: 1.0,
		},
		"abs with input negative number": {
			query:    `abs($)`,
			input:    -10,
			expected: 10.0,
		},
		"abs with input positive number": {
			query:    `abs($)`,
			input:    7.5,
			expected: 7.5,
		},
		"min with two integers": {
			query:    `min(1, 2)`,
			expected: 1.0,
		},
		"min with three decimals": {
			query:    `min(5.5, 3.25, 4.75)`,
			expected: 3.25,
		},
		"min with negative numbers": {
			query:    `min(-1, -5, 0)`,
			expected: -5.0,
		},
		"min with mixed integer and decimal": {
			query:    `min(1.25, 2)`,
			expected: 1.25,
		},
		"min with expressions": {
			query:    `min(2+3, 5-1)`,
			expected: 4.0,
		},
		"min with input data": {
			query:    `min($, 10)`,
			input:    5,
			expected: 5.0,
		},
		"min with input data smaller": {
			query:    `min($, 10)`,
			input:    3,
			expected: 3.0,
		},
		"min with input data larger": {
			query:    `min($, 10)`,
			input:    15,
			expected: 10.0,
		},
		"min with many arguments": {
			query:    `min(10, 5, 8, 3, 12, 1, 7)`,
			expected: 1.0,
		},
		"min with zero": {
			query:    `min(0, 5, -2)`,
			expected: -2.0,
		},
		"min with list expansion": {
			query:    `min(1, 2, [3, 4, 5])`,
			expected: 1.0,
		},
		"min with list expansion where list contains minimum": {
			query:    `min(10, 15, [5, 20, 25])`,
			expected: 5.0,
		},
		"min with multiple lists": {
			query:    `min([1, 3], [2, 4])`,
			expected: 1.0,
		},
		"min with list and mixed numbers": {
			query:    `min(7.5, [3.25, 8.75], 5)`,
			expected: 3.25,
		},
		"min with empty list and numbers": {
			query:    `min(5, 10, [])`,
			expected: 5.0,
		},
		"min with single element list": {
			query:    `min([3.25], 5)`,
			expected: 3.25,
		},
		"min with list containing expressions": {
			query:    `min(10, [2+3, 5-1])`,
			expected: 4.0,
		},
		"max with two integers": {
			query:    `max(1, 2)`,
			expected: 2.0,
		},
		"max with three decimals": {
			query:    `max(5.5, 3.25, 4.75)`,
			expected: 5.5,
		},
		"max with negative numbers": {
			query:    `max(-1, -5, 0)`,
			expected: 0.0,
		},
		"max with mixed integer and decimal": {
			query:    `max(1.25, 2)`,
			expected: 2.0,
		},
		"max with expressions": {
			query:    `max(2+3, 5-1)`,
			expected: 5.0,
		},
		"max with input data": {
			query:    `max($, 10)`,
			input:    5,
			expected: 10.0,
		},
		"max with input data smaller": {
			query:    `max($, 10)`,
			input:    3,
			expected: 10.0,
		},
		"max with input data larger": {
			query:    `max($, 10)`,
			input:    15,
			expected: 15.0,
		},
		"max with many arguments": {
			query:    `max(10, 5, 8, 3, 12, 1, 7)`,
			expected: 12.0,
		},
		"max with zero": {
			query:    `max(0, 5, -2)`,
			expected: 5.0,
		},
		"max with list expansion": {
			query:    `max(1, 2, [3, 4, 5])`,
			expected: 5.0,
		},
		"max with list expansion where list contains maximum": {
			query:    `max(10, 15, [5, 20, 25])`,
			expected: 25.0,
		},
		"max with multiple lists": {
			query:    `max([1, 3], [2, 4])`,
			expected: 4.0,
		},
		"max with list and mixed numbers": {
			query:    `max(7.5, [3.25, 8.75], 5)`,
			expected: 8.75,
		},
		"max with empty list and numbers": {
			query:    `max(5, 10, [])`,
			expected: 10.0,
		},
		"max with single element list": {
			query:    `max([3.25], 5)`,
			expected: 5.0,
		},
		"max with list containing expressions": {
			query:    `max(10, [2+3, 5-1])`,
			expected: 10.0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, tc.input)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_ContainsFunction(t *testing.T) {
	testCases := map[string]struct {
		query    string
		input    any
		expected any
	}{
		"contains with list - positive match": {
			query:    `contains([1, 2, 3], 2)`,
			expected: true,
		},
		"contains with list - negative match": {
			query:    `contains([1, 2, 3], 4)`,
			expected: false,
		},
		"contains with empty list": {
			query:    `contains([], 1)`,
			expected: false,
		},
		"contains with string - positive match": {
			query:    `contains("hello world", "world")`,
			expected: true,
		},
		"contains with string - negative match": {
			query:    `contains("hello world", "xyz")`,
			expected: false,
		},
		"contains with empty string": {
			query:    `contains("", "test")`,
			expected: false,
		},
		"contains with string - substring match": {
			query:    `contains("hello", "ell")`,
			expected: true,
		},
		"contains with map - positive match": {
			query:    `contains({"a": 1, "b": 2}, "a")`,
			expected: true,
		},
		"contains with map - negative match": {
			query:    `contains({"a": 1, "b": 2}, "c")`,
			expected: false,
		},
		"contains with empty map": {
			query:    `contains({}, "key")`,
			expected: false,
		},
		"contains with string and number": {
			query:    `contains("test123", 123)`,
			expected: true,
		},
		"contains with list of mixed types": {
			query:    `contains([1, "hello", true], "hello")`,
			expected: true,
		},
		"contains with boolean in list": {
			query:    `contains([1, "hello", true], true)`,
			expected: true,
		},
		"contains with number in list": {
			query:    `contains([1, 2, 3.5, 4], 3.5)`,
			expected: true,
		},
		"contains with map with numeric key": {
			query:    `contains({"1": "a", "2": "b"}, "1")`,
			expected: true,
		},
		"contains with input data": {
			query:    `contains($, "test")`,
			input:    []any{"hello", "test", "world"},
			expected: true,
		},
		"contains with input string": {
			query:    `contains($, "lo wo")`,
			input:    "hello world",
			expected: true,
		},
		"contains with input map": {
			query:    `contains($, "key1")`,
			input:    map[string]any{"key1": "value1", "key2": "value2"},
			expected: true,
		},
		"round with positive decimal less than .5": {
			query:    `round(3.2)`,
			expected: 3.0,
		},
		"round with positive decimal greater than .5": {
			query:    `round(3.8)`,
			expected: 4.0,
		},
		"round with positive decimal exactly .5": {
			query:    `round(3.5)`,
			expected: 4.0,
		},
		"round with negative decimal less than .5": {
			query:    `round(-3.2)`,
			expected: -3.0,
		},
		"round with negative decimal greater than .5": {
			query:    `round(-3.8)`,
			expected: -4.0,
		},
		"round with negative decimal exactly .5": {
			query:    `round(-3.5)`,
			expected: -3.0,
		},
		"round with zero": {
			query:    `round(0)`,
			expected: 0.0,
		},
		"round with positive integer": {
			query:    `round(5.0)`,
			expected: 5.0,
		},
		"round with negative integer": {
			query:    `round(-5.0)`,
			expected: -5.0,
		},
		"round with small positive decimal": {
			query:    `round(0.1)`,
			expected: 0.0,
		},
		"round with small negative decimal": {
			query:    `round(-0.1)`,
			expected: 0.0,
		},
		"round with expression": {
			query:    `round(2.5 + 1.2)`,
			expected: 4.0,
		},
		"round with complex expression": {
			query:    `round((3.7 + 2.3) * 1.5)`,
			expected: 9.0,
		},
		"round with input data": {
			query:    `round($)`,
			input:    2.7,
			expected: 3.0,
		},
		"round with input negative": {
			query:    `round($)`,
			input:    -2.7,
			expected: -3.0,
		},
		"round with input half": {
			query:    `round($)`,
			input:    2.5,
			expected: 3.0,
		},
		"round with input negative half": {
			query:    `round($)`,
			input:    -2.5,
			expected: -2.0,
		},
		"round with very small decimal": {
			query:    `round(0.0001)`,
			expected: 0.0,
		},
		"round with large number": {
			query:    `round(123456789.5)`,
			expected: 123456790.0,
		},
		"round with negative large number": {
			query:    `round(-123456789.5)`,
			expected: -123456789.0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, tc.input)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_Function_Floor(t *testing.T) {
	testCases := map[string]struct {
		query    string
		input    any
		expected float64
	}{
		"floor with positive decimal less than .5": {
			query:    `floor(3.2)`,
			expected: 3.0,
		},
		"floor with positive decimal greater than .5": {
			query:    `floor(3.8)`,
			expected: 3.0,
		},
		"floor with negative decimal less than .5": {
			query:    `floor(-3.2)`,
			expected: -4.0,
		},
		"floor with negative decimal greater than .5": {
			query:    `floor(-3.8)`,
			expected: -4.0,
		},
		"floor with zero": {
			query:    `floor(0)`,
			expected: 0.0,
		},
		"floor with positive integer": {
			query:    `floor(5.0)`,
			expected: 5.0,
		},
		"floor with negative integer": {
			query:    `floor(-5.0)`,
			expected: -5.0,
		},
		"floor with small positive decimal": {
			query:    `floor(0.1)`,
			expected: 0.0,
		},
		"floor with small negative decimal": {
			query:    `floor(-0.1)`,
			expected: -1.0,
		},
		"floor with expression": {
			query:    `floor(2.5 + 1.2)`,
			expected: 3.0,
		},
		"floor with complex expression": {
			query:    `floor((3.7 + 2.3) * 1.5)`,
			expected: 9.0,
		},
		"floor with input data": {
			query:    `floor($)`,
			input:    2.7,
			expected: 2.0,
		},
		"floor with input negative": {
			query:    `floor($)`,
			input:    -2.7,
			expected: -3.0,
		},
		"floor with very small decimal": {
			query:    `floor(0.0001)`,
			expected: 0.0,
		},
		"floor with large number": {
			query:    `floor(123456789.9)`,
			expected: 123456789.0,
		},
		"floor with negative large number": {
			query:    `floor(-123456789.1)`,
			expected: -123456790.0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, tc.input)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_Function_Ceil(t *testing.T) {
	testCases := map[string]struct {
		query    string
		input    any
		expected float64
	}{
		"ceil with positive decimal less than .5": {
			query:    `ceil(3.2)`,
			expected: 4.0,
		},
		"ceil with positive decimal greater than .5": {
			query:    `ceil(3.8)`,
			expected: 4.0,
		},
		"ceil with negative decimal less than .5": {
			query:    `ceil(-3.2)`,
			expected: -3.0,
		},
		"ceil with negative decimal greater than .5": {
			query:    `ceil(-3.8)`,
			expected: -3.0,
		},
		"ceil with zero": {
			query:    `ceil(0)`,
			expected: 0.0,
		},
		"ceil with positive integer": {
			query:    `ceil(5.0)`,
			expected: 5.0,
		},
		"ceil with negative integer": {
			query:    `ceil(-5.0)`,
			expected: -5.0,
		},
		"ceil with small positive decimal": {
			query:    `ceil(0.1)`,
			expected: 1.0,
		},
		"ceil with small negative decimal": {
			query:    `ceil(-0.1)`,
			expected: 0.0,
		},
		"ceil with expression": {
			query:    `ceil(2.5 + 1.2)`,
			expected: 4.0,
		},
		"ceil with complex expression": {
			query:    `ceil((3.7 + 2.3) * 1.5)`,
			expected: 9.0,
		},
		"ceil with input data": {
			query:    `ceil($)`,
			input:    2.7,
			expected: 3.0,
		},
		"ceil with input negative": {
			query:    `ceil($)`,
			input:    -2.7,
			expected: -2.0,
		},
		"ceil with very small decimal": {
			query:    `ceil(0.0001)`,
			expected: 1.0,
		},
		"ceil with large number": {
			query:    `ceil(123456789.9)`,
			expected: 123456790.0,
		},
		"ceil with negative large number": {
			query:    `ceil(-123456789.1)`,
			expected: -123456789.0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, tc.input)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}

func Test_Eval_Function_Errors(t *testing.T) {
	testCases := map[string]struct {
		query         string
		input         any
		expectedError error
	}{
		"undefined function": {
			query:         `unknown("test")`,
			expectedError: runtime.ErrUndefinedFunction,
		},
		"len with too many arguments": {
			query:         `len("hello", "world")`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"len with no arguments": {
			query:         `len()`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"len with number": {
			query:         `len(123)`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"len with boolean": {
			query:         `len(true)`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"len with input number": {
			query:         `len($)`,
			input:         42,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"len with input boolean": {
			query:         `len($)`,
			input:         true,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"contains with too many arguments": {
			query:         `contains([1, 2, 3], 2, "extra")`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"contains with too few arguments": {
			query:         `contains([1, 2, 3])`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"contains with no arguments": {
			query:         `contains()`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"contains with number as first argument": {
			query:         `contains(123, 2)`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"contains with boolean as first argument": {
			query:         `contains(true, "test")`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"contains with input number as first argument": {
			query:         `contains($, "test")`,
			input:         42,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"abs with no arguments": {
			query:         `abs()`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"abs with too many arguments": {
			query:         `abs(1, 2)`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"abs with string argument": {
			query:         `abs("hello")`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"abs with boolean argument": {
			query:         `abs(true)`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"abs with list argument": {
			query:         `abs([1, 2, 3])`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"abs with map argument": {
			query:         `abs({"a": 1})`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"abs with input string": {
			query:         `abs($)`,
			input:         "hello",
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"abs with input boolean": {
			query:         `abs($)`,
			input:         true,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"min with no arguments": {
			query:         `min()`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"min with one argument": {
			query:         `min(1)`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"min with string argument": {
			query:         `min("hello", 1)`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"min with string argument second": {
			query:         `min(1, "world")`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"min with boolean argument": {
			query:         `min(true, 1)`,
			expectedError: runtime.ErrInvalidArgumentType,
		},

		"min with map argument": {
			query:         `min({"a": 1}, 1)`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"min with input string": {
			query:         `min($, 1)`,
			input:         "hello",
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"min with input boolean": {
			query:         `min($, 1)`,
			input:         true,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"min with mixed valid and invalid arguments": {
			query:         `min(1, "invalid", 2)`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"min with list containing non-numeric elements": {
			query:         `min(1, [2, "hello", 3])`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"min with list expansion resulting in single argument": {
			query:         `min([5])`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"min with empty list only": {
			query:         `min([])`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"min with single argument and empty list": {
			query:         `min(5, [])`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"max with no arguments": {
			query:         `max()`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"max with one argument": {
			query:         `max(1)`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"max with string argument": {
			query:         `max("hello", 1)`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"max with string argument second": {
			query:         `max(1, "world")`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"max with boolean argument": {
			query:         `max(true, 1)`,
			expectedError: runtime.ErrInvalidArgumentType,
		},

		"max with map argument": {
			query:         `max({"a": 1}, 1)`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"max with input string": {
			query:         `max($, 1)`,
			input:         "hello",
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"max with input boolean": {
			query:         `max($, 1)`,
			input:         true,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"max with mixed valid and invalid arguments": {
			query:         `max(1, "invalid", 2)`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"max with list containing non-numeric elements": {
			query:         `max(1, [2, "hello", 3])`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"max with list expansion resulting in single argument": {
			query:         `max([5])`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"max with empty list only": {
			query:         `max([])`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"max with single argument and empty list": {
			query:         `max(5, [])`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"round with no arguments": {
			query:         `round()`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"round with too many arguments": {
			query:         `round(1.5, 2.5, 3.5)`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"round with string argument": {
			query:         `round("hello")`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"round with boolean argument": {
			query:         `round(true)`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"round with list argument": {
			query:         `round([1, 2, 3])`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"round with map argument": {
			query:         `round({"a": 1})`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"round with input string": {
			query:         `round($)`,
			input:         "hello",
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"round with input boolean": {
			query:         `round($)`,
			input:         true,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"round with input list": {
			query:         `round($)`,
			input:         []any{1, 2, 3},
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"round with input map": {
			query:         `round($)`,
			input:         map[string]any{"a": 1},
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"floor with no arguments": {
			query:         `floor()`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"floor with too many arguments": {
			query:         `floor(1.5, 2.5)`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"floor with string argument": {
			query:         `floor("hello")`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"floor with boolean argument": {
			query:         `floor(true)`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"floor with list argument": {
			query:         `floor([1, 2, 3])`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"floor with map argument": {
			query:         `floor({"a": 1})`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"floor with input string": {
			query:         `floor($)`,
			input:         "hello",
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"floor with input boolean": {
			query:         `floor($)`,
			input:         true,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"floor with input list": {
			query:         `floor($)`,
			input:         []any{1, 2, 3},
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"floor with input map": {
			query:         `floor($)`,
			input:         map[string]any{"a": 1},
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"ceil with no arguments": {
			query:         `ceil()`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"ceil with too many arguments": {
			query:         `ceil(1.5, 2.5)`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"ceil with string argument": {
			query:         `ceil("hello")`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"ceil with boolean argument": {
			query:         `ceil(true)`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"ceil with list argument": {
			query:         `ceil([1, 2, 3])`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"ceil with map argument": {
			query:         `ceil({"a": 1})`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"ceil with input string": {
			query:         `ceil($)`,
			input:         "hello",
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"ceil with input boolean": {
			query:         `ceil($)`,
			input:         true,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"ceil with input list": {
			query:         `ceil($)`,
			input:         []any{1, 2, 3},
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"ceil with input map": {
			query:         `ceil($)`,
			input:         map[string]any{"a": 1},
			expectedError: runtime.ErrInvalidArgumentType,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			_, err = runtime.Eval(expr, tc.input)
			require.Error(t, err, "Expected runtime error")
			require.ErrorIs(t, err, tc.expectedError, "Error should be of expected type")
		})
	}
}

func Test_Eval_FilterFunction(t *testing.T) {
	testCases := map[string]struct {
		query       string
		input       any
		expectedLen int
		verifyFunc  func(t *testing.T, result parser.ExprList)
	}{
		"filter numbers greater than 3": {
			query:       `filter([1, 2, 3, 4, 5], _ > 3)`,
			expectedLen: 2,
			verifyFunc: func(t *testing.T, result parser.ExprList) {
				require.Len(t, result.Values, 2)
				num1, ok := result.Values[0].(parser.ExprNumber)
				require.True(t, ok)
				float1, _ := num1.Value.Float64()
				require.Equal(t, 4.0, float1)
				num2, ok := result.Values[1].(parser.ExprNumber)
				require.True(t, ok)
				float2, _ := num2.Value.Float64()
				require.Equal(t, 5.0, float2)
			},
		},
		"filter exact string match": {
			query:       `filter(["andrew", "alex", "anthony"], _ == "alex")`,
			expectedLen: 1,
			verifyFunc: func(t *testing.T, result parser.ExprList) {
				require.Len(t, result.Values, 1)
				str, ok := result.Values[0].(parser.ExprString)
				require.True(t, ok)
				require.Equal(t, "alex", str.Value)
			},
		},
		"filter strings with length 5": {
			query:       `filter(["hello", "world", "something"], len(_) == 5)`,
			expectedLen: 2,
			verifyFunc: func(t *testing.T, result parser.ExprList) {
				require.Len(t, result.Values, 2)
				str1, ok := result.Values[0].(parser.ExprString)
				require.True(t, ok)
				require.Equal(t, "hello", str1.Value)
				str2, ok := result.Values[1].(parser.ExprString)
				require.True(t, ok)
				require.Equal(t, "world", str2.Value)
			},
		},
		"filter with greater than or equal": {
			query:       `filter([1, 2, 3, 4, 5], _ >= 3)`,
			expectedLen: 3,
			verifyFunc: func(t *testing.T, result parser.ExprList) {
				require.Len(t, result.Values, 3)
				num1, ok := result.Values[0].(parser.ExprNumber)
				require.True(t, ok)
				float1, _ := num1.Value.Float64()
				require.Equal(t, 3.0, float1)
				num2, ok := result.Values[1].(parser.ExprNumber)
				require.True(t, ok)
				float2, _ := num2.Value.Float64()
				require.Equal(t, 4.0, float2)
				num3, ok := result.Values[2].(parser.ExprNumber)
				require.True(t, ok)
				float3, _ := num3.Value.Float64()
				require.Equal(t, 5.0, float3)
			},
		},
		"filter with less than": {
			query:       `filter([1, 2, 3, 4, 5], _ < 3)`,
			expectedLen: 2,
			verifyFunc: func(t *testing.T, result parser.ExprList) {
				require.Len(t, result.Values, 2)
				num1, ok := result.Values[0].(parser.ExprNumber)
				require.True(t, ok)
				float1, _ := num1.Value.Float64()
				require.Equal(t, 1.0, float1)
				num2, ok := result.Values[1].(parser.ExprNumber)
				require.True(t, ok)
				float2, _ := num2.Value.Float64()
				require.Equal(t, 2.0, float2)
			},
		},
		"filter with less than or equal": {
			query:       `filter([1, 2, 3, 4, 5], _ <= 2)`,
			expectedLen: 2,
			verifyFunc: func(t *testing.T, result parser.ExprList) {
				require.Len(t, result.Values, 2)
				num1, ok := result.Values[0].(parser.ExprNumber)
				require.True(t, ok)
				float1, _ := num1.Value.Float64()
				require.Equal(t, 1.0, float1)
				num2, ok := result.Values[1].(parser.ExprNumber)
				require.True(t, ok)
				float2, _ := num2.Value.Float64()
				require.Equal(t, 2.0, float2)
			},
		},
		"filter with not equals": {
			query:       `filter([1, 2, 3, 4, 5], _ != 3)`,
			expectedLen: 4,
			verifyFunc: func(t *testing.T, result parser.ExprList) {
				require.Len(t, result.Values, 4)
				num1, ok := result.Values[0].(parser.ExprNumber)
				require.True(t, ok)
				float1, _ := num1.Value.Float64()
				require.Equal(t, 1.0, float1)
				num2, ok := result.Values[1].(parser.ExprNumber)
				require.True(t, ok)
				float2, _ := num2.Value.Float64()
				require.Equal(t, 2.0, float2)
				num3, ok := result.Values[2].(parser.ExprNumber)
				require.True(t, ok)
				float3, _ := num3.Value.Float64()
				require.Equal(t, 4.0, float3)
				num4, ok := result.Values[3].(parser.ExprNumber)
				require.True(t, ok)
				float4, _ := num4.Value.Float64()
				require.Equal(t, 5.0, float4)
			},
		},
		"filter with boolean condition": {
			query:       `filter([true, false, true], _)`,
			expectedLen: 2,
			verifyFunc: func(t *testing.T, result parser.ExprList) {
				require.Len(t, result.Values, 2)
				bool1, ok := result.Values[0].(parser.ExprBoolean)
				require.True(t, ok)
				require.True(t, bool1.Value)
				bool2, ok := result.Values[1].(parser.ExprBoolean)
				require.True(t, ok)
				require.True(t, bool2.Value)
			},
		},
		"filter empty result": {
			query:       `filter([1, 2, 3, 4, 5], _ > 10)`,
			expectedLen: 0,
			verifyFunc: func(t *testing.T, result parser.ExprList) {
				require.Len(t, result.Values, 0)
			},
		},
		"filter entire list matches": {
			query:       `filter([1, 2, 3, 4, 5], _ >= 1)`,
			expectedLen: 5,
			verifyFunc: func(t *testing.T, result parser.ExprList) {
				require.Len(t, result.Values, 5)
			},
		},
		"filter with function in condition": {
			query:       `filter(["a", "hello"], len(_) >= 1)`,
			expectedLen: 2,
			verifyFunc: func(t *testing.T, result parser.ExprList) {
				require.Len(t, result.Values, 2)
				str1, ok := result.Values[0].(parser.ExprString)
				require.True(t, ok)
				require.Equal(t, "a", str1.Value)
				str2, ok := result.Values[1].(parser.ExprString)
				require.True(t, ok)
				require.Equal(t, "hello", str2.Value)
			},
		},
		"filter with function in condition - complex": {
			query:       `filter(["a", "hello", "ab"], len(_) == 2)`,
			expectedLen: 1,
			verifyFunc: func(t *testing.T, result parser.ExprList) {
				require.Len(t, result.Values, 1)
				str, ok := result.Values[0].(parser.ExprString)
				require.True(t, ok)
				require.Equal(t, "ab", str.Value)
			},
		},
		"filter empty list": {
			query:       `filter([], _ > 0)`,
			expectedLen: 0,
			verifyFunc: func(t *testing.T, result parser.ExprList) {
				require.Len(t, result.Values, 0)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, tc.input)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			// Check that the result is an ExprList
			resultList, ok := resultDecoded.(parser.ExprList)
			require.True(t, ok, "Expected ExprList, got %T", resultDecoded)

			// Verify expected length
			require.Equal(t, tc.expectedLen, len(resultList.Values), "List length mismatch")

			// Run specific verification function
			tc.verifyFunc(t, resultList)
		})
	}
}

func Test_Eval_FilterFunction_Errors(t *testing.T) {
	testCases := map[string]struct {
		query         string
		input         any
		expectedError error
	}{
		"filter with non-list first argument": {
			query:         `filter(5, _ > 3)`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"filter with too few arguments": {
			query:         `filter([1, 2, 3])`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"filter with too many arguments": {
			query:         `filter([1, 2, 3], _ > 1, "extra")`,
			expectedError: runtime.ErrInvalidArgumentCount,
		},
		"filter with non-boolean expression": {
			query:         `filter([1, 2, 3], _ + 1)`, // Expression doesn't return boolean
			expectedError: runtime.ErrInvalidArgumentType,
		},
		"filter with string instead of list": {
			query:         `filter("hello", _ == "h")`,
			expectedError: runtime.ErrInvalidArgumentType,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			_, err = runtime.Eval(expr, tc.input)
			require.Error(t, err, "Expected runtime error")
			require.ErrorIs(t, err, tc.expectedError, "Error should be of expected type")
		})
	}
}
func Test_Eval_ListSlice(t *testing.T) {
	testCases := map[string]struct {
		query    string
		expected []any
	}{
		"slice with start and end": {
			query:    "[1, 2, 3, 4, 5][1:3]",
			expected: []any{2.0, 3.0},
		},
		"slice with only start (omitted end)": {
			query:    "[1, 2, 3][1:]",
			expected: []any{2.0, 3.0},
		},
		"slice with only end (omitted start)": {
			query:    "[1, 2, 3][:2]",
			expected: []any{1.0, 2.0},
		},
		"slice with no bounds (full slice)": {
			query:    "[1, 2, 3][:]",
			expected: []any{1.0, 2.0, 3.0},
		},
		"slice with complex start expression": {
			query:    "[1, 2, 3, 4, 5][1+1:4]",
			expected: []any{3.0, 4.0},
		},
		"slice with complex end expression": {
			query:    "[1, 2, 3, 4, 5][1:4-1]",
			expected: []any{2.0, 3.0},
		},
		"slice with complex start and end expressions": {
			query:    "[1, 2, 3, 4, 5][1+1:4-1]",
			expected: []any{3.0},
		},
		"slice with negative start index": {
			query:    "[1, 2, 3, 4, 5][-2:]",
			expected: []any{4.0, 5.0},
		},
		"slice with negative end index": {
			query:    "[1, 2, 3, 4, 5][:-1]",
			expected: []any{1.0, 2.0, 3.0, 4.0},
		},
		"slice with negative start and end indices": {
			query:    "[1, 2, 3, 4, 5][-3:-1]",
			expected: []any{3.0, 4.0},
		},
		"slice with start and end equal": {
			query:    "[1, 2, 3][1:1]",
			expected: []any{}, // Should return empty list
		},
		"slice with start greater than end": {
			query:    "[1, 2, 3][2:1]",
			expected: []any{}, // Should return empty list
		},
		"slice with out of bounds indices": {
			query:    "[1, 2, 3][1:10]",
			expected: []any{2.0, 3.0}, // Should return from start to end of list
		},
		"slice with negative out of bounds indices": {
			query:    "[1, 2, 3][-10:2]",
			expected: []any{1.0, 2.0}, // Should clamp to start of list
		},
		"empty slice result": {
			query:    "[][:]",
			expected: []any{}, // Should return empty list
		},
		"chained slicing and indexing": {
			query:    "[[1, 2], [3, 4], [5, 6]][1:][0][1:]",
			expected: []any{4.0},
		},
		"slice with mixed types": {
			query:    "[1, true, \"hello\", 3.14][:3]",
			expected: []any{1.0, true, "hello"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, nil)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			// Check that the result is an ExprList
			resultList, ok := resultDecoded.(parser.ExprList)
			require.True(t, ok, "Expected ExprList, got %T", resultDecoded)

			// Compare lengths
			require.Equal(t, len(tc.expected), len(resultList.Values), "List length mismatch")

			// Compare each element
			for i, expectedValue := range tc.expected {
				valueDecoded, err := resultList.Values[i].Decode()
				require.NoError(t, err, "Failed to decode list element %d", i)
				require.Equal(t, expectedValue, valueDecoded, "Element %d does not match expected value", i)
			}
		})
	}
}

func Test_Eval_StringSlice(t *testing.T) {
	testCases := map[string]struct {
		query    string
		expected string
	}{
		"slice string with start and end": {
			query:    `"hello"[1:4]`,
			expected: "ell",
		},
		"slice string with only start (omitted end)": {
			query:    `"hello"[1:]`,
			expected: "ello",
		},
		"slice string with only end (omitted start)": {
			query:    `"hello"[:3]`,
			expected: "hel",
		},
		"slice string with no bounds (full slice)": {
			query:    `"hello"[:]`,
			expected: "hello",
		},
		"slice string with negative start": {
			query:    `"world"[-2:]`,
			expected: "ld",
		},
		"slice string with negative end": {
			query:    `"world"[:-1]`,
			expected: "worl",
		},
		"slice string with negative start and end": {
			query:    `"hello"[-4:-1]`,
			expected: "ell",
		},
		"slice string with start and end equal": {
			query:    `"hello"[2:2]`,
			expected: "",
		},
		"slice string with start greater than end": {
			query:    `"hello"[3:2]`,
			expected: "",
		},
		"slice empty string": {
			query:    `""[:]`,
			expected: "",
		},
		"slice string with out of bounds indices": {
			query:    `"hi"[0:10]`,
			expected: "hi",
		},
		"slice string with negative out of bounds indices": {
			query:    `"hello"[-10:3]`,
			expected: "hel",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, nil)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			// Check that the result is an ExprString
			resultString, ok := resultDecoded.(string)
			require.True(t, ok, "Expected ExprString, got %T", resultDecoded)

			require.Equal(t, tc.expected, resultString, "String value does not match expected value")
		})
	}
}
