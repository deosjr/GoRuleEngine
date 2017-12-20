package model

import "strings"

// DSL level

type Object struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name     string
	TypeInfo Token
}

type Relation struct {
	Name   string
	Fields []Field
	// TODO arity?
}

type Rule struct {
	Name string
	Args []Term
	Body []Expression // rules and truth statements
}

type Test struct {
	Name  string
	Facts []Expression // TODO: Object instantiations and Relations!
	Body  []Expression // only rule calls ?
}

// an expression is either a
// - rule call (which evaluates to boolean)
// - comparison
// - object instantiation (name = new X), ONLY IN TESTS
// all of these fit the functor(args) pattern

// observation for Mercury: rule calls
// will always have mode ALL args as input, determinate

// expression tree, internal nodes are functions
// leaf nodes are typed terms
type Node interface {
	todo()
}

type Expression struct {
	Functor string
	Args    []Node
}

type Term struct {
	Value     interface{}
	TypeInfo  Token
	fieldInfo string
}

// TODO: what makes term/expression a node interface?
func (t Term) todo()       {}
func (e Expression) todo() {}

func (t Term) ObjectName() string {
	if t.TypeInfo != OBJECT {
		panic("getting objectName of non-object")
	}
	return t.fieldInfo
}

func (t Term) FieldName() string {
	if t.TypeInfo != IDENT {
		panic("getting fieldName of non-field")
	}
	return t.fieldInfo
}

func StringTerm(value string) Term {
	return Term{Value: value, TypeInfo: STRING}
}
func IntTerm(value int) Term {
	return Term{Value: value, TypeInfo: INT}
}
func ObjectTerm(ident, objectName string) Term {
	return Term{Value: ident, TypeInfo: OBJECT, fieldInfo: objectName}
}
func FieldTerm(fieldName string, value interface{}) Term {
	return Term{Value: value, TypeInfo: IDENT, fieldInfo: fieldName}
}
func IdentifierTerm(literal string) Term {
	return Term{Value: literal, TypeInfo: IDENT}
}

func NewObject(name string, fields []Field) Object {
	return Object{Name: name, Fields: fields}
}

type InternalRepresentation struct {
	Objects   map[string]Object
	Relations map[string]Relation
	Rules     map[string]Rule
	Tests     []Test
}

func newInternalRepresentation() InternalRepresentation {
	return InternalRepresentation{
		Objects:   map[string]Object{},
		Relations: map[string]Relation{},
		Rules:     map[string]Rule{},
		Tests:     []Test{},
	}
}

func Read(s string) InternalRepresentation {
	p := newParser(strings.NewReader(s))
	return p.parse()
}
