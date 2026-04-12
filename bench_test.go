package anyexpr

import "testing"

type benchEnv struct {
	Name    string
	Body    string
	Subject string
	Email   string
	Tags    []string
}

func BenchmarkMatch_SimpleEquality(b *testing.B) {
	c, _ := NewCompiler[benchEnv]()
	prog, _ := c.Compile(NewSource("bench", `eq(Name, "alice")`))
	e := benchEnv{Name: "Alice"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		prog.Match(e)
	}
}

func BenchmarkMatch_SubstringCheck(b *testing.B) {
	c, _ := NewCompiler[benchEnv]()
	prog, _ := c.Compile(NewSource("bench", `has(Body, "invoice")`))
	e := benchEnv{Body: "Please find the attached invoice for Q4"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		prog.Match(e)
	}
}

func BenchmarkMatch_RegexPattern(b *testing.B) {
	c, _ := NewCompiler[benchEnv]()
	prog, _ := c.Compile(NewSource("bench", `re(Subject, "^Re:")`))
	e := benchEnv{Subject: "Re: Quarterly Report"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		prog.Match(e)
	}
}

func BenchmarkMatch_ComplexComposite(b *testing.B) {
	c, _ := NewCompiler[benchEnv]()
	prog, _ := c.Compile(NewSource("bench",
		`has(Name, "alice") && (starts(Subject, "re:") || ends(Email, "example.com"))`))
	e := benchEnv{
		Name:    "Alice Smith",
		Subject: "Re: Hello",
		Email:   "alice@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		prog.Match(e)
	}
}

func BenchmarkEval_Extract(b *testing.B) {
	c, _ := NewCompiler[benchEnv]()
	prog, _ := c.Compile(NewSource("bench", `extract(Body, "INV-\\d+")`))
	e := benchEnv{Body: "Please pay INV-12345 by Friday"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		prog.Eval(e)
	}
}

func BenchmarkEval_Domain(b *testing.B) {
	c, _ := NewCompiler[benchEnv]()
	prog, _ := c.Compile(NewSource("bench", `email_domain(Email)`))
	e := benchEnv{Email: "alice@example.com"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		prog.Eval(e)
	}
}

func BenchmarkCompile(b *testing.B) {
	c, _ := NewCompiler[benchEnv]()
	src := NewSource("bench", `has(Name, "alice") && starts(Subject, "re:")`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Compile(src)
	}
}
