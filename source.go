package anyexpr

// Source is the input to compilation. It pairs a name with an expression.
type Source struct {
	name string
	expr string
}

// SourceOpt configures a Source. Reserved for future use.
type SourceOpt func(*sourceConfig)

type sourceConfig struct{}

// NewSource creates a new Source. The name is used in error messages,
// logging, and tracing. The expr is the expression string to compile.
func NewSource(name, expr string, opts ...SourceOpt) *Source {
	_ = opts // reserved for future use
	return &Source{name: name, expr: expr}
}

// Name returns the source name.
func (s *Source) Name() string { return s.name }

// Expr returns the expression string.
func (s *Source) Expr() string { return s.expr }
