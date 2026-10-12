package querybuilder

import (
	"errors"
	"fmt"
	"strings"
)

type expressionKind uint8

const (
	expressionTerm expressionKind = iota
	expressionPhrase
	expressionField
	expressionAnd
	expressionOr
	expressionNot
	expressionAll
	expressionNone
)

// expression represents matching requirements independently of Bleve scoring.
// Field nodes supply a default field to their descendants; an explicit nested
// field overrides that default.
type expression struct {
	kind     expressionKind
	value    string
	children []*expression
}

const maxExpressionDepth = 128

var errExpressionDepth = errors.New("query nesting exceeds 128 levels")

type expressionParser struct {
	tokens []Token
	pos    int
	depth  int
}

func parseExpression(tokens []Token) (*expression, error) {
	if len(tokens) == 0 {
		return &expression{kind: expressionNone}, nil
	}
	p := expressionParser{tokens: tokens}
	root, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	if p.pos != len(tokens) {
		return nil, fmt.Errorf("unexpected %q at character %d", p.current().Value, p.current().start+1)
	}
	return root, nil
}

func (p *expressionParser) current() Token {
	if p.pos == len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *expressionParser) parseOr() (*expression, error) {
	first, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	children := []*expression{first}
	for p.current().Type == TokenOr {
		p.pos++
		next, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		children = append(children, next)
	}
	return compoundExpression(expressionOr, children), nil
}

func (p *expressionParser) parseAnd() (*expression, error) {
	first, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	children := []*expression{first}
	for {
		switch p.current().Type {
		case TokenEOF, TokenRightParen, TokenOr:
			return compoundExpression(expressionAnd, children), nil
		case TokenAnd:
			p.pos++
		}
		next, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		children = append(children, next)
	}
}

func (p *expressionParser) parseUnary() (*expression, error) {
	if p.depth >= maxExpressionDepth {
		return nil, errExpressionDepth
	}
	p.depth++
	defer func() { p.depth-- }()
	token := p.current()
	if token.Type == TokenEOF {
		return nil, errors.New("expected a term at end of query")
	}
	p.pos++
	switch token.Type {
	case TokenWord:
		return &expression{kind: expressionTerm, value: token.Value}, nil
	case TokenQuoted:
		return &expression{kind: expressionPhrase, value: token.Value}, nil
	case TokenNot, TokenField:
		// Preserve empty field values for the existing field validation and
		// fallback rules. In particular, url_re: must report an invalid regexp.
		var child *expression
		var err error
		if token.Type == TokenField && (p.current().Type == TokenEOF || p.current().Type == TokenRightParen || p.current().Type == TokenOr) {
			child = &expression{kind: expressionTerm}
		} else {
			child, err = p.parseUnary()
		}
		if err != nil {
			return nil, err
		}
		kind := expressionNot
		if token.Type == TokenField {
			kind = expressionField
		}
		return &expression{kind: kind, value: token.Value, children: []*expression{child}}, nil
	case TokenLeftParen:
		child, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if p.current().Type != TokenRightParen {
			return nil, fmt.Errorf("unclosed group at character %d", token.start+1)
		}
		p.pos++
		return child, nil
	default:
		return nil, fmt.Errorf("expected a term, found %q at character %d", token.Value, token.start+1)
	}
}

func compoundExpression(kind expressionKind, children []*expression) *expression {
	if len(children) == 1 {
		return children[0]
	}
	return &expression{kind: kind, children: children}
}

func simplifyExpression(e *expression, field string) *expression {
	if e.kind == expressionField {
		field = e.value
	}
	for i, child := range e.children {
		e.children[i] = simplifyExpression(child, field)
	}
	switch e.kind {
	case expressionTerm:
		if field == "" && e.value == "*" {
			return &expression{kind: expressionAll}
		}
	case expressionNot:
		child := e.children[0]
		switch child.kind {
		case expressionAll:
			return &expression{kind: expressionNone}
		case expressionNone:
			return &expression{kind: expressionAll}
		case expressionNot:
			return child.children[0]
		}
	case expressionAnd, expressionOr:
		identity, absorbing := expressionAll, expressionNone
		if e.kind == expressionOr {
			identity, absorbing = absorbing, identity
		}
		var children []*expression
		for _, child := range e.children {
			if child.kind == absorbing {
				return &expression{kind: absorbing}
			}
			if child.kind != identity {
				children = append(children, child)
			}
		}
		if len(children) == 0 {
			return &expression{kind: identity}
		}
		return compoundExpression(e.kind, children)
	}
	return e
}

// embeddingText preserves the existing semantic search input behavior while
// removing standalone wildcards, including those inside Boolean groups.
func (e *expression) embeddingText() string {
	switch e.kind {
	case expressionTerm, expressionPhrase:
		return e.value
	case expressionField:
		return e.value + ":" + e.children[0].embeddingText()
	case expressionNot:
		return "-" + e.children[0].embeddingText()
	case expressionAnd, expressionOr:
		parts := make([]string, len(e.children))
		for i, child := range e.children {
			parts[i] = child.embeddingText()
		}
		separator := " "
		if e.kind == expressionOr {
			separator = " | "
		}
		return strings.Join(parts, separator)
	default:
		return ""
	}
}
