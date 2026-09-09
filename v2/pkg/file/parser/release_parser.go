package parser

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// Errors
var (
	ErrNoMatch         = errors.New("no match")
	ErrInvalidPattern  = errors.New("invalid pattern")
	ErrUnknownPlatform = errors.New("unknown platform")
	ErrUnknownArch     = errors.New("unknown architecture")
	ErrUnknownFormat   = errors.New("unknown format")
)

// Artifact represents structured information extracted from a filename.
type Artifact struct {
	Platform string // normalized: darwin|linux|windows
	Arch     string // normalized: arm64|amd64|386|armv7|...
	Format   string // e.g. tar.gz, zip, tar.zst
	Version  string // e.g. 10.3.1, 1.3.0-rc.2, v3.0.0-beta.11
	Ext      string // e.g. gz, zip (single extension)
	Variant  string // enum-like value constrained by pattern
}

// FilenameParser is a compiled filename pattern parser.
// Safe for concurrent use after compilation.
type FilenameParser struct {
	pattern     string
	re          *regexp.Regexp
	hasVersion  bool
	hasFormat   bool
	hasArch     bool
	hasPlatform bool
	mu          sync.RWMutex // For future extensibility
}

// ParseResult contains successful parse and any warnings.
type ParseResult struct {
	Artifact Artifact
	Warnings []string
}

// CompileOptions configures pattern compilation behavior.
type CompileOptions struct {
	// AllowPartialMatch allows matching even if some fields are missing
	AllowPartialMatch bool
	// CaseSensitive makes platform/arch matching case-sensitive
	CaseSensitive bool
}

// DefaultOptions returns sensible default compilation options.
func DefaultOptions() CompileOptions {
	return CompileOptions{
		AllowPartialMatch: false,
		CaseSensitive:     false,
	}
}

// CompileFilenamePattern compiles a filename pattern into a parser.
//
// Supported tags:
//   - {:platform} - OS platform (darwin, linux, windows)
//   - {:platform?} - Optional platform
//   - {:arch} - CPU architecture (arm64, amd64, 386, etc.)
//   - {:arch?} - Optional architecture
//   - {:format} - Archive format (tar.gz, zip, etc.)
//   - {:format?} - Optional format
//   - {:version} - Semantic version
//   - {:version?} - Optional version
//   - {:ext} - Single file extension
//   - {:variant["a","b"]} - Enum-like allowed values
//
// The compiled regex is anchored (^...$).
func CompileFilenamePattern(pattern string) (*FilenameParser, error) {
	return CompileFilenamePatternWithOptions(pattern, DefaultOptions())
}

// CompileFilenamePatternWithOptions compiles with custom options.
func CompileFilenamePatternWithOptions(
	pattern string,
	opts CompileOptions,
) (*FilenameParser, error) {
	if strings.TrimSpace(pattern) == "" {
		return nil, fmt.Errorf("%w: empty pattern", ErrInvalidPattern)
	}

	parser := &FilenameParser{
		pattern: pattern,
	}

	regexSrc, err := compileToRegex(pattern, opts, parser)
	if err != nil {
		return nil, err
	}

	re, err := regexp.Compile(regexSrc)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: regex compile failed: %v",
			ErrInvalidPattern,
			err,
		)
	}

	parser.re = re
	return parser, nil
}

// ParseFilename parses a filename using the compiled pattern.
func (p *FilenameParser) ParseFilename(
	filename string,
) (ParseResult, error) {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return ParseResult{}, ErrNoMatch
	}

	// Prevent multiline attacks
	if strings.ContainsAny(filename, "\r\n") {
		return ParseResult{}, fmt.Errorf(
			"%w: filename contains newlines",
			ErrNoMatch,
		)
	}

	m := p.re.FindStringSubmatch(filename)
	if m == nil {
		return ParseResult{}, ErrNoMatch
	}

	// Build capture group map
	values := make(map[string]string, p.re.NumSubexp())
	for i, name := range p.re.SubexpNames() {
		if i == 0 || name == "" {
			continue
		}
		values[name] = m[i]
	}

	var result ParseResult
	var warnings []string

	// Parse platform
	if v, ok := values["platform"]; ok && v != "" {
		normalized := normalizePlatform(v)
		if normalized == "" {
			return ParseResult{}, fmt.Errorf(
				"%w: %q",
				ErrUnknownPlatform,
				v,
			)
		}
		result.Artifact.Platform = normalized
	}

	// Parse architecture
	if v, ok := values["arch"]; ok && v != "" {
		normalized := normalizeArch(v)
		if normalized == "" {
			return ParseResult{}, fmt.Errorf(
				"%w: %q",
				ErrUnknownArch,
				v,
			)
		}
		result.Artifact.Arch = normalized
	}

	// Parse format
	if v, ok := values["format"]; ok && v != "" {
		normalized := normalizeFormat(v)
		if normalized == "" {
			return ParseResult{}, fmt.Errorf(
				"%w: %q",
				ErrUnknownFormat,
				v,
			)
		}
		result.Artifact.Format = normalized
	}

	// Parse extension (no normalization needed)
	if v, ok := values["ext"]; ok {
		result.Artifact.Ext = strings.ToLower(v)
	}

	// Parse version (already validated by regex)
	if v, ok := values["version"]; ok {
		result.Artifact.Version = v
	}

	// Parse variant (already validated by regex)
	if v, ok := values["variant"]; ok {
		result.Artifact.Variant = v
	}

	result.Warnings = warnings
	return result, nil
}

// ParseFilename is a convenience wrapper for one-off parsing.
func ParseFilename(pattern, filename string) (ParseResult, error) {
	p, err := CompileFilenamePattern(pattern)
	if err != nil {
		return ParseResult{}, err
	}
	return p.ParseFilename(filename)
}

// ParseFilenamesFromLines parses filenames from line-separated input.
// Lines may contain comma-separated values; only the first part is used.
type LineParseResult struct {
	Artifacts []Artifact
	Errors    []LineError
}

type LineError struct {
	Line     int
	Filename string
	Err      error
}

func (p *FilenameParser) ParseFilenamesFromLines(
	input string,
) LineParseResult {
	var result LineParseResult

	lines := strings.Split(input, "\n")
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Extract filename before comma
		filename := line
		if idx := strings.IndexByte(line, ','); idx >= 0 {
			filename = strings.TrimSpace(line[:idx])
		}
		if filename == "" {
			continue
		}

		parseResult, err := p.ParseFilename(filename)
		if err != nil {
			result.Errors = append(result.Errors, LineError{
				Line:     lineNum + 1,
				Filename: filename,
				Err:      err,
			})
			continue
		}

		result.Artifacts = append(result.Artifacts, parseResult.Artifact)
	}

	return result
}

// ParseFilenamesFromLines is a convenience wrapper.
func ParseFilenamesFromLines(
	pattern string,
	input string,
) (LineParseResult, error) {
	p, err := CompileFilenamePattern(pattern)
	if err != nil {
		return LineParseResult{}, err
	}
	return p.ParseFilenamesFromLines(input), nil
}

// Pattern compiler

func compileToRegex(
	pattern string,
	opts CompileOptions,
	parser *FilenameParser,
) (string, error) {
	var b strings.Builder
	b.Grow(len(pattern) * 2) // Pre-allocate
	b.WriteString("^")

	for i := 0; i < len(pattern); {
		// Find next control token: "{:" (tag) or "{?" (optional group)
		nextTag := strings.Index(pattern[i:], "{:")
		if nextTag >= 0 {
			nextTag += i
		}
		nextGroup := strings.Index(pattern[i:], "{?")
		if nextGroup >= 0 {
			nextGroup += i
		}

		j := -1
		switch {
		case nextTag >= 0 && nextGroup >= 0:
			if nextTag < nextGroup {
				j = nextTag
			} else {
				j = nextGroup
			}
		case nextTag >= 0:
			j = nextTag
		case nextGroup >= 0:
			j = nextGroup
		default:
			// No more special tokens
			b.WriteString(regexp.QuoteMeta(pattern[i:]))
			i = len(pattern)
			continue
		}

		// Emit literal segment
		b.WriteString(regexp.QuoteMeta(pattern[i:j]))

		if strings.HasPrefix(pattern[j:], "{?") {
			// Optional group: "{? ... }"
			end, err := findGroupEnd(pattern, j)
			if err != nil {
				return "", err
			}

			inner := pattern[j+2 : end] // exclude "{?" and trailing "}"
			innerRegex, err := compileToRegexInner(inner, opts, parser)
			if err != nil {
				return "", err
			}

			b.WriteString("(?:")
			b.WriteString(innerRegex)
			b.WriteString(")?")
			i = end + 1
			continue
		}

		// Regular tag: "{:...}"
		end := strings.IndexByte(pattern[j:], '}')
		if end < 0 {
			return "", fmt.Errorf(
				"%w: unclosed tag at position %d",
				ErrInvalidPattern,
				j,
			)
		}
		end += j

		tag := pattern[j : end+1]
		frag, err := tagToRegex(tag, opts, parser)
		if err != nil {
			return "", err
		}
		b.WriteString(frag)
		i = end + 1
	}

	b.WriteString("$")
	return b.String(), nil
}

func tagToRegex(
	tag string,
	opts CompileOptions,
	parser *FilenameParser,
) (string, error) {
	// Check for optional marker
	optional := strings.HasSuffix(tag, "?}")
	baseTag := tag
	if optional {
		baseTag = strings.TrimSuffix(tag, "?}") + "}"
	}

	var regex string
	var err error

	switch baseTag {
	case "{:platform}":
		parser.hasPlatform = true
		regex = fmt.Sprintf(
			"(?P<platform>%s)",
			alternation(platformTokens(), opts.CaseSensitive),
		)

	case "{:arch}":
		parser.hasArch = true
		regex = fmt.Sprintf(
			"(?P<arch>%s)",
			alternation(archTokens(), opts.CaseSensitive),
		)

	case "{:format}":
		parser.hasFormat = true
		// Format is case-insensitive by nature of file extensions
		regex = fmt.Sprintf(
			"(?P<format>%s)",
			alternation(formatTokens(), false),
		)

	case "{:ext}":
		regex = `(?P<ext>[A-Za-z0-9]+)`

	case "{:version}":
		parser.hasVersion = true
		// Improved version regex supporting CalVer and SemVer
		regex = `(?P<version>v?[0-9]+(?:\.[0-9]+){0,3}` +
			`(?:-[0-9A-Za-z]+(?:\.[0-9A-Za-z]+)*)?` +
			`(?:\+[0-9A-Za-z]+(?:\.[0-9A-Za-z]+)*)?)`

	default:
		// Check for variant enum
		if strings.HasPrefix(baseTag, "{:variant[") &&
			strings.HasSuffix(baseTag, "]}") {
			allowed, err := parseVariantAllowed(baseTag)
			if err != nil {
				return "", err
			}
			if len(allowed) == 0 {
				return "", fmt.Errorf(
					"%w: variant must have at least one value",
					ErrInvalidPattern,
				)
			}
			regex = fmt.Sprintf(
				"(?P<variant>%s)",
				alternation(allowed, true), // Variants are case-sensitive
			)
		} else {
			return "", fmt.Errorf(
				"%w: unsupported tag %q",
				ErrInvalidPattern,
				tag,
			)
		}
	}

	if optional {
		regex = fmt.Sprintf("(?:%s)?", regex)
	}

	return regex, err
}

func parseVariantAllowed(tag string) ([]string, error) {
	// Extract content between [ and ]
	inner := strings.TrimPrefix(tag, "{:variant[")
	inner = strings.TrimSuffix(inner, "]}")
	inner = strings.TrimSpace(inner)

	if inner == "" {
		return nil, fmt.Errorf(
			"%w: variant list is empty",
			ErrInvalidPattern,
		)
	}

	var result []string
	for len(inner) > 0 {
		inner = strings.TrimSpace(inner)
		if len(inner) == 0 {
			break
		}

		if inner[0] != '`' {
			return nil, fmt.Errorf(
				"%w: variant values must be backtick-quoted",
				ErrInvalidPattern,
			)
		}

		val, rest, err := consumeQuoted(inner)
		if err != nil {
			return nil, err
		}

		result = append(result, val)

		rest = strings.TrimSpace(rest)
		if rest == "" {
			break
		}
		if rest[0] != ',' {
			return nil, fmt.Errorf(
				"%w: expected comma between variant values",
				ErrInvalidPattern,
			)
		}
		inner = rest[1:]
	}

	return result, nil
}

func consumeQuoted(s string) (string, string, error) {
	if s == "" || s[0] != '`' {
		return "", "", fmt.Errorf(
			"%w: expected backtick-quoted string",
			ErrInvalidPattern,
		)
	}

	var b strings.Builder
	escaped := false

	for i := 1; i < len(s); i++ {
		c := s[i]
		if escaped {
			switch c {
			case '`', '\\':
				b.WriteByte(c)
			default:
				// Unknown escape - keep the backslash
				b.WriteByte('\\')
				b.WriteByte(c)
			}
			escaped = false
			continue
		}

		if c == '\\' {
			escaped = true
			continue
		}

		if c == '`' {
			return b.String(), s[i+1:], nil
		}

		b.WriteByte(c)
	}

	return "", "", fmt.Errorf(
		"%w: unterminated backtick-quoted string",
		ErrInvalidPattern,
	)
}

// Dictionaries & normalization

func platformTokens() []string {
	return []string{
		"darwin", "macos", "mac", "osx", "apple-darwin",
		"linux",
		"windows", "win", "win32", "win64",
	}
}

func normalizePlatform(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case "darwin", "macos", "mac", "osx", "apple-darwin":
		return "darwin"
	case "linux":
		return "linux"
	case "windows", "win", "win32", "win64":
		return "windows"
	default:
		return ""
	}
}

func archTokens() []string {
	return []string{
		"arm64", "aarch64",
		"amd64", "x86_64", "x64",
		"386", "i386", "x86",
		"armv7", "armv6",
	}
}

func normalizeArch(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case "arm64", "aarch64":
		return "arm64"
	case "amd64", "x86_64", "x64":
		return "amd64"
	case "386", "i386", "x86":
		return "386"
	case "armv7":
		return "armv7"
	case "armv6":
		return "armv6"
	default:
		return ""
	}
}

func formatTokens() []string {
	return []string{
		"tar.gz", "tar.zst", "tar.xz", "tar.bz2",
		"tgz", "zip", "gz", "zst", "xz", "bz2",
		"exe",
	}
}

func normalizeFormat(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	// Only accept known formats
	for _, token := range formatTokens() {
		if s == token {
			return s
		}
	}
	return ""
}

func alternation(tokens []string, caseSensitive bool) string {
	// Sort by length descending for greedy matching
	sorted := make([]string, len(tokens))
	copy(sorted, tokens)
	sort.Slice(sorted, func(i, j int) bool {
		if len(sorted[i]) == len(sorted[j]) {
			return sorted[i] < sorted[j]
		}
		return len(sorted[i]) > len(sorted[j])
	})

	// Escape tokens
	escaped := make([]string, 0, len(sorted))
	for _, token := range sorted {
		escaped = append(escaped, regexp.QuoteMeta(token))
	}

	joined := strings.Join(escaped, "|")

	if caseSensitive {
		return joined
	}
	return "(?i:" + joined + ")"
}

// compileToRegexInner compiles a pattern fragment without ^ and $ anchors.
func compileToRegexInner(
	pattern string,
	opts CompileOptions,
	parser *FilenameParser,
) (string, error) {
	// Reuse compileToRegex, then strip anchors. This keeps behavior consistent.
	src, err := compileToRegex(pattern, opts, parser)
	if err != nil {
		return "", err
	}
	// compileToRegex always returns ^...$, so strip them.
	if strings.HasPrefix(src, "^") {
		src = strings.TrimPrefix(src, "^")
	}
	if strings.HasSuffix(src, "$") {
		src = strings.TrimSuffix(src, "$")
	}
	return src, nil
}

func findGroupEnd(pattern string, start int) (int, error) {
	// start points at '{' and pattern[start:start+2] == "{?"
	// Groups cannot be nested (keeps parsing simple and predictable).
	for i := start + 2; i < len(pattern); i++ {
		// Reject nested optional groups.
		if pattern[i] == '{' && strings.HasPrefix(pattern[i:], "{?") {
			return 0, fmt.Errorf(
				"%w: nested optional groups are not supported at position %d",
				ErrInvalidPattern,
				i,
			)
		}

		// Skip over normal tags "{:...}" so their '}' doesn't close the group.
		if pattern[i] == '{' && strings.HasPrefix(pattern[i:], "{:") {
			end := strings.IndexByte(pattern[i:], '}')
			if end < 0 {
				return 0, fmt.Errorf(
					"%w: unclosed tag at position %d",
					ErrInvalidPattern,
					i,
				)
			}
			i += end // jump to the tag's closing '}'
			continue
		}

		// First '}' that isn't part of a "{:...}" tag closes the optional group.
		if pattern[i] == '}' {
			return i, nil
		}
	}

	return 0, fmt.Errorf(
		"%w: unclosed optional group at position %d",
		ErrInvalidPattern,
		start,
	)
}
