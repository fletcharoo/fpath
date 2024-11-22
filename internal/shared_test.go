package internal_test

import (
	"reflect"
	"testing"

	"github.com/fletcharoo/fpath/internal"
)

func Test_LookupPath(t *testing.T) {
	type testStruct struct {
		Field string
	}

	testCases := map[string]struct {
		data     any
		path     []string
		expected any
	}{
		"map": {
			data: map[string]any{
				"hello": "world",
			},
			path:     []string{"hello"},
			expected: "world",
		},
		"map nested": {
			data: map[string]any{
				"hello": map[string]any{
					"world": "something",
				},
			},
			path:     []string{"hello", "world"},
			expected: "something",
		},
		"slice": {
			data:     []any{"hello", "world", "something"},
			path:     []string{"2"},
			expected: "something",
		},
		"struct": {
			data: testStruct{
				Field: "hello",
			},
			path:     []string{"Field"},
			expected: "hello",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result, err := internal.LookupPath(tc.data, tc.path)
			if err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("Unexpected result\nExpected: %v\nActual: %v", tc.expected, result)
			}
		})
	}
}
