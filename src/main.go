package main

import (
	"model"
	"prolog"
)

/*

1. Read DSL -> internal representation
1.1 do type checking
2. Generate Prolog
3. Execute Prolog tests

Endgoal:
	- run not as a cmdline program
	but as a rest api taking
	DSL input and returning model
	validity and possibly a program
	in Prolog or even Mercury.

	- interactive creation of rule set
	using test-driven development
	validated by logic programming.

*/

func main() {

	// 1. Read DSL

	// TODO: 18 + 1 (eval) (golog does this in-place?)
	// TODO: difference between facts(relations with arity?) and rules
	// TODO: if-then-else
	// TODO: support for-all and exists
	// TODO: inheritance
	// TODO: nested object instantiation in tests

	// stupid contrived example: money for legs
	// chickens have 2 legs, sheep have 4 (always!)
	// difference between field (variable) and fact

	// you get $2 per leg
	// moneyForLegs(animal::animal, money::int) :-
	// 	  money == animal.legs * 2

	// animal ---> chicken ; sheep (how to represent this?)
	// legs(A::animal, Legs::int) :-
	//	  (A is-a sheep ; Legs = 4);
	//	  (A is-a chicken ; Legs = 2).
	// moneyForLegs(Animal::animal, Money::int) :-
	//    legs(Animal, Legs),
	//    Money = Legs * 2.

	// All these colons might not be necessary,
	// but I think it makes it more readable?

	s := `
		object prisoner {
			age  : int,
			name : string
		}
		# TODO: arity?
		relation cellmates {
			p : prisoner,
			cellmate : prisoner
		}
		rule hasRightToPhonecall { 
			input {
				p : prisoner
			}
			rules {
				p.age >= 18
			}
		}
		test "Right to phonecall" {
			facts {
				p1 : prisoner {
					age:  23,
					name: john #string quotes are optional, needed for spaces
				},
				p2 : prisoner {
					age: 15,
					name: henry
				}
				# these need to be asserted in test?
				#cellmates(p1, p2),
				#cellmates(p2, p1)
			}
			rules {
				hasRightToPhonecall(p1)
				#hasRightToPhonecall(p2) # TODO: not/1
			}
		}
	`
	ir := model.Read(s)

	// 2. Generate Prolog

	m := prolog.Generate(ir)

	// 3. Execute Prolog tests

	prolog.TestRulebase(m)
}
