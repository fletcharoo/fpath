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
		"Equals": {
			input: "==",
			expectedTokens: []Token{
				{Type: TokenType_Equals},
			},
		},
		"NotEquals": {
			input: "!=",
			expectedTokens: []Token{
				{Type: TokenType_NotEquals},
			},
		},
		"LessThan": {
			input: "<",
			expectedTokens: []Token{
				{Type: TokenType_LessThan},
			},
		},
		"GreaterThanOrEqual": {
			input: ">=",
			expectedTokens: []Token{
				{Type: TokenType_GreaterThanOrEqual},
			},
		},
		"And": {
			input: "&&",
			expectedTokens: []Token{
				{Type: TokenType_And},
			},
		},
		"Or": {
			input: "||",
			expectedTokens: []Token{
				{Type: TokenType_Or},
			},
		},
		"LeftBracket": {
			input: "[",
			expectedTokens: []Token{
				{Type: TokenType_LeftBracket},
			},
		},
		"RightBracket": {
			input: "]",
			expectedTokens: []Token{
				{Type: TokenType_RightBracket},
			},
		},
		"Comma": {
			input: ",",
			expectedTokens: []Token{
				{Type: TokenType_Comma},
			},
		},
		"LeftBrace": {
			input: "{",
			expectedTokens: []Token{
				{Type: TokenType_LeftBrace},
			},
		},
		"RightBrace": {
			input: "}",
			expectedTokens: []Token{
				{Type: TokenType_RightBrace},
			},
		},
		"Colon": {
			input: ":",
			expectedTokens: []Token{
				{Type: TokenType_Colon},
			},
		},
		"Dollar": {
			input: "$",
			expectedTokens: []Token{
				{Type: TokenType_Dollar},
			},
		},
		"Question": {
			input: "?",
			expectedTokens: []Token{
				{Type: TokenType_Question},
			},
		},
		"Modulo": {
			input: "%",
			expectedTokens: []Token{
				{Type: TokenType_Modulo},
			},
		},
		"Caret": {
			input: "^",
			expectedTokens: []Token{
				{Type: TokenType_Caret},
			},
		},
		"IntegerDivision": {
			input: "//",
			expectedTokens: []Token{
				{Type: TokenType_IntegerDivision},
			},
		},
		"SlashThenSlash": {
			input: "/ /",
			expectedTokens: []Token{
				{Type: TokenType_Slash},
				{Type: TokenType_Slash},
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
	testCases := map[string]struct {
		input string
	}{
		"backtick": {
			input: "  123  `",
		},
		"single equals": {
			input: "  123  =",
		},
		"single exclamation": {
			input: "  123  !",
		},
		"single ampersand": {
			input: "  123  &",
		},
		"single pipe": {
			input: "  123  |",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			expected := Token{
				Type:  TokenType_Number,
				Value: "123",
			}
			lexer := New(tc.input)

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
		})
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

func Test_Token_String(t *testing.T) {
	testCases := map[string]struct {
		token    Token
		expected string
	}{
		"Number": {
			token:    Token{Type: TokenType_Number, Value: "123"},
			expected: "Number",
		},
		"Label": {
			token:    Token{Type: TokenType_Label, Value: "test"},
			expected: "Label",
		},
		"Undefined": {
			token:    Token{Type: TokenType_Undefined, Value: ""},
			expected: "Undefined",
		},
		"Invalid type": {
			token:    Token{Type: 999, Value: ""},
			expected: "Undefined",
		},
		"IntegerDivision": {
			token:    Token{Type: TokenType_IntegerDivision, Value: ""},
			expected: "IntegerDivision",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := tc.token.String()
			if result != tc.expected {
				t.Fatalf("Unexpected result\nExpected: %s\nActual: %s", tc.expected, result)
			}
		})
	}
}

func Test_New(t *testing.T) {
	input := "test input"
	lexer := New(input)

	if lexer.input == nil {
		t.Fatalf("Expected input to be initialized")
	}

	if string(lexer.input) != input {
		t.Fatalf("Unexpected input\nExpected: %s\nActual: %s", input, string(lexer.input))
	}

	if lexer.index != 0 {
		t.Fatalf("Unexpected index\nExpected: 0\nActual: %d", lexer.index)
	}

	if lexer.buf != nil {
		t.Fatalf("Expected buffer to be nil")
	}
}

func Test_Lexer_getTokenNumber(t *testing.T) {
	testCases := map[string]struct {
		input    string
		expected Token
	}{
		"Integer": {
			input:    "123",
			expected: Token{Type: TokenType_Number, Value: "123"},
		},
		"Decimal": {
			input:    "123.45",
			expected: Token{Type: TokenType_Number, Value: "123.45"},
		},
		"Leading zero": {
			input:    "0123",
			expected: Token{Type: TokenType_Number, Value: "0123"},
		},
		"Zero decimal": {
			input:    "0.5",
			expected: Token{Type: TokenType_Number, Value: "0.5"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lexer := New(tc.input)
			result, err := lexer.getTokenNumber()

			if err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			if err := _tokensMatch(tc.expected, result); err != nil {
				t.Fatalf("Unexpected result: %s", err)
			}
		})
	}
}

func Test_Lexer_getTokenLabel(t *testing.T) {
	testCases := map[string]struct {
		input    string
		expected Token
	}{
		"Simple label": {
			input:    "test",
			expected: Token{Type: TokenType_Label, Value: "test"},
		},
		"With underscore": {
			input:    "test_label",
			expected: Token{Type: TokenType_Label, Value: "test_label"},
		},
		"With numbers": {
			input:    "test123",
			expected: Token{Type: TokenType_Label, Value: "test123"},
		},
		"Boolean true": {
			input:    "true",
			expected: Token{Type: TokenType_Boolean, Value: "true"},
		},
		"Boolean false": {
			input:    "false",
			expected: Token{Type: TokenType_Boolean, Value: "false"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lexer := New(tc.input)
			result, err := lexer.getTokenLabel()

			if err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			if err := _tokensMatch(tc.expected, result); err != nil {
				t.Fatalf("Unexpected result: %s", err)
			}
		})
	}
}

func Test_Lexer_getTokenStringLiteral(t *testing.T) {
	testCases := map[string]struct {
		input    string
		expected Token
	}{
		"Simple string": {
			input:    "hello\"",
			expected: Token{Type: TokenType_StringLiteral, Value: "hello"},
		},
		"Empty string": {
			input:    "\"",
			expected: Token{Type: TokenType_StringLiteral, Value: ""},
		},
		"With spaces": {
			input:    "hello world\"",
			expected: Token{Type: TokenType_StringLiteral, Value: "hello world"},
		},
		"With special chars": {
			input:    "hello!@#$%^&*()\"",
			expected: Token{Type: TokenType_StringLiteral, Value: "hello!@#$%^&*()"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lexer := New(tc.input)
			result, err := lexer.getTokenStringLiteral()

			if err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			if err := _tokensMatch(tc.expected, result); err != nil {
				t.Fatalf("Unexpected result: %s", err)
			}
		})
	}
}

func Test_Lexer_getTokenStringLiteral_EOF(t *testing.T) {
	input := "hello world"
	lexer := New(input)

	_, err := lexer.getTokenStringLiteral()

	if err == nil {
		t.Fatalf("Error expected but not returned")
	}

	if !errors.Is(err, errUnexpectedEOF) {
		t.Fatalf("Unexpected result\nExpected: %s\nActual: %s", errUnexpectedEOF, err)
	}
}

func Test_Lexer_getTokenBoolean(t *testing.T) {
	testCases := map[string]struct {
		input    string
		expected Token
	}{
		"True": {
			input:    "true",
			expected: Token{Type: TokenType_Boolean, Value: "true"},
		},
		"False": {
			input:    "false",
			expected: Token{Type: TokenType_Boolean, Value: "false"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lexer := New(tc.input)
			result, err := lexer.getTokenBoolean()

			if err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			if err := _tokensMatch(tc.expected, result); err != nil {
				t.Fatalf("Unexpected result: %s", err)
			}
		})
	}
}

func Test_Lexer_getTokenBoolean_Invalid(t *testing.T) {
	input := "maybe"
	lexer := New(input)

	_, err := lexer.getTokenBoolean()

	if err == nil {
		t.Fatalf("Error expected but not returned")
	}

	if !errors.Is(err, errInvalidRune) {
		t.Fatalf("Unexpected result\nExpected: %s\nActual: %s", errInvalidRune, err)
	}

	// Check that index was reset
	if lexer.index != 0 {
		t.Fatalf("Expected index to be reset to 0, got %d", lexer.index)
	}
}
