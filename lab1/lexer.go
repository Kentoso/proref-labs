package main

import (
	"fmt"
	"slices"
	"unicode"
)

type TokenType string

const (
	EOF      TokenType = "EOF"
	EOL      TokenType = "EOL"
	SEGM     TokenType = "SEGM"
	FIELD    TokenType = "FIELD"
	LCHILD   TokenType = "LCHILD"
	XDFLD    TokenType = "XDFLD"
	IDENT    TokenType = "IDENT"
	ATTR     TokenType = "ATTR"
	EQUALS   TokenType = "="
	COMMA    TokenType = ","
	LPAREN   TokenType = "("
	RPAREN   TokenType = ")"
	LABEL    TokenType = "LABEL"
	SKIPLINE TokenType = "SKIPLINE"
)

// Token represents a lexical token.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
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

type tokenCreator struct {
	currentLine   *int
	currentColumn *int
}

func (tc tokenCreator) createToken(tokenType TokenType, literal string) Token {
	return Token{Type: tokenType, Literal: literal, Line: *tc.currentLine, Column: *tc.currentColumn}
}

func lex(input string) ([]Token, error) {
	var tokens []Token
	currentLine := 0
	currentColumn := 0

	tc := tokenCreator{currentLine: &currentLine, currentColumn: &currentColumn}

	i := 0
	for i < len(input) {
		ch := input[i]

		if ch == '\n' {
			if currentColumn != 80 {
				return nil, fmt.Errorf("line %d: line length is not equal to 80", currentLine)
			}

			tokens = append(tokens, tc.createToken(EOL, ""))
			i++
			currentLine++
			currentColumn = 0
			continue
		}

		if unicode.IsSpace(rune(ch)) {
			currentColumn++
			i++
			continue
		}

		// skip line numbers
		if currentColumn >= 72 {
			i++
			currentColumn++
			continue
		}

		switch ch {
		case '*':
			if currentColumn != 71 {
				return nil, fmt.Errorf("line %d column: %d: invalid line continuation position", currentLine, currentColumn)
			}

			i += 9
			currentColumn += 9
			nextChar := input[i]

			if nextChar == '\n' {
				i++
				currentLine++
				currentColumn = 0
			} else {
				return nil, fmt.Errorf("line %d column: %d: no newline after line continuation", currentLine, currentColumn)
			}
		case '=':
			tokens = append(tokens, tc.createToken(EQUALS, "="))
			currentColumn++
			i++
		case ',':
			tokens = append(tokens, tc.createToken(COMMA, ","))
			currentColumn++
			i++
		case '(':
			tokens = append(tokens, tc.createToken(LPAREN, "("))
			currentColumn++
			i++
		case ')':
			tokens = append(tokens, tc.createToken(RPAREN, ")"))
			currentColumn++
			i++
		default:
			start := i
			for i < len(input) && !unicode.IsSpace(rune(input[i])) &&
				input[i] != '=' && input[i] != ',' && input[i] != '(' && input[i] != ')' {
				i++
			}

			tokenStr := input[start:i]
			if tokenStr == "SEGM" {
				tokens = append(tokens, tc.createToken(SEGM, tokenStr))
			} else if tokenStr == "FIELD" {
				tokens = append(tokens, tc.createToken(FIELD, tokenStr))
			} else if tokenStr == "LCHILD" {
				tokens = append(tokens, tc.createToken(LCHILD, tokenStr))
			} else if tokenStr == "XDFLD" {
				tokens = append(tokens, tc.createToken(XDFLD, tokenStr))
			} else if slices.Contains(attributes, tokenStr) {
				tokens = append(tokens, tc.createToken(ATTR, tokenStr))
			} else if currentColumn == 0 {
				tokens = append(tokens, tc.createToken(LABEL, tokenStr))
			} else if currentColumn == 7 || tokenStr == "DBD" || tokenStr == "DATASET" {
				tokens = append(tokens, tc.createToken(SKIPLINE, tokenStr))
			} else {
				tokens = append(tokens, tc.createToken(IDENT, tokenStr))
			}

			currentColumn += i - start
		}
	}
	tokens = append(tokens, Token{Type: EOL, Literal: ""})
	return tokens, nil
}
