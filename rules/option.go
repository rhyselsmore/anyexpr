package rules

// --- Evaluator options ---

// EvaluatorOpt configures an Evaluator.
type EvaluatorOpt func(*evaluatorConfig)

type evaluatorConfig struct {
	evalDefaults evaluationConfig
}

// OnEvaluation sets default evaluation options applied to every Run
// call. Per-call options passed to Run clobber these defaults.
func OnEvaluation(opts ...EvaluationOpt) EvaluatorOpt {
	return func(cfg *evaluatorConfig) {
		for _, o := range opts {
			o(&cfg.evalDefaults)
		}
	}
}

// --- Evaluation options ---

// EvaluationOpt configures a single evaluation (Run call).
type EvaluationOpt func(*evaluationConfig)

type evaluationConfig struct {
	sel selector
}

// WithTags limits evaluation to rules with at least one matching tag.
func WithTags(tags ...string) EvaluationOpt {
	return func(cfg *evaluationConfig) {
		if cfg.sel.onlyTags == nil {
			cfg.sel.onlyTags = make(map[string]bool)
		}
		for _, t := range tags {
			cfg.sel.onlyTags[t] = true
		}
	}
}

// WithNames limits evaluation to rules with matching names.
func WithNames(names ...string) EvaluationOpt {
	return func(cfg *evaluationConfig) {
		if cfg.sel.onlyNames == nil {
			cfg.sel.onlyNames = make(map[string]bool)
		}
		for _, n := range names {
			cfg.sel.onlyNames[n] = true
		}
	}
}

// ExcludeTags excludes rules with any of the given tags.
func ExcludeTags(tags ...string) EvaluationOpt {
	return func(cfg *evaluationConfig) {
		if cfg.sel.excludeTags == nil {
			cfg.sel.excludeTags = make(map[string]bool)
		}
		for _, t := range tags {
			cfg.sel.excludeTags[t] = true
		}
	}
}

// ExcludeNames excludes rules with any of the given names.
func ExcludeNames(names ...string) EvaluationOpt {
	return func(cfg *evaluationConfig) {
		if cfg.sel.excludeNames == nil {
			cfg.sel.excludeNames = make(map[string]bool)
		}
		for _, n := range names {
			cfg.sel.excludeNames[n] = true
		}
	}
}

// --- Compile options ---

// CompileOpt configures a Compile call. Reserved for future use.
type CompileOpt func(*compileConfig)

type compileConfig struct{}

// --- Merge options ---

// MergeOpt configures a Merge call.
type MergeOpt func(*mergeConfig)

type mergeConfig struct {
	allowOverride bool
}

// AllowOverride permits the second ruleset to replace rules with
// colliding names, keeping the original's position in evaluation order.
var AllowOverride MergeOpt = func(cfg *mergeConfig) { cfg.allowOverride = true }
