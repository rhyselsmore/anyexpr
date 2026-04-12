package rules

import "fmt"

// --- Registry options ---

// RegistryOpt configures a Registry.
type RegistryOpt func(*registryConfig) error

type registryConfig struct {
	actions  map[string]Def
	handlers map[string]any
}

// WithAction registers a named action type on the registry.
func WithAction(name string, c Cardinality, v ValueKind, terminal bool) RegistryOpt {
	return func(cfg *registryConfig) error {
		if _, exists := cfg.actions[name]; exists {
			return fmt.Errorf("%w: %q", ErrDuplicateRegistration, name)
		}
		cfg.actions[name] = Def{
			Name:        name,
			Terminal:    terminal,
			Cardinality: c,
			Value:       v,
		}
		return nil
	}
}

// WithHandler registers a named handler on the registry. The handler
// function is stored type-erased. The Engine[T, V] constructor
// type-asserts it at build time against func(*Context[T, V]) error.
func WithHandler(name string, h any, c Cardinality, terminal bool) RegistryOpt {
	return func(cfg *registryConfig) error {
		if _, exists := cfg.actions[name]; exists {
			return fmt.Errorf("%w: %q", ErrDuplicateRegistration, name)
		}
		if _, exists := cfg.handlers[name]; exists {
			return fmt.Errorf("%w: %q", ErrDuplicateRegistration, name)
		}
		cfg.actions[name] = Def{
			Name:        name,
			Terminal:    terminal,
			Cardinality: c,
			Value:       NoValue,
			IsHandler:   true,
		}
		cfg.handlers[name] = h
		return nil
	}
}

// --- Compile options ---

// CompileOpt configures a Compile call. Reserved for future use.
type CompileOpt func(*compileConfig)

type compileConfig struct{}

// --- Engine options ---

// EngineOpt configures an Engine.
type EngineOpt func(*engineConfig)

type engineConfig struct {
	sel selector
}

// WithTags limits the engine to rules with at least one matching tag.
func WithTags(tags ...string) EngineOpt {
	return func(cfg *engineConfig) {
		if cfg.sel.onlyTags == nil {
			cfg.sel.onlyTags = make(map[string]bool)
		}
		for _, t := range tags {
			cfg.sel.onlyTags[t] = true
		}
	}
}

// WithNames limits the engine to rules with matching names.
func WithNames(names ...string) EngineOpt {
	return func(cfg *engineConfig) {
		if cfg.sel.onlyNames == nil {
			cfg.sel.onlyNames = make(map[string]bool)
		}
		for _, n := range names {
			cfg.sel.onlyNames[n] = true
		}
	}
}

// WithExcludeTags excludes rules with any of the given tags.
func WithExcludeTags(tags ...string) EngineOpt {
	return func(cfg *engineConfig) {
		if cfg.sel.excludeTags == nil {
			cfg.sel.excludeTags = make(map[string]bool)
		}
		for _, t := range tags {
			cfg.sel.excludeTags[t] = true
		}
	}
}

// WithExcludeNames excludes rules with any of the given names.
func WithExcludeNames(names ...string) EngineOpt {
	return func(cfg *engineConfig) {
		if cfg.sel.excludeNames == nil {
			cfg.sel.excludeNames = make(map[string]bool)
		}
		for _, n := range names {
			cfg.sel.excludeNames[n] = true
		}
	}
}

// --- Run options ---

// RunOpt configures a single Run or DryRun call.
type RunOpt func(*runConfig)

type runConfig struct {
	sel selector
}

// OnlyTags limits this execution to rules with at least one matching tag.
func OnlyTags(tags ...string) RunOpt {
	return func(cfg *runConfig) {
		if cfg.sel.onlyTags == nil {
			cfg.sel.onlyTags = make(map[string]bool)
		}
		for _, t := range tags {
			cfg.sel.onlyTags[t] = true
		}
	}
}

// OnlyNames limits this execution to rules with matching names.
func OnlyNames(names ...string) RunOpt {
	return func(cfg *runConfig) {
		if cfg.sel.onlyNames == nil {
			cfg.sel.onlyNames = make(map[string]bool)
		}
		for _, n := range names {
			cfg.sel.onlyNames[n] = true
		}
	}
}

// ExcludeTags excludes rules with any of the given tags for this execution.
func ExcludeTags(tags ...string) RunOpt {
	return func(cfg *runConfig) {
		if cfg.sel.excludeTags == nil {
			cfg.sel.excludeTags = make(map[string]bool)
		}
		for _, t := range tags {
			cfg.sel.excludeTags[t] = true
		}
	}
}

// ExcludeNames excludes rules with any of the given names for this execution.
func ExcludeNames(names ...string) RunOpt {
	return func(cfg *runConfig) {
		if cfg.sel.excludeNames == nil {
			cfg.sel.excludeNames = make(map[string]bool)
		}
		for _, n := range names {
			cfg.sel.excludeNames[n] = true
		}
	}
}
