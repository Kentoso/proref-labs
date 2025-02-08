package main

import (
	"slices"
	"unicode"
)

type TokenType string

const (
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"
	EOL     TokenType = "EOL"
	SEGM    TokenType = "SEGM"
	FIELD   TokenType = "FIELD"
	LCHILD  TokenType = "LCHILD"
	XDFLD   TokenType = "XDFLD"
	IDENT   TokenType = "IDENT"
	ATTR    TokenType = "ATTR"
	EQUALS  TokenType = "="
	COMMA   TokenType = ","
	LPAREN  TokenType = "("
	RPAREN  TokenType = ")"
)

// Token represents a lexical token.
type Token struct {
	Type    TokenType
	Literal string
}

var attributes = []string{
	"BYTES",
	"COMPRTN",
	"CONST",
	"DDATA",
	"DSGROUP",
	"EXIT",
	"EXTRN",
	"FREQ",
	"INDEX",
	"NAME",
	"NULLVAL",
	"PAIR",
	"PARENT",
	"POINTER",
	"PTR",
	"RKSIZE",
	"RMNAME",
	"RULES",
	"SEGMENT",
	"SOURCE",
	"SRCH",
	"SSPTR",
	"START",
	"SUBSEQ",
	"TYPE",
}

func lex(input string) []Token {
	var tokens []Token
	i := 0
	for i < len(input) {
		ch := input[i]

		if ch == '\n' {
			tokens = append(tokens, Token{Type: EOL, Literal: ""})
		}

		if unicode.IsSpace(rune(ch)) {
			i++
			continue
		}
		switch ch {
		case '=':
			tokens = append(tokens, Token{Type: EQUALS, Literal: "="})
			i++
		case ',':
			tokens = append(tokens, Token{Type: COMMA, Literal: ","})
			i++
		case '(':
			tokens = append(tokens, Token{Type: LPAREN, Literal: "("})
			i++
		case ')':
			tokens = append(tokens, Token{Type: RPAREN, Literal: ")"})
			i++
		default:
			start := i
			for i < len(input) && !unicode.IsSpace(rune(input[i])) &&
				input[i] != '=' && input[i] != ',' && input[i] != '(' && input[i] != ')' {
				i++
			}
			tokenStr := input[start:i]
			if tokenStr == "SEGM" {
				tokens = append(tokens, Token{Type: SEGM, Literal: tokenStr})
			} else if tokenStr == "FIELD" {
				tokens = append(tokens, Token{Type: FIELD, Literal: tokenStr})
			} else if tokenStr == "LCHILD" {
				tokens = append(tokens, Token{Type: LCHILD, Literal: tokenStr})
			} else if tokenStr == "XDFLD" {
				tokens = append(tokens, Token{Type: XDFLD, Literal: tokenStr})
			} else if slices.Contains(attributes, tokenStr) {
				tokens = append(tokens, Token{Type: ATTR, Literal: tokenStr})
			} else {
				tokens = append(tokens, Token{Type: IDENT, Literal: tokenStr})
			}
		}
	}
	tokens = append(tokens, Token{Type: EOL, Literal: ""})
	return tokens
}
