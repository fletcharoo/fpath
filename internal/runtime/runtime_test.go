package runtime_test

import (
	"testing"

	"github.com/fletcharoo/fpath/internal/lexer"
	"github.com/fletcharoo/fpath/internal/parser"
	"github.com/fletcharoo/fpath/internal/runtime"
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
			expected: 4.0,
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
			require.Contains(t, err.Error(), "incompatible types", "Error message should mention incompatible types")
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
			require.Contains(t, err.Error(), "division by zero", "Error message should mention division by zero")
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
			require.Contains(t, err.Error(), "incompatible types", "Error message should mention incompatible types")
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
			require.Contains(t, err.Error(), "incompatible types", "Error message should mention incompatible types")
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
			require.Contains(t, err.Error(), "incompatible types", "Error message should mention incompatible types")
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
			require.Contains(t, err.Error(), "incompatible types", "Error message should mention incompatible types")
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
			require.Contains(t, err.Error(), "incompatible types", "Error message should mention incompatible types")
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
			require.Contains(t, err.Error(), "incompatible types", "Error message should mention incompatible types")
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
			require.Contains(t, err.Error(), "AND operation requires boolean expressions", "Error message should mention AND operation requires boolean expressions")
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
