// parser.go
package parser

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Default pattern constraints
const (
	DefaultMaxInputLength   = 512
	DefaultMaxTemplateLen   = 1024
	DefaultMaxRegexLen      = 4096
	DefaultParseTimeout     = 100 * time.Millisecond
	DefaultRegexTestLen     = 100
	DefaultRegexTestTimeout = 50 * time.Millisecond
)

// Predefined patterns with bounded quantifiers
var (
	PatVersion = `[0-9]{1,4}(?:\.[0-9]{1,4}){0,3}(?:[-+][0-9A-Za-z._+-]{1,50})?`
	PatArch    = `(?i:(?:x86_64(?:_v[234])?|x86-64(?:-v[234])?|amd64|arm64|aarch64|i386|i686|universal))`
	PatOS      = `(?i:linux|darwin|macos|mac|osx|windows|win32|win64|freebsd|openbsd|netbsd)`
	PatTriple  = `(?:unknown-linux-gnu|apple-darwin|pc-windows-(?:msvc|gnu)|linux-musl)`
	PatVariant = `[A-Za-z0-9+._-]{1,100}`
	PatWord    = `[A-Za-z0-9._-]{1,100}`
	PatIdent   = `[A-Za-z][A-Za-z0-9._-]{0,99}`
)

// FieldSpec defines the regexp fragment and optional normalizer for a field.
type FieldSpec struct {
	Pattern   string
	Normalize func(string) (string, error)
}

// ParseResult represents a successful parse.
type ParseResult struct {
	Fields   map[string]string
	Template string
	Raw      string
}

// ParseError represents a parsing error.
type ParseError struct {
	Input    string
	Position int
	Message  string
	Err      error
}

func (e *ParseError) Error() string {
	if e.Position >= 0 {
		return fmt.Sprintf("parse error at position %d in %q: %s", e.Position, e.Input, e.Message)
	}
	return fmt.Sprintf("parse error in %q: %s", e.Input, e.Message)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// MetricsCollector allows observability hooks.
type MetricsCollector interface {
	RecordParseAttempt(template string, success bool, duration time.Duration)
	RecordRegexTimeout(template string)
	RecordTemplateCompilation(template string, success bool)
}

// Parser compiles reverse templates to regexps with named capture groups.
type Parser struct {
	fieldSpecs     map[string]FieldSpec
	defaultPattern string
	compiled       []*compiledTemplate
	sealed         bool

	// Configuration
	maxInputLen      int
	maxTemplateLen   int
	maxRegexLen      int
	parseTimeout     time.Duration
	regexTestTimeout time.Duration

	// Observability
	logger  *log.Logger
	metrics MetricsCollector

	mu sync.RWMutex
}

type compiledTemplate struct {
	template string
	re       *regexp.Regexp
	fields   []string
}

// ParserBuilder provides a fluent interface for constructing a Parser.
type ParserBuilder struct {
	specs            map[string]FieldSpec
	defaultPattern   string
	maxInputLen      int
	maxTemplateLen   int
	maxRegexLen      int
	parseTimeout     time.Duration
	regexTestTimeout time.Duration
	logger           *log.Logger
	metrics          MetricsCollector
}

// NewParserBuilder creates a new parser builder with sensible defaults.
func NewParserBuilder() *ParserBuilder {
	return &ParserBuilder{
		specs:            make(map[string]FieldSpec),
		defaultPattern:   `[^/_\s\.-]{1,100}(?:[._-][^/_\s\.-]{1,100}){0,10}`,
		maxInputLen:      DefaultMaxInputLength,
		maxTemplateLen:   DefaultMaxTemplateLen,
		maxRegexLen:      DefaultMaxRegexLen,
		parseTimeout:     DefaultParseTimeout,
		regexTestTimeout: DefaultRegexTestTimeout,
	}
}

// WithField registers a field specification.
func (b *ParserBuilder) WithField(name string, spec FieldSpec) *ParserBuilder {
	b.specs[name] = spec
	return b
}

// WithDefaultPattern sets the default pattern for unspecified fields.
func (b *ParserBuilder) WithDefaultPattern(pattern string) *ParserBuilder {
	b.defaultPattern = pattern
	return b
}

// WithMaxInputLength sets the maximum allowed input length.
func (b *ParserBuilder) WithMaxInputLength(n int) *ParserBuilder {
	b.maxInputLen = n
	return b
}

// WithMaxTemplateLength sets the maximum allowed template length.
func (b *ParserBuilder) WithMaxTemplateLength(n int) *ParserBuilder {
	b.maxTemplateLen = n
	return b
}

// WithParseTimeout sets the timeout for parsing operations.
func (b *ParserBuilder) WithParseTimeout(d time.Duration) *ParserBuilder {
	b.parseTimeout = d
	return b
}

// WithLogger sets the logger.
func (b *ParserBuilder) WithLogger(logger *log.Logger) *ParserBuilder {
	b.logger = logger
	return b
}

// WithMetrics sets the metrics collector.
func (b *ParserBuilder) WithMetrics(metrics MetricsCollector) *ParserBuilder {
	b.metrics = metrics
	return b
}

// Build constructs the Parser.
func (b *ParserBuilder) Build() (*Parser, error) {
	if b.defaultPattern == "" {
		return nil, fmt.Errorf("default pattern cannot be empty")
	}

	// Test compile the default pattern
	if _, err := regexp.Compile(b.defaultPattern); err != nil {
		return nil, fmt.Errorf("invalid default pattern: %w", err)
	}

	return &Parser{
		fieldSpecs:       b.specs,
		defaultPattern:   b.defaultPattern,
		compiled:         make([]*compiledTemplate, 0),
		maxInputLen:      b.maxInputLen,
		maxTemplateLen:   b.maxTemplateLen,
		maxRegexLen:      b.maxRegexLen,
		parseTimeout:     b.parseTimeout,
		regexTestTimeout: b.regexTestTimeout,
		logger:           b.logger,
		metrics:          b.metrics,
	}, nil
}

// NewReleaseParser creates a parser with sensible defaults for release filenames.
func NewReleaseParser() (*Parser, error) {
	return NewParserBuilder().
		WithField("name", FieldSpec{Pattern: PatIdent}).
		WithField("version", FieldSpec{Pattern: PatVersion}).
		WithField("arch", FieldSpec{
			Pattern:   PatArch,
			Normalize: ArchNormalizer,
		}).
		WithField("os", FieldSpec{
			Pattern:   PatOS,
			Normalize: OSNormalizer,
		}).
		WithField("triple", FieldSpec{Pattern: PatTriple}).
		WithField("variant", FieldSpec{Pattern: PatVariant}).
		Build()
}

// ArchNormalizer normalizes architecture strings.
func ArchNormalizer(s string) (string, error) {
	switch strings.ToLower(s) {
	case "x86_64", "x86-64", "amd64":
		return "amd64", nil
	case "arm64", "aarch64":
		return "arm64", nil
	case "x86_64_v3", "x86-64-v3":
		return "amd64_v3", nil
	case "x86_64_v2", "x86-64-v2":
		return "amd64_v2", nil
	case "x86_64_v4", "x86-64-v4":
		return "amd64_v4", nil
	case "i386", "i686":
		return "386", nil
	case "universal":
		return "universal", nil
	default:
		return strings.ToLower(s), nil
	}
}

// OSNormalizer normalizes OS strings.
func OSNormalizer(s string) (string, error) {
	switch strings.ToLower(s) {
	case "darwin", "mac", "macos", "osx":
		return "darwin", nil
	case "linux":
		return "linux", nil
	case "windows", "win32", "win64":
		return "windows", nil
	case "freebsd":
		return "freebsd", nil
	case "openbsd":
		return "openbsd", nil
	case "netbsd":
		return "netbsd", nil
	default:
		return s, nil
	}
}

// AddTemplate compiles and registers a template.
//
// Reverse-template features:
//   - Literal text is escaped.
//   - {field} becomes (?P<field>pattern).
//   - {field?} makes the field optional as an optional named group.
//   - Optional segment: "[...]?" makes the bracketed part optional.
//   - To include literal '{' or '}', use '{{' or '}}'.
//
// Example: "{name}-{version}[-{variant}?]?_{os}-{arch}.tar.gz"
func (p *Parser) AddTemplate(tmpl string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.sealed {
		return fmt.Errorf("parser is sealed, cannot add templates")
	}

	if len(tmpl) > p.maxTemplateLen {
		return &ParseError{
			Input:   tmpl,
			Message: fmt.Sprintf("template too long: %d > %d", len(tmpl), p.maxTemplateLen),
		}
	}

	if strings.TrimSpace(tmpl) == "" {
		return &ParseError{Input: tmpl, Message: "empty template"}
	}

	src, fields, err := p.compileTemplate(tmpl)
	if err != nil {
		if p.metrics != nil {
			p.metrics.RecordTemplateCompilation(tmpl, false)
		}
		return err
	}

	if len(src) > p.maxRegexLen {
		if p.metrics != nil {
			p.metrics.RecordTemplateCompilation(tmpl, false)
		}
		return &ParseError{
			Input:   tmpl,
			Message: fmt.Sprintf("generated regex too complex: %d > %d", len(src), p.maxRegexLen),
		}
	}

	re, err := regexp.Compile("^" + src + "$")
	if err != nil {
		if p.metrics != nil {
			p.metrics.RecordTemplateCompilation(tmpl, false)
		}
		return &ParseError{
			Input:   tmpl,
			Message: "failed to compile regex",
			Err:     err,
		}
	}

	// Test for catastrophic backtracking
	if err := p.testRegexPerformance(re, tmpl); err != nil {
		if p.metrics != nil {
			p.metrics.RecordTemplateCompilation(tmpl, false)
			p.metrics.RecordRegexTimeout(tmpl)
		}
		return err
	}

	p.compiled = append(p.compiled, &compiledTemplate{
		template: tmpl,
		re:       re,
		fields:   fields,
	})

	if p.metrics != nil {
		p.metrics.RecordTemplateCompilation(tmpl, true)
	}

	if p.logger != nil {
		p.logger.Printf("Added template: %s (fields: %v)", tmpl, fields)
	}

	return nil
}

// Seal prevents further template additions and optimizes for parsing.
func (p *Parser) Seal() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sealed = true
}

// Parse attempts to parse the filename and returns a structured result.
func (p *Parser) Parse(filename string) (*ParseResult, error) {
	if err := p.validateInput(filename); err != nil {
		return nil, err
	}

	fields, tmpl, ok := p.tryParseWithTimeout(filename)
	if !ok {
		return nil, &ParseError{
			Input:   filename,
			Message: "no template matched",
		}
	}

	return &ParseResult{
		Fields:   fields,
		Template: tmpl,
		Raw:      filename,
	}, nil
}

// TryParse attempts to parse the filename, returning nil map if no match.
// This is the legacy API for backwards compatibility.
func (p *Parser) TryParse(filename string) (map[string]string, string, bool) {
	if err := p.validateInput(filename); err != nil {
		return nil, "", false
	}
	return p.tryParseWithTimeout(filename)
}

// validateInput checks input constraints.
func (p *Parser) validateInput(filename string) error {
	if len(filename) > p.maxInputLen {
		return &ParseError{
			Input:   filename,
			Message: fmt.Sprintf("input too long: %d > %d", len(filename), p.maxInputLen),
		}
	}

	if strings.TrimSpace(filename) == "" {
		return &ParseError{Input: filename, Message: "empty input"}
	}

	// Check for null bytes
	if strings.ContainsRune(filename, 0) {
		return &ParseError{Input: filename, Message: "input contains null byte"}
	}

	return nil
}

// tryParseWithTimeout performs parsing with timeout protection.
func (p *Parser) tryParseWithTimeout(filename string) (map[string]string, string, bool) {
	type result struct {
		fields map[string]string
		tmpl   string
		ok     bool
	}

	ch := make(chan result, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				if p.logger != nil {
					p.logger.Printf("Panic during parse: %v", r)
				}
				ch <- result{nil, "", false}
			}
		}()

		f, t, ok := p.doTryParse(filename)
		ch <- result{f, t, ok}
	}()

	select {
	case r := <-ch:
		return r.fields, r.tmpl, r.ok
	case <-time.After(p.parseTimeout):
		if p.logger != nil {
			p.logger.Printf("Parse timeout for: %s", filename)
		}
		return nil, "", false
	}
}

// doTryParse performs the actual parsing logic.
func (p *Parser) doTryParse(filename string) (map[string]string, string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	start := time.Now()

	for _, ct := range p.compiled {
		m := ct.re.FindStringSubmatch(filename)
		if m == nil {
			continue
		}
		out := make(map[string]string)
		names := ct.re.SubexpNames()
		for i := 1; i < len(names) && i < len(m); i++ {
			name := names[i]
			if name == "" {
				continue
			}
			val := m[i]
			if val == "" {
				continue
			}
			if spec, ok := p.fieldSpecs[name]; ok && spec.Normalize != nil {
				if norm, err := spec.Normalize(val); err == nil {
					val = norm
				}
			}
			out[name] = val
		}
		if p.metrics != nil {
			p.metrics.RecordParseAttempt(ct.template, true, time.Since(start))
		}
		return out, ct.template, true
	}

	if p.metrics != nil {
		p.metrics.RecordParseAttempt("", false, time.Since(start))
	}
	return nil, "", false
}

// testRegexPerformance tests the regex against pathological input.
func (p *Parser) testRegexPerformance(re *regexp.Regexp, tmpl string) error {
	// Test with repeated patterns that might trigger backtracking
	testInputs := []string{
		strings.Repeat("a", DefaultRegexTestLen),
		strings.Repeat("a-", DefaultRegexTestLen/2),
		strings.Repeat("1.", DefaultRegexTestLen/2),
	}

	done := make(chan bool, 1)

	go func() {
		for _, input := range testInputs {
			re.MatchString(input)
		}
		done <- true
	}()

	select {
	case <-done:
		return nil
	case <-time.After(p.regexTestTimeout):
		return &ParseError{
			Input:   tmpl,
			Message: "regex performance test timed out - potential catastrophic backtracking",
		}
	}
}

// compileTemplate converts the reverse template into a regexp source.
func (p *Parser) compileTemplate(tmpl string) (string, []string, error) {
	var b strings.Builder
	fields := []string{}
	i := 0

	for i < len(tmpl) {
		ch := tmpl[i]
		switch ch {
		case '{':
			// Escaped '{{'
			if i+1 < len(tmpl) && tmpl[i+1] == '{' {
				b.WriteString(regexp.QuoteMeta("{"))
				i += 2
				continue
			}

			// Field: {name} or {name?}
			closeIdx := strings.IndexByte(tmpl[i+1:], '}')
			nextOpenIdx := strings.IndexByte(tmpl[i+1:], '{')
			if closeIdx == -1 || nextOpenIdx > -1 && closeIdx > nextOpenIdx {
				return "", nil, &ParseError{
					Input:    tmpl,
					Position: i,
					Message:  fmt.Sprintf("unclosed '{' at %d", i),
				}
			}

			raw := tmpl[i+1 : i+1+closeIdx]
			name := strings.TrimSuffix(raw, "?")
			optional := raw != name

			if name == "" {
				return "", nil, &ParseError{
					Input:    tmpl,
					Position: i,
					Message:  "empty field name",
				}
			}

			// Validate field name
			if !isValidFieldName(name) {
				return "", nil, &ParseError{
					Input:    tmpl,
					Position: i,
					Message:  fmt.Sprintf("invalid field name: %s", name),
				}
			}

			pat := p.defaultPattern
			if spec, ok := p.fieldSpecs[name]; ok && strings.TrimSpace(spec.Pattern) != "" {
				pat = spec.Pattern
			}

			if optional {
				fmt.Fprintf(&b, "(?P<%s>%s)?", name, pat)
			} else {
				fmt.Fprintf(&b, "(?P<%s>%s)", name, pat)
			}

			fields = append(fields, name)
			i = i + 1 + closeIdx + 1

		case '}':
			// Escaped '}}'
			if i+1 < len(tmpl) && tmpl[i+1] == '}' {
				b.WriteString(regexp.QuoteMeta("}"))
				i += 2
				continue
			}

			return "", nil, &ParseError{
				Input:    tmpl,
				Position: i,
				Message:  "unmatched '}'",
			}

		case '[':
			// Optional segment: [...]?
			closeIdx := findMatchingBracket(tmpl, i+1)
			if closeIdx == -1 {
				return "", nil, &ParseError{
					Input:    tmpl,
					Position: i,
					Message:  "unclosed '['",
				}
			}

			seg := tmpl[i+1 : closeIdx]

			// Check if followed by '?'
			isOptional := closeIdx+1 < len(tmpl) && tmpl[closeIdx+1] == '?'

			// Compile segment content
			segSrc, segFields, err := p.compileSegment(seg)
			if err != nil {
				return "", nil, fmt.Errorf("in segment at position %d: %w", i, err)
			}

			fields = append(fields, segFields...)

			if isOptional {
				fmt.Fprintf(&b, "(?:%s)?", segSrc)
				i = closeIdx + 2
			} else {
				fmt.Fprintf(&b, "(?:%s)", segSrc)
				i = closeIdx + 1
			}

		case ']':
			return "", nil, &ParseError{
				Input:    tmpl,
				Position: i,
				Message:  "unmatched ']'",
			}

		default:
			b.WriteString(regexp.QuoteMeta(string(ch)))
			i++
		}
	}

	return b.String(), uniqueStrings(fields), nil
}

// compileSegment compiles a segment (content inside [...]).
func (p *Parser) compileSegment(seg string) (string, []string, error) {
	var b strings.Builder
	fields := []string{}
	i := 0

	for i < len(seg) {
		ch := seg[i]
		switch ch {
		case '{':
			if i+1 < len(seg) && seg[i+1] == '{' {
				b.WriteString(regexp.QuoteMeta("{"))
				i += 2
				continue
			}

			closeIdx := strings.IndexByte(seg[i+1:], '}')
			if closeIdx == -1 {
				return "", nil, fmt.Errorf("unclosed '{' in segment")
			}

			raw := seg[i+1 : i+1+closeIdx]
			name := strings.TrimSuffix(raw, "?")
			optional := raw != name

			if name == "" {
				return "", nil, fmt.Errorf("empty field name in segment")
			}

			if !isValidFieldName(name) {
				return "", nil, fmt.Errorf("invalid field name in segment: %s", name)
			}

			pat := p.defaultPattern
			if spec, ok := p.fieldSpecs[name]; ok && strings.TrimSpace(spec.Pattern) != "" {
				pat = spec.Pattern
			}

			if optional {
				fmt.Fprintf(&b, "(?P<%s>%s)?", name, pat)
			} else {
				fmt.Fprintf(&b, "(?P<%s>%s)", name, pat)
			}

			fields = append(fields, name)
			i = i + 1 + closeIdx + 1

		case '}':
			if i+1 < len(seg) && seg[i+1] == '}' {
				b.WriteString(regexp.QuoteMeta("}"))
				i += 2
				continue
			}
			return "", nil, fmt.Errorf("unmatched '}' in segment")

		case '[':
			return "", nil, fmt.Errorf("nested optional segments not supported")

		default:
			b.WriteString(regexp.QuoteMeta(string(ch)))
			i++
		}
	}

	return b.String(), fields, nil
}

// isValidFieldName checks if a field name is valid.
func isValidFieldName(name string) bool {
	if name == "" {
		return false
	}

	// Must start with letter or underscore
	first := rune(name[0])
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}

	// Rest can be alphanumeric or underscore
	for _, r := range name[1:] {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}

	return true
}

// findMatchingBracket finds the index of the matching ']' for a '['.
func findMatchingBracket(tmpl string, start int) int {
	depth := 1
	for i := start; i < len(tmpl); i++ {
		switch tmpl[i] {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// uniqueStrings returns unique strings preserving order.
func uniqueStrings(ss []string) []string {
	seen := make(map[string]struct{}, len(ss))
	out := make([]string, 0, len(ss))
	for _, s := range ss {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}
