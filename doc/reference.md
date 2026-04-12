# Built-in Function Reference

All built-in functions are available in every expression without
additional configuration. Functions prefixed with `x` are case-sensitive
variants of their counterpart.

---

## String Matching

### `has(s, substr) bool`

Returns true if `s` contains `substr`. Case-insensitive.

```
has(Subject, "invoice")          // true for "Your Invoice #123"
has(Subject, "INVOICE")          // true for "Your Invoice #123"
```

### `xhas(s, substr) bool`

Returns true if `s` contains `substr`. Case-sensitive.

```
xhas(Subject, "Invoice")        // true for "Your Invoice #123"
xhas(Subject, "invoice")        // false for "Your Invoice #123"
```

### `starts(s, prefix) bool`

Returns true if `s` starts with `prefix`. Case-insensitive.

```
starts(Name, "al")              // true for "Alice"
```

### `xstarts(s, prefix) bool`

Returns true if `s` starts with `prefix`. Case-sensitive.

```
xstarts(Name, "Al")             // true for "Alice"
xstarts(Name, "al")             // false for "Alice"
```

### `ends(s, suffix) bool`

Returns true if `s` ends with `suffix`. Case-insensitive.

```
ends(Email, "example.com")      // true for "alice@Example.COM"
```

### `xends(s, suffix) bool`

Returns true if `s` ends with `suffix`. Case-sensitive.

```
xends(Email, "example.com")     // true for "alice@example.com"
xends(Email, "example.com")     // false for "alice@Example.COM"
```

### `eq(a, b) bool`

Returns true if `a` and `b` are equal. Case-insensitive.

```
eq(Status, "active")            // true for "ACTIVE", "Active", etc.
```

---

## Pattern Matching

### `re(s, pattern) bool`

Returns true if `s` matches the regular expression `pattern`.
Case-insensitive. Invalid patterns return false.

```
re(Subject, "^re:")             // true for "Re: Hello"
re(Body, "inv-\\d+")            // true for "See INV-123"
```

### `xre(s, pattern) bool`

Returns true if `s` matches the regular expression `pattern`.
Case-sensitive. Invalid patterns return false.

```
xre(Subject, "^Re:")            // true for "Re: Hello"
xre(Subject, "^re:")            // false for "Re: Hello"
```

### `glob(s, pattern) bool`

Returns true if `s` matches the glob `pattern`. Case-insensitive.
Supports `*` (any characters) and `?` (single character).

```
glob(Filename, "*.txt")         // true for "README.txt"
glob(Name, "ali?e")             // true for "Alice"
```

---

## Transformation

### `lower(s) string`

Returns `s` converted to lowercase.

```
lower(Name) == "alice"
```

### `upper(s) string`

Returns `s` converted to uppercase.

```
upper(Code) == "USD"
```

### `trim(s) string`

Returns `s` with leading and trailing whitespace removed.

```
trim(Input) != ""
```

### `replace(s, old, new) string`

Replaces all occurrences of `old` with `new` in `s`. An optional fourth
argument limits the number of replacements.

```
replace(Name, " ", "-")         // "Alice-Smith" from "Alice Smith"
replace(Body, "foo", "bar", 1)  // replace only the first occurrence
```

### `split(s, sep) []string`

Splits `s` on `sep` and returns the resulting parts. An optional third
argument limits the number of splits.

```
split(Tags, ",")                // ["a", "b", "c"] from "a,b,c"
split(Line, ":", 2)             // split into at most 2 parts
```

### `words(s) []string`

Splits `s` on whitespace and returns the resulting words.
Returns an empty slice for empty input.

```
len(words(Title)) > 3
```

### `lines(s) []string`

Splits `s` on newline characters and returns the resulting lines.
Returns an empty slice for empty input.

```
len(lines(Body)) > 10
```

---

## Extraction

### `extract(s, pattern) string`

Returns the first match of the regular expression `pattern` in `s`.
Returns `""` if no match is found or the pattern is invalid.

```
extract(Body, "INV-\\d+")       // "INV-123" from "See INV-123"
extract(Log, "error: .+")       // first error line
```

### `email_domain(addr) string`

Returns the domain portion of an email address (everything after the
last `@`). Returns `""` if there is no `@`.

```
email_domain(Email)             // "example.com" from "alice@example.com"
```

---

## Utility

### `len(v) int`

Returns the length of a string, array, slice, or map.

```
len(Subject) > 0                // non-empty subject
len(Tags) >= 2                  // at least two tags
len(split(Body, ",")) == 3      // exactly three comma-separated values
```

---

## Operators

Expressions support the standard operators provided by
[expr-lang](https://expr-lang.org/docs/language-definition):

| Category    | Operators                          |
|-------------|------------------------------------|
| Logical     | `&&`, `\|\|`, `!`, `not`           |
| Comparison  | `==`, `!=`, `<`, `>`, `<=`, `>=`   |
| Arithmetic  | `+`, `-`, `*`, `/`, `%`, `**`      |
| String      | `+` (concatenation), `matches`     |
| Membership  | `in`, `not in`                     |
| Collections | `any`, `all`, `none`, `one`, `map`, `filter`, `count` |
| Nil         | `??` (nil coalescing)              |

Refer to the [expr-lang documentation](https://expr-lang.org/docs/language-definition)
for full details.
