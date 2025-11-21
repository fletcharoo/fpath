package lexer

import (
	"errors"
	"fmt"
	"io"
	"unicode"
)

const (
	TokenType_Undefined = iota
	TokenType_Number
	TokenType_Label
	TokenType_StringLiteral
	TokenType_Boolean
	TokenType_Dollar
	TokenType_Plus
	TokenType_Minus
	TokenType_Asterisk
	TokenType_Slash
	TokenType_Equals
	TokenType_NotEquals
	TokenType_GreaterThan
	TokenType_GreaterThanOrEqual
	TokenType_LessThan
	TokenType_LessThanOrEqual
	TokenType_And
	TokenType_Or
	TokenType_LeftParan
	TokenType_RightParan
	TokenType_LeftBracket
	TokenType_RightBracket
	TokenType_LeftBrace
	TokenType_RightBrace
	TokenType_Colon
	TokenType_Comma
)

var (
	TokenTypeString map[int]string = map[int]string{
		TokenType_Undefined:          "Undefined",
		TokenType_Number:             "Number",
		TokenType_Label:              "Label",
		TokenType_StringLiteral:      "StringLiteral",
		TokenType_Boolean:            "Boolean",
		TokenType_Dollar:             "Dollar",
		TokenType_Plus:               "Plus",
		TokenType_Minus:              "Minus",
		TokenType_Asterisk:           "Asterisk",
		TokenType_Slash:              "Slash",
		TokenType_Equals:             "Equals",
		TokenType_NotEquals:          "NotEquals",
		TokenType_GreaterThan:        "GreaterThan",
		TokenType_GreaterThanOrEqual: "GreaterThanOrEqual",
		TokenType_LessThan:           "LessThan",
		TokenType_LessThanOrEqual:    "LessThanOrEqual",
		TokenType_And:                "And",
		TokenType_Or:                 "Or",
		TokenType_LeftParan:          "LeftParan",
		TokenType_RightParan:         "RightParan",
		TokenType_LeftBracket:        "LeftBracket",
		TokenType_RightBracket:       "RightBracket",
		TokenType_LeftBrace:          "LeftBrace",
		TokenType_RightBrace:         "RightBrace",
		TokenType_Colon:              "Colon",
		TokenType_Comma:              "Comma",
	}
)

var (
	errUnexpectedEOF = errors.New("unexpected EOF")
	errInvalidRune   = errors.New("invalid rune")
)

// isLabelRune returns whether the provided rune is a valid label rune.
// Valid label runes are letters, numbers, and underscores.
func isLabelRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_'
}

type Token struct {
	Type  int
	Value string
}

// String makes Token implement the Stringer interface.
func (t Token) String() string {
	s, ok := TokenTypeString[t.Type]

	if !ok {
		return TokenTypeString[TokenType_Undefined]
	}

	return s
}

// New returns a new Lexer configured to read from a slice
// []rune value of the input string.
func New(input string) *Lexer {
	return &Lexer{
		input: []rune(input),
	}
}

// Lexer adds the functionality to get and peek tokens from a
// string using a buffer.
type Lexer struct {
	input []rune
	index int
	buf   *Token
}

// getRune returns the rune at the current index of the input and increments the
// index.
// If the index is larger than the length of the input, getRune returns an
// io.EOF error.
func (l *Lexer) getRune() (r rune, err error) {
	if l.index == len(l.input) {
		return 0, io.EOF
	}

	r = l.input[l.index]
	l.index++
	return r, nil
}

// peekRune returns the rune at the current index of the input but doesn't
// increment the index.
// If the index is larger than the length of the input, peekRune returns an
// io.EOF error.
func (l *Lexer) peekRune() (r rune, err error) {
	if l.index == len(l.input) {
		return 0, io.EOF
	}

	return l.input[l.index], nil
}

// getToken returns the next token in the input string.
// If there are no more tokens to process in the string, getToken returns an
// io.EOF error.
func (l *Lexer) GetToken() (tok Token, err error) {
	if l.buf != nil {
		tok = *l.buf
		l.buf = nil
		return tok, nil
	}

	var r rune

	for {
		r, err = l.peekRune()

		if err != nil {
			return tok, err
		}

		if unicode.IsSpace(r) {
			l.index++
			continue
		}

		if unicode.IsNumber(r) {
			return l.getTokenNumber()
		}

		if isLabelRune(r) {
			return l.getTokenLabel()
		}

		switch r {
		case '"':
			l.index++
			return l.getTokenStringLiteral()
		case '$':
			l.index++
			return Token{
				Type: TokenType_Dollar,
			}, nil
		case '+':
			l.index++
			return Token{
				Type: TokenType_Plus,
			}, nil
		case '-':
			l.index++
			return Token{
				Type: TokenType_Minus,
			}, nil
		case '*':
			l.index++
			return Token{
				Type: TokenType_Asterisk,
			}, nil
		case '/':
			l.index++
			return Token{
				Type: TokenType_Slash,
			}, nil
		case '=':
			l.index++
			// Check if this is the start of == operator
			nextRune, peekErr := l.peekRune()
			if peekErr == nil && nextRune == '=' {
				l.index++
				return Token{
					Type: TokenType_Equals,
				}, nil
			}
			// Single = is not supported, return error
			err = fmt.Errorf("%w: %s", errInvalidRune, string(r))
			return
		case '!':
			l.index++
			// Check if this is the start of != operator
			nextRune, peekErr := l.peekRune()
			if peekErr == nil && nextRune == '=' {
				l.index++
				return Token{
					Type: TokenType_NotEquals,
				}, nil
			}
			// Single ! is not supported, return error
			err = fmt.Errorf("%w: %s", errInvalidRune, string(r))
			return
		case '&':
			l.index++
			// Check if this is the start of && operator
			nextRune, peekErr := l.peekRune()
			if peekErr == nil && nextRune == '&' {
				l.index++
				return Token{
					Type: TokenType_And,
				}, nil
			}
			// Single & is not supported, return error
			err = fmt.Errorf("%w: %s", errInvalidRune, string(r))
			return
		case '|':
			l.index++
			// Check if this is the start of || operator
			nextRune, peekErr := l.peekRune()
			if peekErr == nil && nextRune == '|' {
				l.index++
				return Token{
					Type: TokenType_Or,
				}, nil
			}
			// Single | is not supported, return error
			err = fmt.Errorf("%w: %s", errInvalidRune, string(r))
			return
		case '>':
			l.index++
			// Check if this is the start of >= operator
			nextRune, peekErr := l.peekRune()
			if peekErr == nil && nextRune == '=' {
				l.index++
				return Token{
					Type: TokenType_GreaterThanOrEqual,
				}, nil
			}
			return Token{
				Type: TokenType_GreaterThan,
			}, nil
		case '<':
			l.index++
			// Check if this is the start of <= operator
			nextRune, peekErr := l.peekRune()
			if peekErr == nil && nextRune == '=' {
				l.index++
				return Token{
					Type: TokenType_LessThanOrEqual,
				}, nil
			}
			return Token{
				Type: TokenType_LessThan,
			}, nil
		case '(':
			l.index++
			return Token{
				Type: TokenType_LeftParan,
			}, nil
		case ')':
			l.index++
			return Token{
				Type: TokenType_RightParan,
			}, nil
		case '[':
			l.index++
			return Token{
				Type: TokenType_LeftBracket,
			}, nil
		case ']':
			l.index++
			return Token{
				Type: TokenType_RightBracket,
			}, nil
		case '{':
			l.index++
			return Token{
				Type: TokenType_LeftBrace,
			}, nil
		case '}':
			l.index++
			return Token{
				Type: TokenType_RightBrace,
			}, nil
		case ':':
			l.index++
			return Token{
				Type: TokenType_Colon,
			}, nil
		case ',':
			l.index++
			return Token{
				Type: TokenType_Comma,
			}, nil
		default:
			err = fmt.Errorf("%w: %s", errInvalidRune, string(r))
			return
		}
	}
}

// peekToken returns the current token in the input string but does not
// increment the index.
// If there are no more tokens to process in the string, getToken returns an
// io.EOF error.
func (l *Lexer) PeekToken() (tok Token, err error) {
	if l.buf != nil {
		tok = *l.buf
		return tok, nil
	}

	tok, err = l.GetToken()
	l.buf = &tok
	return tok, err
}

// getTokenNumber returns the current number token in the input string.
// If there are no more tokens to process in the string, getToken returns an
// io.EOF error.
func (l *Lexer) getTokenNumber() (tok Token, err error) {
	tok.Type = TokenType_Number
	var r rune
	hasDecimalPoint := false

	for {
		r, err = l.peekRune()

		if err == io.EOF {
			return tok, nil
		}

		if err != nil {
			return tok, err
		}

		if unicode.IsNumber(r) {
			l.index++
			tok.Value += string(r)
			continue
		}

		if r == '.' && !hasDecimalPoint {
			l.index++
			tok.Value += string(r)
			hasDecimalPoint = true
			continue
		}

		return tok, nil
	}
}

// getTokenLabel returns the current label token in the input string.
// If there are no more tokens to process in the string, getToken returns an
// io.EOF error.
func (l *Lexer) getTokenLabel() (tok Token, err error) {
	tok.Type = TokenType_Label
	var r rune

	for {
		r, err = l.peekRune()

		if err != nil {
			break
		}

		if isLabelRune(r) {
			l.index++
			tok.Value += string(r)
			continue
		}

		break
	}

	// Check if this is a boolean literal
	if tok.Value == "true" || tok.Value == "false" {
		tok.Type = TokenType_Boolean
	}

	if err == io.EOF {
		return tok, nil
	}

	return tok, err
}

// getTokenStringLiteral returns the current string literal token in the input
// string.
// If the token reaches the end of the string, getTokenStringLiteral returns an
// UnexpectedEOF error.
func (l *Lexer) getTokenStringLiteral() (tok Token, err error) {
	tok.Type = TokenType_StringLiteral
	var r rune

	for {
		r, err = l.getRune()

		if err == io.EOF {
			err = errUnexpectedEOF
			return
		}

		if err != nil {
			return
		}

		if r == '"' {
			break
		}

		tok.Value += string(r)
	}

	return tok, nil
}

// getTokenBoolean returns the current boolean literal token in the input
// string. It recognizes "true" and "false".
func (l *Lexer) getTokenBoolean() (tok Token, err error) {
	tok.Type = TokenType_Boolean
	var r rune

	// Peek ahead to see if this is "true" or "false"
	startIndex := l.index
	for {
		r, err = l.peekRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			return tok, err
		}
		if !isLabelRune(r) {
			break
		}
		l.index++
		tok.Value += string(r)
	}

	// Validate that it's either "true" or "false"
	if tok.Value != "true" && tok.Value != "false" {
		// Reset index and treat as invalid rune
		l.index = startIndex
		err = fmt.Errorf("%w: %s", errInvalidRune, tok.Value)
		return
	}

	return tok, nil
}
