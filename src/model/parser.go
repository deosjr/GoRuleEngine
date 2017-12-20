package model

import (
	"fmt"
	"io"
)

type parser struct {
	scanner *scanner
	tok     Token
	lit     string

	// currently defined variables -> objectType
	varsInScope map[string]string
}

func newParser(r io.Reader) *parser {
	s := newScanner(r)
	s.next()
	p := &parser{scanner: s}
	p.next()
	return p
}

func (p *parser) next() {
	p.tok, p.lit = p.scanner.scan()
	if p.tok == COMMENT {
		p.next()
	}
}

func (p *parser) expect(expected Token) string {
	if p.tok == expected {
		lit := p.lit
		p.next()
		return lit
	}
	panic(fmt.Errorf("expected %v got %s", expected, p.tok.String()))
}

func (p *parser) expectOneOf(expected ...Token) (Token, string) {
	for _, e := range expected {
		if p.tok == e {
			tok, lit := p.tok, p.lit
			p.next()
			return tok, lit
		}
	}
	panic(fmt.Errorf("expected %v got %s", expected, p.tok.String()))
}

func (p *parser) expectSequence(expected ...Token) {
	for _, e := range expected {
		p.expect(e)
	}
}

func (p *parser) parseObject() Object {
	objectName := p.expect(IDENT)
	o := Object{Name: objectName}
	p.expect(LBRACE)
	for {
		name := p.expect(IDENT)
		p.expect(COLON)
		typeInfo := p.expect(IDENT)
		f := p.parseField(name, typeInfo)
		o.Fields = append(o.Fields, f)
		if !p.commaOrRbrace() {
			break
		}
	}
	return o
}

func (p *parser) parseField(name, typeInfo string) Field {
	return Field{Name: name, TypeInfo: lookupType(typeInfo)}
}

func (p *parser) parseTermWithType(name, typeInfo string) Term {
	typ := lookupType(typeInfo)
	t := Term{Value: name, TypeInfo: typ}
	if typ == OBJECT {
		t.fieldInfo = typeInfo
	}
	return t
}

func (p *parser) parseTerm(tok Token, lit string) Term {
	if tok == IDENT {
		// TODO: assumption: object is declared higher up the file
		// or at least earlier
		if objectName, ok := p.varsInScope[lit]; ok {
			return ObjectTerm(lit, objectName)
		}
	}
	return Term{TypeInfo: tok, Value: tokenValue(tok, lit)}
}

func (p *parser) parseRelation() Relation {
	relationName := p.expect(IDENT)
	r := Relation{Name: relationName}
	p.expect(LBRACE)
	for {
		name := p.expect(IDENT)
		p.expect(COLON)
		typeInfo := p.expect(IDENT)
		f := p.parseField(name, typeInfo)
		r.Fields = append(r.Fields, f)
		if !p.commaOrRbrace() {
			break
		}
	}
	return r
}

// parse list of terms as args: functor(a1, a2, a3...)
func (p *parser) parseRuleCall() (e Expression, more bool) {
	functor := p.expect(IDENT)
	e = Expression{Functor: functor, Args: []Node{}}
	p.expect(LPAREN)
	for {
		tok, lit := p.expectOneOf(IDENT, INT, FLOAT, STRING)
		t := p.parseTerm(tok, lit)
		e.Args = append(e.Args, t)
		tok, _ = p.expectOneOf(COMMA, RPAREN)
		if tok == RPAREN {
			break
		}
	}
	return e, p.commaOrRbrace()
}

func (p *parser) parseExpression() (e Expression, more bool) {
	// scanner is already looking 1 rune ahead
	if p.tok == IDENT && p.scanner.ch == '(' {
		return p.parseRuleCall()
	}

	return p.parseExpressionTree(COMMA, RBRACE)
}

func (p *parser) parseExpressionTree(terminators ...Token) (e Expression, more bool) {
	n1 := p.parseNode()
	op := p.parseOperator()
	n2 := p.parseNode()

	e = Expression{Functor: op, Args: []Node{n1, n2}}

	for p.tok.IsOperator() {
		op := p.parseOperator()
		n3 := p.parseNode()
		e = merge(e, op, n3)
	}

	tok, _ := p.expectOneOf(terminators...)
	// TODO: clean up this more logic
	if tok == COMMA {
		return e, true
	}
	return e, false
}

func (p *parser) parseNode() (n Node) {
	switch p.tok {
	case IDENT, INT, FLOAT, STRING:
		n = p.parseTerm(p.tok, p.lit)
		p.next()
	case LPAREN:
		p.next()
		n, _ = p.parseExpressionTree(RPAREN)
	}
	return n
}

func (p *parser) parseOperator() string {
	if !p.tok.IsOperator() {
		panic(fmt.Sprintf("expected operator but found %v", p.tok))
	}
	op := p.tok.String()
	p.next()
	return op
}

// assumption: tree is binary (has two children!)
func merge(tree Expression, op string, node Node) Expression {
	root := lookupOperator(tree.Functor).Precedence()
	opp := lookupOperator(op).Precedence()
	if root >= opp {
		return Expression{Functor: op, Args: []Node{tree, node}}
	}
	left := tree.Args[0]
	right := tree.Args[1]
	newTree := Expression{Functor: op, Args: []Node{right, node}}
	return Expression{Functor: tree.Functor, Args: []Node{left, newTree}}
}

func (p *parser) parseFact() (e Expression, more bool) {
	// parse relation, looks like a rule call
	// scanner is already looking 1 rune ahead
	if p.tok == IDENT && p.scanner.ch == '(' {
		return p.parseRuleCall()
	}
	// otherwise parse an object instantiation
	return p.parseObjectInstantiation()
}

func (p *parser) parseObjectInstantiation() (oi Expression, more bool) {
	varName := p.expect(IDENT)
	p.expect(COLON)
	objectName := p.expect(IDENT)
	o := ObjectTerm(varName, objectName)
	oi = Expression{Functor: "new", Args: []Node{o}}
	p.expect(LBRACE)
	for {
		fieldName := p.expect(IDENT)
		p.expect(COLON)
		tok, lit := p.expectOneOf(IDENT, INT, FLOAT, STRING)
		f := FieldTerm(fieldName, tokenValue(tok, lit))
		oi.Args = append(oi.Args, f)
		if !p.commaOrRbrace() {
			break
		}
	}
	return oi, p.commaOrRbrace()
}

func (p *parser) parseRule() Rule {
	p.varsInScope = map[string]string{}
	ruleName := p.expect(IDENT)
	r := Rule{Name: ruleName}
	p.expectSequence(LBRACE, INPUT, LBRACE)
	for {
		name := p.expect(IDENT)
		p.expect(COLON)
		typeInfo := p.expect(IDENT)
		f := p.parseTermWithType(name, typeInfo)
		if f.TypeInfo == OBJECT {
			p.varsInScope[name] = typeInfo
		}
		r.Args = append(r.Args, f)
		if !p.commaOrRbrace() {
			break
		}
	}
	p.expectSequence(RULES, LBRACE)
	for {
		expression, more := p.parseExpression()
		r.Body = append(r.Body, expression)
		if !more {
			break
		}
	}
	p.expect(RBRACE)
	return r
}

func (p *parser) parseTest() Test {
	testName := p.expect(IDENT)
	t := Test{Name: testName}
	p.expectSequence(LBRACE, FACTS, LBRACE)
	for {
		fact, more := p.parseFact()
		t.Facts = append(t.Facts, fact)
		if !more {
			break
		}
	}
	p.expectSequence(RULES, LBRACE)
	for {
		rule, more := p.parseRuleCall()
		t.Body = append(t.Body, rule)
		if !more {
			break
		}
	}
	p.expect(RBRACE)
	return t
}

func (p *parser) parse() InternalRepresentation {
	ir := newInternalRepresentation()
	for {
		tok := p.tok
		p.next()
		switch tok {
		case ILLEGAL:
			panic("ILLEGAL")
		case EOF:
			return ir
		case OBJECT:
			o := p.parseObject()
			ir.Objects[o.Name] = o
		case RELATION:
			r := p.parseRelation()
			ir.Relations[r.Name] = r
		case RULE:
			r := p.parseRule()
			ir.Rules[r.Name] = r
		case TEST:
			t := p.parseTest()
			ir.Tests = append(ir.Tests, t)
		default:
			panic(fmt.Errorf("Error at %s", tok))
		}
	}
	return ir
}

func (p *parser) commaOrRbrace() bool {
	tok, _ := p.expectOneOf(COMMA, RBRACE)
	if tok == COMMA {
		return true
	}
	if tok == RBRACE {
		return false
	}
	panic("more failed!")
}
