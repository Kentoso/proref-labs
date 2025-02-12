package main

import (
	"fmt"
	"strings"
)

// KeyValue holds one parsed field (for example, NAME=TYPE).
type KeyValue struct {
	Key   string
	Value string
}

// Parser holds the tokens and current position.
type Parser struct {
	tokens []Token
	pos    int
}

type Segm struct {
	Attributes []KeyValue
	Fields     []*Field
	LChilds    []*LChild
	XDFlds     []*XDFld
}

type Field struct {
	Attributes []KeyValue
}

type LChild struct {
	Attributes []KeyValue
}

type XDFld struct {
	Attributes []KeyValue
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

func (p *Parser) curToken() Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return Token{Type: EOF, Literal: ""}
}

func (p *Parser) nextToken() {
	p.pos++
}

func (p *Parser) ParseDBD() ([]*Segm, error) {
	var segms []*Segm

	var currentSegm *Segm

	for {
		skipped := p.parseSkip()
		for skipped {
			skipped = p.parseSkip()
			if currentSegm != nil {
				segms = append(segms, currentSegm)
				currentSegm = nil
			}
		}

		segmAttrs, errSegm := p.parseSegm()
		if errSegm == nil {
			if currentSegm != nil {
				segms = append(segms, currentSegm)
			}
			currentSegm = &Segm{}
			currentSegm.Attributes = segmAttrs
			continue
		}

		field, errField := p.parseField()
		if errField == nil && currentSegm != nil {
			currentSegm.Fields = append(currentSegm.Fields, field)
			continue
		}

		lchild, errLChild := p.parseLChild()
		if errLChild == nil && currentSegm != nil {
			currentSegm.LChilds = append(currentSegm.LChilds, lchild)
			continue
		}

		xdFld, errXDFld := p.parseXDFld()
		if errXDFld == nil && currentSegm != nil {
			currentSegm.XDFlds = append(currentSegm.XDFlds, xdFld)
			continue
		}

		if currentSegm != nil {
			segms = append(segms, currentSegm)
			currentSegm = nil
		}

		if p.curToken().Type == EOF {
			break
		}

		fmt.Printf("%v\n", currentSegm)

		errFormat := "unexpected token at %v %v: %q\nSEGM: %w\nFIELD: %w\nLCHILD: %w\nXDFLD: %w\n"

		return nil, fmt.Errorf(errFormat, p.curToken().Line, p.curToken().Column, p.curToken().Literal, errSegm, errField, errLChild, errXDFld)
	}

	return segms, nil
}

func (p *Parser) parseSkip() bool {
	skipped := false

	tok := p.curToken()
	if tok.Type == SKIPLINE {
		skipped = true
		for tok.Type != EOL {
			p.nextToken()
			tok = p.curToken()
		}
		p.nextToken()
		tok = p.curToken()
	}
	if tok.Type == LABEL {
		skipped = true
		p.nextToken()
	}

	return skipped
}

func (p *Parser) parseSegm() ([]KeyValue, error) {
	tok := p.curToken()
	if tok.Type != SEGM {
		return nil, fmt.Errorf("expected SEGM, got %q", tok.Literal)
	}
	p.nextToken()

	attributes, err := p.parseAttributeList()
	if err != nil {
		return nil, err
	}
	return attributes, nil
}

func (p *Parser) parseField() (*Field, error) {
	tok := p.curToken()
	if tok.Type != FIELD {
		return nil, fmt.Errorf("expected FIELD, got %q", tok.Literal)
	}
	p.nextToken()
	attributes, err := p.parseAttributeList()
	if err != nil {
		return nil, err
	}

	return &Field{Attributes: attributes}, nil
}

func (p *Parser) parseLChild() (*LChild, error) {
	tok := p.curToken()
	if tok.Type != LCHILD {
		return nil, fmt.Errorf("expected LCHILD, got %q", tok.Literal)
	}
	p.nextToken()
	attributes, err := p.parseAttributeList()
	if err != nil {
		return nil, err
	}

	return &LChild{Attributes: attributes}, nil
}

func (p *Parser) parseXDFld() (*XDFld, error) {
	tok := p.curToken()
	if tok.Type != XDFLD {
		return nil, fmt.Errorf("expected XDFLD, got %q", tok.Literal)
	}
	p.nextToken()
	attributes, err := p.parseAttributeList()
	if err != nil {
		return nil, err
	}

	return &XDFld{Attributes: attributes}, nil
}

// parseAttributeList parses one or more key=value pairs separated by commas.
func (p *Parser) parseAttributeList() ([]KeyValue, error) {
	var fields []KeyValue
	for {
		tok := p.curToken()

		if tok.Type == EOL {
			p.nextToken()
			break
		}
		// Skip extraneous commas.
		if tok.Type == COMMA {
			p.nextToken()
			continue
		}
		kv, err := p.parseAttribute()
		if err != nil {
			return nil, err
		}
		fields = append(fields, kv)
	}
	return fields, nil
}

// parseAttribute parses a single field: key "=" value.
func (p *Parser) parseAttribute() (KeyValue, error) {
	// Parse the key.
	keyTok := p.curToken()
	if keyTok.Type != ATTR {
		return KeyValue{}, fmt.Errorf("expected key identifier, got %q", keyTok.Literal)
	}
	key := keyTok.Literal
	p.nextToken()
	// Expect an "=".
	equalsTok := p.curToken()
	if equalsTok.Type != EQUALS {
		return KeyValue{}, fmt.Errorf("expected '=', got %q", equalsTok.Literal)
	}
	p.nextToken()
	// Parse the value.
	value, err := p.parseValue()
	if err != nil {
		return KeyValue{}, err
	}
	return KeyValue{Key: key, Value: value}, nil
}

// parseValue parses a value, which may be a simple identifier or a parenthesized list.
func (p *Parser) parseValue() (string, error) {
	tok := p.curToken()
	if tok.Type == LPAREN {
		// For a parenthesized list, capture everything until the matching RPAREN.
		p.nextToken() // skip '('
		var parts []string
		for {
			t := p.curToken()
			if t.Type == RPAREN {
				p.nextToken() // skip ')'
				break
			}
			// Skip commas within the list.
			if t.Type == COMMA {
				p.nextToken()
				continue
			}
			if t.Type == IDENT {
				parts = append(parts, t.Literal)
				p.nextToken()
			} else {
				return "", fmt.Errorf("unexpected token in parenthesized list: %q", t.Literal)
			}
		}
		return "(" + strings.Join(parts, ",") + ")", nil
	} else if tok.Type == IDENT || tok.Type == ATTR {
		p.nextToken()
		return tok.Literal, nil
	}
	return "", fmt.Errorf("unexpected token for value: %q", tok.Literal)
}
