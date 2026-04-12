package anyexpr_test

import (
	"fmt"
	"log"

	"github.com/rhyselsmore/anyexpr"
)

func Example() {
	type Email struct {
		From    string
		Subject string
		Body    string
	}

	compiler, err := anyexpr.NewCompiler[Email]()
	if err != nil {
		log.Fatal(err)
	}

	src := anyexpr.NewSource("invoice-filter",
		`has(Subject, "invoice") && ends(From, "stripe.com")`)

	prog, err := compiler.Compile(src)
	if err != nil {
		log.Fatal(err)
	}

	msg := Email{
		From:    "billing@stripe.com",
		Subject: "Your January Invoice",
		Body:    "...",
	}

	matched, err := prog.Match(msg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(matched)
	// Output: true
}

func Example_eval() {
	type Email struct {
		From  string
		Email string
	}

	compiler, err := anyexpr.NewCompiler[Email]()
	if err != nil {
		log.Fatal(err)
	}

	prog, err := compiler.Compile(
		anyexpr.NewSource("extract-domain", `domain(Email)`))
	if err != nil {
		log.Fatal(err)
	}

	result, err := prog.Eval(Email{
		From:  "Alice",
		Email: "alice@example.com",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
	// Output: example.com
}

func Example_check() {
	type Env struct {
		Name string
	}

	compiler, err := anyexpr.NewCompiler[Env]()
	if err != nil {
		log.Fatal(err)
	}

	sources := []*anyexpr.Source{
		anyexpr.NewSource("rule-1", `has(Name, "alice")`),
		anyexpr.NewSource("rule-2", `starts(Name, "b")`),
	}

	if err := compiler.Check(sources); err != nil {
		log.Fatal(err)
	}
	fmt.Println("all valid")
	// Output: all valid
}

func Example_customFunction() {
	type Env struct {
		Value string
	}

	compiler, err := anyexpr.NewCompiler[Env](
		anyexpr.WithFunction("reverse", func(params ...any) (any, error) {
			s := params[0].(string)
			runes := []rune(s)
			for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
				runes[i], runes[j] = runes[j], runes[i]
			}
			return string(runes), nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	prog, err := compiler.Compile(
		anyexpr.NewSource("palindrome", `eq(Value, reverse(Value))`))
	if err != nil {
		log.Fatal(err)
	}

	matched, _ := prog.Match(Env{Value: "racecar"})
	fmt.Println(matched)
	// Output: true
}
