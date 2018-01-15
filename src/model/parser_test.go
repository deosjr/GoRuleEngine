package model

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestParseExpression(t *testing.T) {
	for i, tt := range []struct {
		input string
		want  Expression
	}{
		{
			input: "15 + 8",
			want: Expression{Functor: "+",
				Args: []Node{
					IntTerm(15),
					IntTerm(8),
				},
			},
		},
		{
			input: "15 + 8 * 4",
			want: Expression{Functor: "+",
				Args: []Node{
					IntTerm(15),
					Expression{Functor: "*",
						Args: []Node{
							IntTerm(8),
							IntTerm(4),
						},
					},
				},
			},
		},
		{
			input: "15 * 8 + 4",
			want: Expression{Functor: "+",
				Args: []Node{
					Expression{Functor: "*",
						Args: []Node{
							IntTerm(15),
							IntTerm(8),
						},
					},
					IntTerm(4),
				},
			},
		},
		{
			input: "(15 + 8) * 4",
			want: Expression{Functor: "*",
				Args: []Node{
					Expression{Functor: "+",
						Args: []Node{
							IntTerm(15),
							IntTerm(8),
						},
					},
					IntTerm(4),
				},
			},
		},
		{
			input: `functor(arg1, arg2, 42)`,
			want: Expression{Functor: "functor",
				Args: []Node{
					IdentifierTerm("arg1"),
					IdentifierTerm("arg2"),
					IntTerm(42),
				},
			},
		},
	} {
		// add a }, ends the expression
		p := newParser(strings.NewReader(fmt.Sprintf("%s}", tt.input)))
		got, _ := p.parseExpression()
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%d): got %#v want %#v", i, got, tt.want)
		}
	}
}

func TestReadObject(t *testing.T) {
	for i, tt := range []struct {
		input string
		want  Object
	}{
		{
			input: `
			object prisoner {
				age  : int,
				name : string
			}`,
			want: NewObject("prisoner", []Field{
				{"age", INT},
				{"name", STRING},
			}),
		},
	} {
		ir := Read(tt.input)
		got, ok := ir.Objects[tt.want.Name]
		if !ok {
			t.Errorf("%d): object %q not found in %#v", i, tt.want.Name, ir)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%d): got %#v want %#v", i, got, tt.want)
		}
	}
}

func TestReadRule(t *testing.T) {
	for i, tt := range []struct {
		input string
		want  Rule
	}{
		{
			input: `
			rule hasRightToPhonecall { 
				input {
					s : prisoner
				}
				rules {
					s.age >= 18
				}
			}`,
			want: Rule{
				Name: "hasRightToPhonecall",
				Args: []Term{ObjectTerm("s", "prisoner")},
				Body: []Expression{
					{Functor: ">=",
						Args: []Node{
							Expression{Functor: ".",
								Args: []Node{
									ObjectTerm("s", "prisoner"),
									IdentifierTerm("age"),
								},
							},
							IntTerm(18),
						},
					},
				},
			},
		},
	} {
		ir := Read(tt.input)
		got, ok := ir.Rules[tt.want.Name]
		if !ok {
			t.Errorf("%d): rule %q not found in %#v", i, tt.want.Name, ir)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%d): got %#v want %#v", i, got, tt.want)
		}
	}
}

func TestReadTestcase(t *testing.T) {
	for i, tt := range []struct {
		input string
		want  Test
	}{
		{
			input: `
			test "Right to phonecall" {
				facts {
					prisonerVarName : prisoner {
						age:  23,
						name: john
					}
				}
				rules {
					hasRightToPhonecall(prisonerVarName)
				}
			}`,
			want: Test{
				Name: "Right to phonecall",
				Facts: []Expression{
					{Functor: "new",
						Args: []Node{
							ObjectTerm("prisonerVarName", "prisoner"),
							FieldTerm("age", 23),
							FieldTerm("name", "john"),
						},
					},
				},
				Body: []Expression{
					{Functor: "hasRightToPhonecall",
						Args: []Node{
							IdentifierTerm("prisonerVarName"),
						},
					},
				},
			},
		},
	} {
		ir := Read(tt.input)
		if len(ir.Tests) != 1 {
			wantName := tt.want.Name
			t.Errorf("%d): test %q not found in %#v", i, wantName, ir)
			continue
		}
		got := ir.Tests[0]
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%d): got %#v want %#v", i, got, tt.want)
		}
	}
}
