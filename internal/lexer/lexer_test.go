package lexer

import (
	"errors"
	"fmt"
	"io"
	"testing"
)

func _tokensMatch(expected, actual Token) (err error) {
	if expected.Type != actual.Type {
		err = fmt.Errorf("Unexpected type\nExpected: %s\nActual: %s", TokenTypeString[expected.Type], TokenTypeString[actual.Type])
		return
	}

	if expected.Value != actual.Value {
		err = fmt.Errorf("Unexpected value\nExpected: %s\nActual: %s", expected.Value, actual.Value)
	}

	return nil
}

func Test_isLabelRune(t *testing.T) {
	testCases := map[string]struct {
		r        rune
		expected bool
	}{
		"1": {
			r:        '1',
			expected: true,
		},
		"f": {
			r:        'f',
			expected: true,
		},
		"_": {
			r:        '_',
			expected: true,
		},
		"$": {
			r:        '$',
			expected: false,
		},
		"-": {
			r:        '-',
			expected: false,
		},
		"Tab": {
			r:        '\t',
			expected: false,
		},
		"Space": {
			r:        ' ',
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := isLabelRune(tc.r)
			if result != tc.expected {
				t.Fatalf("Unexpected result\nExpected: %v\nActual: %v", tc.expected, result)
			}
		})
	}
}

func Test_Lexer_getRune(t *testing.T) {
	input := "hello world"
	expectedRunes := []rune{'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd'}
	lexer := New(input)

	for i, expected := range expectedRunes {
		r, err := lexer.getRune()

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}

		if r != expected {
			t.Fatalf("Unexpected rune at position %d\nExpected: %q\n Actual: %q", i, expected, r)
		}
	}
}

func Test_Lexer_getRune_EOF(t *testing.T) {
	input := "h"
	expected := 'h'
	lexer := New(input)

	r, err := lexer.getRune()

	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if r != expected {
		t.Fatalf("Unexpected rune\nExpected: %q\nActual: %q", expected, r)
	}

	for i := 0; i < 100; i++ {
		if _, err = lexer.getRune(); err != io.EOF {
			t.Fatalf("Unexpected error: %s", err)
		}
	}
}

func Test_Lexer_peekRune(t *testing.T) {
	input := "hello world"
	expected := 'h'
	lexer := New(input)

	for i := 0; i < 10; i++ {
		r, err := lexer.peekRune()

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}

		if r != expected {
			t.Fatalf("Unexpected rune\nExpected: %q\nActual: %q", expected, r)
		}
	}
}

func Test_Lexer_peekRune_EOF(t *testing.T) {
	input := "h"
	expected := 'h'
	lexer := New(input)

	r, err := lexer.getRune()

	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if r != expected {
		t.Fatalf("Unexpected rune\nExpected: %q\nActual: %q", expected, r)
	}

	for i := 0; i < 100; i++ {
		if _, err = lexer.peekRune(); err != io.EOF {
			t.Fatalf("Unexpected error: %s", err)
		}
	}
}

func Test_Lexer_getToken(t *testing.T) {
	testCases := map[string]struct {
		input          string
		expectedTokens []Token
	}{
		"Whitespace": {
			input: "  123  ",
			expectedTokens: []Token{
				{Type: TokenType_Number, Value: "123"},
			},
		},
		"Number": {
			input: "123",
			expectedTokens: []Token{
				{Type: TokenType_Number, Value: "123"},
			},
		},
		"Label": {
			input: "fletcher",
			expectedTokens: []Token{
				{Type: TokenType_Label, Value: "fletcher"},
			},
		},
		"StringLiteral": {
			input: `"hello world"`,
			expectedTokens: []Token{
				{Type: TokenType_StringLiteral, Value: "hello world"},
			},
		},
		"Plus": {
			input: "+",
			expectedTokens: []Token{
				{Type: TokenType_Plus},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lexer := New(tc.input)

			for _, expected := range tc.expectedTokens {
				tok, err := lexer.GetToken()

				if err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}

				if err := _tokensMatch(expected, tok); err != nil {
					t.Fatalf("Unexpected result: %s", err)
				}
			}
		})
	}
}

func Test_Lexer_getToken_EOF(t *testing.T) {
	input := "  123  "
	expected := Token{
		Type:  TokenType_Number,
		Value: "123",
	}
	lexer := New(input)

	tok, err := lexer.GetToken()

	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if err := _tokensMatch(expected, tok); err != nil {
		t.Fatalf("Unexpected result: %s", err)
	}

	for i := 0; i < 100; i++ {
		if _, err := lexer.GetToken(); err != io.EOF {
			t.Fatalf("Unexpected error: %s", err)
		}
	}
}

func Test_Lexer_getToken_InvalidRune(t *testing.T) {
	input := "  123  `"
	expected := Token{
		Type:  TokenType_Number,
		Value: "123",
	}
	lexer := New(input)

	tok, err := lexer.GetToken()

	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if err := _tokensMatch(expected, tok); err != nil {
		t.Fatalf("Unexpected result: %s", err)
	}

	_, err = lexer.GetToken()

	if err == nil {
		t.Fatalf("Error expected but not returned")
	}

	if !errors.Is(err, errInvalidRune) {
		t.Fatalf("Unexpected result\nExpected: %s\nActual: %s", errInvalidRune, err)
	}
}

func Test_Lexer_peekToken(t *testing.T) {
	input := "123 +"
	firstExpected := Token{
		Type:  TokenType_Number,
		Value: "123",
	}
	secondExpected := Token{
		Type: TokenType_Plus,
	}
	lexer := New(input)
	shouldBreak := false

	for i := 0; i < 100; i++ {
		t.Run("First peek", func(t *testing.T) {
			tok, err := lexer.PeekToken()

			if err != nil {
				shouldBreak = true
				t.Fatalf("Unexpected error: %s", err)
			}

			if err := _tokensMatch(firstExpected, tok); err != nil {
				shouldBreak = true
				t.Fatalf("Unexpected result: %s", err)
			}
		})

		if shouldBreak {
			break
		}
	}

	t.Run("First get", func(t *testing.T) {
		tok, err := lexer.GetToken()

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}

		if err := _tokensMatch(firstExpected, tok); err != nil {
			t.Fatalf("Unexpected result: %s", err)
		}
	})

	for i := 0; i < 100; i++ {
		t.Run("Second peek", func(t *testing.T) {
			tok, err := lexer.PeekToken()

			if err != nil {
				shouldBreak = true
				t.Fatalf("Unexpected error: %s", err)
			}

			if err := _tokensMatch(secondExpected, tok); err != nil {
				shouldBreak = true
				t.Fatalf("Unexpected result: %s", err)
			}
		})

		if shouldBreak {
			break
		}
	}

	t.Run("Second get", func(t *testing.T) {
		tok, err := lexer.GetToken()

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}

		if err := _tokensMatch(secondExpected, tok); err != nil {
			t.Fatalf("Unexpected result: %s", err)
		}
	})
}

func Test_Lexer_getTokenStringLiteral_UnexpectedEOF(t *testing.T) {
	input := `"hello `
	lexer := New(input)

	_, err := lexer.GetToken()

	if err == nil {
		t.Fatalf("Error expected but not returned")
	}

	if !errors.Is(err, errUnexpectedEOF) {
		t.Fatalf("Unexpected result\nExpected: %s\nActual: %s", errUnexpectedEOF, err)
	}
}
