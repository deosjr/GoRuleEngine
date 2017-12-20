package prolog

import (
	"strings"
	"testing"

	. "model"
)

func TestPrintObject(t *testing.T) {
	g := &generator{
		objectMap: map[string]string{
			"prisoner": "o_1",
		},
	}

	for i, tt := range []struct {
		object Object
		want   string
	}{
		{
			object: NewObject("prisoner", []Field{
				{"age", INT},
				{"name", STRING},
			}),
			want: `prisoner_age(A, B) :- B = prisoner(A,_).
					prisoner_name(A, B) :- B = prisoner(_,A).`,
		},
	} {
		got := printObject(g, tt.object)
		helperFunc(t, i, got, tt.want)
	}
}

func TestPrintRule(t *testing.T) {
	g := &generator{
		objectMap: map[string]string{
			"prisoner": "o_1",
		},
		n: 1,
	}

	for i, tt := range []struct {
		rule Rule
		want string
	}{
		{
			rule: Rule{
				Name: "hasRightToPhonecall",
				Args: []Term{ObjectTerm("prisonerVarName", "prisoner")},
				Body: []Expression{
					{Functor: ">=",
						Args: []Node{
							Expression{Functor: ".",
								Args: []Node{
									ObjectTerm("prisonerVarName", "prisoner"),
									IdentifierTerm("age"),
								},
							},
							IntTerm(18),
						},
					},
				},
			},
			want: `hasRightToPhonecall(PrisonerVarName) :- 
					o_1_age(V_2, PrisonerVarName),
					@>=(V_2,18).`,
		},
	} {
		got := printRule(g, tt.rule)
		helperFunc(t, i, got, tt.want)
	}
}

func TestPrintTest(t *testing.T) {
	g := &generator{
		objectMap: map[string]string{
			"prisoner": "o_1",
		},
		ir: InternalRepresentation{
			Objects: map[string]Object{
				"prisoner": NewObject("prisoner", []Field{
					{"age", INT},
					{"name", STRING},
				}),
			},
		},
	}

	for i, tt := range []struct {
		test Test
		want string
	}{
		{
			test: Test{
				Name: "Prisoner has right to phonecall",
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
			want: `test('Prisoner has right to phonecall') :- 
					PrisonerVarName = o_1(23,'john'),
					hasRightToPhonecall(PrisonerVarName).`,
		},
	} {
		got := printTest(g, tt.test)
		helperFunc(t, i, got, tt.want)
	}
}

func helperFunc(t *testing.T, i int, got, want string) {
	// TODO: upgrade to go1.9
	//t.Helper()
	got = strings.Replace(got, "\n", "", -1)
	got = strings.Replace(got, "\t", "", -1)
	want = strings.Replace(want, "\n", "", -1)
	want = strings.Replace(want, "\t", "", -1)
	if got != want {
		t.Errorf("%d): got %s want %s", i, got, want)
	}
}
