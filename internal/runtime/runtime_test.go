package runtime_test

import (
	"os"
	"testing"

	"github.com/fletcharoo/fpath/internal/lexer"
	"github.com/fletcharoo/fpath/internal/parser"
	"github.com/fletcharoo/fpath/internal/runtime"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestMain(m *testing.M) {
	r := m.Run()
	snaps.Clean(m, snaps.CleanOpts{Sort: true})
	os.Exit(r)
}

func Test_Eval(t *testing.T) {
	testCases := map[string]struct {
		query string
		input any
	}{
		"number": {
			query: "2",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lex := lexer.New(tc.query)
			expr, err := parser.New(lex).Parse()
			if err != nil {
				t.Fatalf("Unexpected parser error: %s", err)
			}

			result, err := runtime.Eval(expr, err)
			if err != nil {
				t.Fatalf("Unexpected runtime error: %s", err)
			}

			resultDecoded, err := result.Decode()
			if err != nil {
				t.Fatalf("Failed to decode result: %s", err)
			}

			snaps.MatchSnapshot(t, resultDecoded)
		})
	}
}
