package rules

// Context is passed to handlers during execution. T is the environment
// type (the thing being evaluated). V is the domain-specific vars type
// (attachment loaders, DB connections, etc.).
type Context[T, V any] struct {
	Env     T
	Actions ResolvedActions
	Vars    V
}
