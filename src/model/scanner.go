package model

import (
	"bufio"
	"io"
	"unicode"
	"unicode/utf8"
)

type scanner struct {
	r  *bufio.Reader
	ch rune // current character
	// TODO: position info
}

func newScanner(r io.Reader) *scanner {
	return &scanner{r: bufio.NewReader(r)}
}

func (s *scanner) next() {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		s.ch = 0 //EOF
	}
	s.ch = ch
}

func (s *scanner) lookahead(r rune) bool {
	s.next()
	if s.ch == r {
		return true
	}
	s.r.UnreadRune()
	return false
}

func (s *scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' || s.ch == '\r' {
		s.next()
	}
}

// TODO: dealing with CR?
func (s *scanner) scanComment() (tok Token, lit string) {
	for s.ch != '\n' {
		lit += string(s.ch)
		s.next()
	}
	return COMMENT, lit
}

func (s *scanner) scanIdentifier() (tok Token, lit string) {
	for isLetter(s.ch) || isDigit(s.ch) {
		lit += string(s.ch)
		s.next()
	}
	tok = lookupToken(lit)
	return
}

func (s *scanner) scanString() (tok Token, lit string) {
	s.next()
	for s.ch != '"' {
		lit += string(s.ch)
		s.next()
	}
	s.next()
	return STRING, lit
}

// TODO: float
func (s *scanner) scanNumber() (tok Token, lit string) {
	for isDigit(s.ch) {
		lit += string(s.ch)
		s.next()
	}
	return INT, lit
}

func isLetter(r rune) bool {
	return 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || r == '_' || r >= utf8.RuneSelf && unicode.IsLetter(r)
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9' || r >= utf8.RuneSelf && unicode.IsDigit(r)
}

// as per https://golang.org/src/go/scanner/scanner.go

// Scan scans the next token and returns the token and its literal
// string if applicable. The source end is indicated by EOF.
//
// If the returned token is a literal or COMMENT, the literal
// string has the corresponding value.
//
// If the returned token is a keyword, the literal string is the keyword.
//
// If the returned token is ILLEGAL, the literal string is the
// offending character.
//
// In all other cases, Scan returns an empty literal string.
func (s *scanner) scan() (tok Token, lit string) {
	// TODOS: scan float
	s.skipWhitespace()

	if isLetter(s.ch) {
		return s.scanIdentifier()
	}
	if '0' <= s.ch && s.ch <= '9' {
		return s.scanNumber()
	}

	switch s.ch {
	case 0:
		tok = EOF
	case '#':
		return s.scanComment()
	case '"':
		return s.scanString()
	case '.':
		tok = PERIOD
	case ',':
		tok = COMMA
	case ':':
		tok = COLON
	case '(':
		tok = LPAREN
	case ')':
		tok = RPAREN
	case '[':
		tok = LBRACK
	case ']':
		tok = RBRACK
	case '{':
		tok = LBRACE
	case '}':
		tok = RBRACE
	case '>':
		if s.lookahead('=') {
			tok = GEQ
		} else {
			tok = GTR
		}
	case '<':
		if s.lookahead('=') {
			tok = LEQ
		} else {
			tok = LSS
		}
	case '=':
		tok = EQL
	case '+':
		tok = ADD
	case '-':
		tok = SUB
	case '*':
		tok = MUL
	case '/':
		tok = QUO
	case '%':
		tok = REM
	}
	s.next()
	return
}
