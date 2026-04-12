package anyexpr

import (
	"path/filepath"
	"regexp"
	"strings"
)

// builtinNames returns the set of all built-in function names.
// Used for conflict detection during registration.
func builtinNames() map[string]bool {
	return map[string]bool{
		"has": true, "starts": true, "ends": true, "eq": true,
		"xhas": true, "xstarts": true, "xends": true,
		"re": true, "xre": true, "glob": true,
		"lower": true, "upper": true, "trim": true,
		"words": true, "lines": true,
		"extract": true, "domain": true,
	}
}

// --- Case-insensitive ---

func biHas(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func biStarts(s, prefix string) bool {
	return strings.HasPrefix(strings.ToLower(s), strings.ToLower(prefix))
}

func biEnds(s, suffix string) bool {
	return strings.HasSuffix(strings.ToLower(s), strings.ToLower(suffix))
}

func biEq(a, b string) bool {
	return strings.EqualFold(a, b)
}

// --- Case-sensitive ---

func biXhas(s, substr string) bool {
	return strings.Contains(s, substr)
}

func biXstarts(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

func biXends(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// --- Pattern matching ---

func biRe(s, pattern string) bool {
	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

func biXre(s, pattern string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

func biGlob(s, pattern string) bool {
	ok, _ := filepath.Match(strings.ToLower(pattern), strings.ToLower(s))
	return ok
}

// --- Transformation ---

func biLower(s string) string { return strings.ToLower(s) }
func biUpper(s string) string { return strings.ToUpper(s) }
func biTrim(s string) string  { return strings.TrimSpace(s) }

func biWords(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Fields(s)
}

func biLines(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, "\n")
}

// --- Extraction ---

func biExtract(s, pattern string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return ""
	}
	return re.FindString(s)
}

func biDomain(addr string) string {
	idx := strings.LastIndex(addr, "@")
	if idx < 0 {
		return ""
	}
	return addr[idx+1:]
}
