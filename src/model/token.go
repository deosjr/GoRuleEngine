package model

// adapted from https://golang.org/src/go/token/token.go

import (
	"fmt"
	"strconv"
)

type Token int

const (
	ILLEGAL Token = iota
	EOF
	COMMENT

	literal_beg
	IDENT  // main
	INT    // 12345
	FLOAT  // 123.45
	STRING // "abc"
	literal_end

	operator_beg
	// Operators and delimiters
	ADD // +
	SUB // -
	MUL // *
	QUO // /
	REM // %

	// THERE IS NO ASSIGN
	EQL // =
	LSS // <
	GTR // >
	NOT // !
	NEQ // !=
	LEQ // <=
	GEQ // >=

	PERIOD // .
	operator_end

	LPAREN // (
	LBRACK // [
	LBRACE // {
	COMMA  // ,

	RPAREN    // )
	RBRACK    // ]
	RBRACE    // }
	SEMICOLON // ;
	COLON     // :

	keyword_beg
	// Keywords
	OBJECT
	RULE
	TEST
	FACTS
	RULES
	INPUT
	RELATION
	keyword_end
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	COMMENT: "COMMENT",

	IDENT:  "ident",
	INT:    "int",
	FLOAT:  "float",
	STRING: "string",

	ADD: "+",
	SUB: "-",
	MUL: "*",
	QUO: "/",
	REM: "%",

	EQL: "==",
	LSS: "<",
	GTR: ">",
	NOT: "!",
	NEQ: "!=",
	LEQ: "<=",
	GEQ: ">=",

	PERIOD: ".",

	LPAREN: "(",
	LBRACK: "[",
	LBRACE: "{",
	COMMA:  ",",

	RPAREN:    ")",
	RBRACK:    "]",
	RBRACE:    "}",
	SEMICOLON: ";",
	COLON:     ":",

	OBJECT:   "object",
	RULE:     "rule",
	TEST:     "test",
	FACTS:    "facts",
	RULES:    "rules",
	INPUT:    "input",
	RELATION: "relation",
}

func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

const (
	LowestPrec  = 0 // non-operators
	UnaryPrec   = 6
	HighestPrec = 7
)

// Precedence returns the operator precedence of the binary
// operator op. If op is not a binary operator, the result
// is LowestPrecedence.
func (op Token) Precedence() int {
	switch op {
	case EQL, NEQ, LSS, LEQ, GTR, GEQ:
		return 3
	case ADD, SUB:
		return 4
	case MUL, QUO, REM:
		return 5
	case PERIOD:
		return 6
	}
	return LowestPrec
}

var types map[string]Token
var keywords map[string]Token
var operators map[string]Token

func init() {
	types = make(map[string]Token)
	for i := literal_beg + 1; i < literal_end; i++ {
		types[tokens[i]] = i
	}
	keywords = make(map[string]Token)
	for i := keyword_beg + 1; i < keyword_end; i++ {
		keywords[tokens[i]] = i
	}
	operators = make(map[string]Token)
	for i := operator_beg + 1; i < operator_end; i++ {
		operators[tokens[i]] = i
	}
}

func lookupType(t string) Token {
	if tok, is_type := types[t]; is_type {
		return tok
	}
	// type not builtin, must be an object
	return OBJECT
}

func lookupToken(ident string) Token {
	if tok, is_keyword := keywords[ident]; is_keyword {
		return tok
	}
	return IDENT
}

func lookupOperator(op string) Token {
	if tok, is_operator := operators[op]; is_operator {
		return tok
	}
	panic(fmt.Sprintf("illegal operator %s", op))
}

func (tok Token) IsLiteral() bool { return literal_beg < tok && tok < literal_end }

func (tok Token) IsOperator() bool { return operator_beg < tok && tok < operator_end }

func (tok Token) IsKeyword() bool { return keyword_beg < tok && tok < keyword_end }

func tokenValue(tok Token, lit string) interface{} {
	switch tok {
	case INT:
		i, err := strconv.Atoi(lit)
		if err != nil {
			panic(err)
		}
		return i
	case FLOAT:
		i, err := strconv.ParseFloat(lit, 64)
		if err != nil {
			panic(err)
		}
		return i
	case STRING, IDENT:
		return lit
	default:
		return lit
	}
}
