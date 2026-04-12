package anyexpr

import "testing"

func TestBiHas(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{"exact", "Hello World", "Hello World", true},
		{"case insensitive", "Hello World", "hello world", true},
		{"substring", "Hello World", "lo Wo", true},
		{"no match", "Hello World", "xyz", false},
		{"empty substr", "Hello World", "", true},
		{"empty s", "", "hello", false},
		{"both empty", "", "", true},
		{"unicode", "Ünïcödé", "ünïcödé", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := biHas(tt.s, tt.substr); got != tt.want {
				t.Errorf("biHas(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

func TestBiStarts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		s      string
		prefix string
		want   bool
	}{
		{"exact", "Hello", "Hello", true},
		{"case insensitive", "Hello World", "hello", true},
		{"no match", "Hello World", "World", false},
		{"empty prefix", "Hello", "", true},
		{"empty s", "", "hello", false},
		{"both empty", "", "", true},
		{"unicode", "Ünïcödé", "ünï", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := biStarts(tt.s, tt.prefix); got != tt.want {
				t.Errorf("biStarts(%q, %q) = %v, want %v", tt.s, tt.prefix, got, tt.want)
			}
		})
	}
}

func TestBiEnds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		s      string
		suffix string
		want   bool
	}{
		{"exact", "Hello", "Hello", true},
		{"case insensitive", "Hello World", "world", true},
		{"no match", "Hello World", "Hello", false},
		{"empty suffix", "Hello", "", true},
		{"empty s", "", "hello", false},
		{"both empty", "", "", true},
		{"unicode", "Ünïcödé", "ödé", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := biEnds(tt.s, tt.suffix); got != tt.want {
				t.Errorf("biEnds(%q, %q) = %v, want %v", tt.s, tt.suffix, got, tt.want)
			}
		})
	}
}

func TestBiEq(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{"exact", "hello", "hello", true},
		{"case insensitive", "Hello", "hello", true},
		{"no match", "hello", "world", false},
		{"both empty", "", "", true},
		{"one empty", "hello", "", false},
		{"unicode", "Ünïcödé", "ünïcödé", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := biEq(tt.a, tt.b); got != tt.want {
				t.Errorf("biEq(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestBiXhas(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{"exact", "Hello World", "Hello World", true},
		{"case sensitive miss", "Hello World", "hello world", false},
		{"substring", "Hello World", "lo Wo", true},
		{"no match", "Hello World", "xyz", false},
		{"empty substr", "Hello World", "", true},
		{"empty s", "", "hello", false},
		{"both empty", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := biXhas(tt.s, tt.substr); got != tt.want {
				t.Errorf("biXhas(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

func TestBiXstarts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		s      string
		prefix string
		want   bool
	}{
		{"exact", "Hello", "Hello", true},
		{"case sensitive miss", "Hello", "hello", false},
		{"no match", "Hello World", "World", false},
		{"empty prefix", "Hello", "", true},
		{"empty s", "", "hello", false},
		{"both empty", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := biXstarts(tt.s, tt.prefix); got != tt.want {
				t.Errorf("biXstarts(%q, %q) = %v, want %v", tt.s, tt.prefix, got, tt.want)
			}
		})
	}
}

func TestBiXends(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		s      string
		suffix string
		want   bool
	}{
		{"exact", "Hello", "Hello", true},
		{"case sensitive miss", "Hello World", "WORLD", false},
		{"match", "Hello World", "World", true},
		{"empty suffix", "Hello", "", true},
		{"empty s", "", "hello", false},
		{"both empty", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := biXends(tt.s, tt.suffix); got != tt.want {
				t.Errorf("biXends(%q, %q) = %v, want %v", tt.s, tt.suffix, got, tt.want)
			}
		})
	}
}

func TestBiRe(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		s       string
		pattern string
		want    bool
	}{
		{"match", "Hello World", "hello", true},
		{"case insensitive", "Hello World", "HELLO", true},
		{"regex", "Hello World", "^hello.*world$", true},
		{"no match", "Hello World", "^goodbye", false},
		{"bad regex", "Hello", "[invalid", false},
		{"empty string", "", ".*", true},
		{"empty pattern", "Hello", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := biRe(tt.s, tt.pattern); got != tt.want {
				t.Errorf("biRe(%q, %q) = %v, want %v", tt.s, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestBiXre(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		s       string
		pattern string
		want    bool
	}{
		{"match", "Hello World", "Hello", true},
		{"case sensitive miss", "Hello World", "hello", false},
		{"regex", "Hello World", "^Hello.*World$", true},
		{"no match", "Hello World", "^goodbye", false},
		{"bad regex", "Hello", "[invalid", false},
		{"empty string", "", ".*", true},
		{"empty pattern", "Hello", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := biXre(tt.s, tt.pattern); got != tt.want {
				t.Errorf("biXre(%q, %q) = %v, want %v", tt.s, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestBiGlob(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		s       string
		pattern string
		want    bool
	}{
		{"match", "hello.txt", "*.txt", true},
		{"case insensitive", "Hello.TXT", "*.txt", true},
		{"no match", "hello.txt", "*.go", false},
		{"question mark", "hello", "hell?", true},
		{"exact", "hello", "hello", true},
		{"empty both", "", "", true},
		{"empty pattern", "hello", "", false},
		{"empty s", "", "*", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := biGlob(tt.s, tt.pattern); got != tt.want {
				t.Errorf("biGlob(%q, %q) = %v, want %v", tt.s, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestBiLower(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"mixed", "Hello World", "hello world"},
		{"already lower", "hello", "hello"},
		{"empty", "", ""},
		{"unicode", "ÜNÏ", "ünï"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := biLower(tt.s); got != tt.want {
				t.Errorf("biLower(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestBiUpper(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"mixed", "Hello World", "HELLO WORLD"},
		{"already upper", "HELLO", "HELLO"},
		{"empty", "", ""},
		{"unicode", "üni", "ÜNI"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := biUpper(tt.s); got != tt.want {
				t.Errorf("biUpper(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestBiTrim(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"spaces", "  hello  ", "hello"},
		{"tabs", "\thello\t", "hello"},
		{"newlines", "\nhello\n", "hello"},
		{"already trimmed", "hello", "hello"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := biTrim(tt.s); got != tt.want {
				t.Errorf("biTrim(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestBiWords(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		s    string
		want []string
	}{
		{"multiple", "hello world foo", []string{"hello", "world", "foo"}},
		{"single", "hello", []string{"hello"}},
		{"empty", "", []string{}},
		{"extra spaces", "  hello   world  ", []string{"hello", "world"}},
		{"tabs", "hello\tworld", []string{"hello", "world"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := biWords(tt.s)
			if len(got) != len(tt.want) {
				t.Fatalf("biWords(%q) = %v, want %v", tt.s, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("biWords(%q)[%d] = %q, want %q", tt.s, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestBiLines(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		s    string
		want []string
	}{
		{"multiple", "a\nb\nc", []string{"a", "b", "c"}},
		{"single", "hello", []string{"hello"}},
		{"empty", "", []string{}},
		{"trailing newline", "a\nb\n", []string{"a", "b", ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := biLines(tt.s)
			if len(got) != len(tt.want) {
				t.Fatalf("biLines(%q) = %v, want %v", tt.s, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("biLines(%q)[%d] = %q, want %q", tt.s, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestBiExtract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		s       string
		pattern string
		want    string
	}{
		{"match", "Invoice INV-123 received", `INV-\d+`, "INV-123"},
		{"no match", "Hello World", `\d+`, ""},
		{"bad regex", "Hello", "[invalid", ""},
		{"empty string", "", `\d+`, ""},
		{"empty pattern", "Hello", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := biExtract(tt.s, tt.pattern); got != tt.want {
				t.Errorf("biExtract(%q, %q) = %q, want %q", tt.s, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestBiDomain(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		addr string
		want string
	}{
		{"normal email", "user@example.com", "example.com"},
		{"no at", "userexample.com", ""},
		{"multiple at", "user@sub@example.com", "example.com"},
		{"empty", "", ""},
		{"at only", "@", ""},
		{"at start", "@example.com", "example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := biDomain(tt.addr); got != tt.want {
				t.Errorf("biDomain(%q) = %q, want %q", tt.addr, got, tt.want)
			}
		})
	}
}
