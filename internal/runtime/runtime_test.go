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
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			require.NoError(t, err, "Unexpected parser error")

			result, err := runtime.Eval(expr, err)
			require.NoError(t, err, "Unexpected runtime error")

			resultDecoded, err := result.Decode()
			require.NoError(t, err, "Failed to decode result")

			require.Equal(t, tc.expected, resultDecoded, "Result does not match expected value")
		})
	}
}
