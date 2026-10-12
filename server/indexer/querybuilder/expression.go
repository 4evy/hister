package querybuilder

import (
	"errors"
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
		return nil, diagnostic(p.current(), "unexpected closing parenthesis; remove it or add an opening parenthesis")
	}
	return root, nil
}

func (p *expressionParser) current() Token {
	if p.pos == len(p.tokens) {
		end := p.tokens[len(p.tokens)-1].end
		return Token{Type: TokenEOF, start: end, end: end}
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
		operator := p.current()
		p.pos++
		if err := p.requireOperand(operator); err != nil {
			return nil, err
		}
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
			operator := p.current()
			p.pos++
			if err := p.requireOperand(operator); err != nil {
				return nil, err
			}
		}
		next, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		children = append(children, next)
	}
}

func (p *expressionParser) requireOperand(operator Token) error {
	switch p.current().Type {
	case TokenEOF, TokenRightParen, TokenAnd, TokenOr:
		if operator.Type == TokenField {
			return diagnostic(operator, "expected a value after %s:", operator.Value)
		}
		return diagnostic(operator, "expected a term after %s", operator.Value)
	default:
		return nil
	}
}

func (p *expressionParser) parseUnary() (*expression, error) {
	if p.depth >= maxExpressionDepth {
		return nil, diagnostic(p.current(), "%w", errExpressionDepth)
	}
	p.depth++
	defer func() { p.depth-- }()
	token := p.current()
	if token.Type == TokenEOF {
		return nil, diagnostic(token, "expected a term at end of query")
	}
	p.pos++
	switch token.Type {
	case TokenWord:
		return &expression{kind: expressionTerm, value: token.Value}, nil
	case TokenQuoted:
		if token.Value == "" {
			return nil, diagnostic(token, "empty quoted value; add text between the quotes")
		}
		return &expression{kind: expressionPhrase, value: token.Value}, nil
	case TokenNot, TokenField:
		if err := p.requireOperand(token); err != nil {
			return nil, err
		}
		child, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		kind := expressionNot
		if token.Type == TokenField {
			kind = expressionField
		}
		return &expression{kind: kind, value: token.Value, children: []*expression{child}}, nil
	case TokenLeftParen:
		if p.current().Type == TokenEOF {
			return nil, diagnostic(token, "unclosed group; add a term and a closing parenthesis")
		}
		if p.current().Type == TokenRightParen {
			return nil, diagnostic(token, "empty group; add a term between the parentheses")
		}
		child, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if p.current().Type != TokenRightParen {
			return nil, diagnostic(token, "unclosed group; add a closing parenthesis")
		}
		p.pos++
		return child, nil
	default:
		return nil, diagnostic(token, "expected a term, found %q", token.Value)
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
