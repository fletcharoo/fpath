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
