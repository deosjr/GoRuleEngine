package prolog

import (
	"fmt"
	"strings"

	. "model"

	"github.com/mndrix/golog"
)

type generator struct {
	n         int
	ir        InternalRepresentation
	objectMap map[string]string
}

func (g *generator) nextInt() int {
	g.n++
	return g.n
}

func (g *generator) newVarName() string {
	return fmt.Sprintf("V_%d", g.nextInt())
}

func (g *generator) objectMapping(external string) string {
	prologName := fmt.Sprintf("o_%d", g.nextInt())
	g.objectMap[external] = prologName
	return prologName
}

func printObject(g *generator, o Object) string {
	s := ""
	arity := len(o.Fields)
	for i, f := range o.Fields {
		s += printField(g, f, o.Name, i, arity)
	}
	return s
}

func printField(g *generator, f Field, objectName string, n, arity int) string {
	underscores := make([]string, arity)
	for i := range underscores {
		if i == n {
			underscores[i] = "A"
			continue
		}
		underscores[i] = "_"
	}
	return fmt.Sprintf("%s_%s(A, B) :- B = %s(%s).\n",
		objectName, f.Name, objectName, strings.Join(underscores, ","))
}

func printRule(g *generator, r Rule) string {
	args := ""
	if len(r.Args) != 0 {
		a := make([]string, len(r.Args))
		for i, v := range r.Args {
			a[i] = printTerm(v)
		}
		args = "(" + strings.Join(a, ",") + ")"
	}
	head := fmt.Sprintf("%s%s", strings.Replace(r.Name, " ", "_", -1), args)

	if len(r.Body) == 0 {
		return head + "."
	}
	body := make([]string, len(r.Body))
	for i, v := range r.Body {
		body[i] = printNode(g, v)
	}
	return fmt.Sprintf("%s :- \n\t%s.",
		head, strings.Join(body, ",\n\t"))
}

func printNode(g *generator, n Node) string {
	switch v := n.(type) {
	case Term:
		return printTerm(v)
	case Expression:
		expression, sideEffects := printExpression(g, v)
		return strings.Join(append(sideEffects, expression), ",\n\t")
	default:
		panic("expected node to be term or expression")
	}
}

// preRequisites need to be printed BEFORE the actual string
func printNodeRecursive(g *generator, n Node) (s string, preRequisites []string) {
	switch v := n.(type) {
	case Term:
		return printTerm(v), nil
	case Expression:
		return printExpression(g, v)
	default:
		panic("expected node to be term or expression")
	}
}

// TODO: recognise the need for and deal with parens
func printExpression(g *generator, e Expression) (string, []string) {

	// exceptions: builtins
	switch e.Functor {
	case "new":
		return printNew(g, e.Args), nil
	case ".":
		return printFieldAccessor(g, e.Args)
	case ">=":
		e.Functor = "@>="
	case ">":
		e.Functor = "@>"
	case "<=":
		e.Functor = "@=<"
	case "<":
		e.Functor = "@<"
	}

	sideEffects := []string{}
	args := make([]string, len(e.Args))
	for i, n := range e.Args {
		ve, vs := printNodeRecursive(g, n)
		args[i] = ve
		sideEffects = append(sideEffects, vs...)
	}
	return fmt.Sprintf("%s(%s)", e.Functor, strings.Join(args, ",")), sideEffects
}

// TODO: do something with typeinfo on terms
func printTerm(t Term) string {
	return printValueWithType(t.Value, t.TypeInfo)
}

func printValueWithType(v interface{}, ti Token) string {
	switch ti {
	case IDENT, OBJECT:
		return strings.Title(v.(string))
	case STRING:
		return fmt.Sprintf("'%s'", v.(string))
	}
	return fmt.Sprintf("%v", v)
}

func printTest(g *generator, t Test) string {
	header := fmt.Sprintf("test('%s')", t.Name)
	body := make([]string, len(t.Body)+len(t.Facts))
	for i, v := range t.Facts {
		body[i] = printNode(g, v)
	}
	for i, v := range t.Body {
		body[i+len(t.Facts)] = printNode(g, v)
	}
	return fmt.Sprintf("%s :- %s.", header, strings.Join(body, ","))
}

// builtin functions:
// new(object.class, varname, constructor args...)
// Varname = class(args)
func printNew(g *generator, args []Node) string {
	objectTerm := args[0].(Term)
	objectName := g.objectMap[objectTerm.ObjectName()]
	varName := strings.Title(objectTerm.Value.(string))
	object := g.ir.Objects[objectTerm.ObjectName()]
	fields := args[1:]

	// TODO: this is sloppy and needs to be optimised
	a := make([]string, len(object.Fields))
	for i, f := range object.Fields {
		a[i] = "_"
		for _, v := range fields {
			fieldTerm := v.(Term)
			if f.Name == fieldTerm.FieldName() {
				a[i] = printValueWithType(fieldTerm.Value, f.TypeInfo)
				break
			}
		}
	}
	return fmt.Sprintf("%s = %s(%s)", varName, objectName, strings.Join(a, ","))
}

// .(Soldier, age) --> {"NewlyIntroducedVarname", o_x_age(NewlyIntroducedVarname, Soldier)}
func printFieldAccessor(g *generator, args []Node) (string, []string) {
	object := args[0].(Term)
	fieldName := args[1].(Term).Value.(string)
	varName := g.newVarName()

	// TODO: recursive access ? (soldier.job.length)
	sideEffects := []string{}
	objectName := g.objectMap[object.ObjectName()]
	fieldAccess := fmt.Sprintf("%s_%s(%s, %s)",
		objectName, fieldName, varName, printTerm(object))
	sideEffects = append(sideEffects, fieldAccess)
	return varName, sideEffects
}

func Generate(ir InternalRepresentation) golog.Machine {
	g := &generator{
		ir:        ir,
		objectMap: map[string]string{},
	}
	m := golog.NewMachine()
	for _, o := range ir.Objects {
		o.Name = g.objectMapping(o.Name)
		prologString := printObject(g, o)
		fmt.Println(prologString)
		m = m.Consult(prologString)
	}
	for _, r := range ir.Rules {
		prologString := printRule(g, r)
		fmt.Println(prologString)
		m = m.Consult(prologString)
	}
	tests := []string{}
	for _, t := range ir.Tests {
		tests = append(tests, fmt.Sprintf("'%s'", t.Name))
		prologString := printTest(g, t)
		fmt.Println(prologString)
		m = m.Consult(prologString)
	}
	testCases := fmt.Sprintf("test_cases([%s]).", strings.Join(tests, ","))
	m = m.Consult(testCases)
	return m
}

func TestRulebase(m golog.Machine) {
	m = m.Consult(`
		run_tests :-
		    test_cases(List),
		    run_test_cases(List, Status),
			printf('Test %s~n', Status).

		run_test_cases([], succes).
		run_test_cases([H|T], Status) :-
		    test(H), !,
		    run_test_cases(T, Status).
		run_test_cases([H|T], failure) :-
		    \+(test(H)), !,
		    printf('Test case failed: %s~n', H),
		    run_test_cases(T, _).
	`)

	m.CanProve("run_tests.")
}
