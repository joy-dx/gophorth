// parser_test.go
package parser

import (
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

// TestGoldenTable uses the golden table pattern for comprehensive test coverage.
func TestGoldenTable(t *testing.T) {
	tests := []struct {
		name        string
		templates   []string
		input       string
		wantMatch   bool
		wantFields  map[string]string
		wantTmpl    string
		description string
	}{
		{
			name:      "UV x86_64 linux",
			templates: []string{"{name}-{arch}-{triple}.tar.gz"},
			input:     "UV-x86_64-unknown-linux-gnu.tar.gz",
			wantMatch: true,
			wantFields: map[string]string{
				"name":   "UV",
				"arch":   "amd64",
				"triple": "unknown-linux-gnu",
			},
			wantTmpl:    "{name}-{arch}-{triple}.tar.gz",
			description: "UV release with triple target",
		},
		{
			name:      "UV aarch64 darwin",
			templates: []string{"{name}-{arch}-{triple}.tar.gz"},
			input:     "UV-aarch64-apple-darwin.tar.gz",
			wantMatch: true,
			wantFields: map[string]string{
				"name":   "UV",
				"arch":   "arm64",
				"triple": "apple-darwin",
			},
			wantTmpl:    "{name}-{arch}-{triple}.tar.gz",
			description: "UV release for macOS ARM",
		},
		{
			name:      "CPython with variant",
			templates: []string{"{name}-{version}-{arch}-{triple}-{variant}.tar.zst"},
			input:     "cpython-3.10.16+20250212-aarch64-apple-darwin-pgo+lto-full.tar.zst",
			wantMatch: true,
			wantFields: map[string]string{
				"name":    "cpython",
				"version": "3.10.16+20250212",
				"arch":    "arm64",
				"triple":  "apple-darwin",
				"variant": "pgo+lto-full",
			},
			wantTmpl:    "{name}-{version}-{arch}-{triple}-{variant}.tar.zst",
			description: "CPython with version, variant, and triple",
		},
		{
			name:      "CPython x86_64_v3",
			templates: []string{"{name}-{version}-{arch}-{triple}-{variant}.tar.zst"},
			input:     "cpython-3.10.16+20250212-x86_64_v3-unknown-linux-gnu-pgo+lto-full.tar.zst",
			wantMatch: true,
			wantFields: map[string]string{
				"name":    "cpython",
				"version": "3.10.16+20250212",
				"arch":    "amd64_v3",
				"triple":  "unknown-linux-gnu",
				"variant": "pgo+lto-full",
			},
			wantTmpl:    "{name}-{version}-{arch}-{triple}-{variant}.tar.zst",
			description: "CPython with x86_64_v3 normalization",
		},
		{
			name:      "Hugo underscore separator",
			templates: []string{"{name}_{version}_{os}-{arch}.tar.gz"},
			input:     "hugo_0.151.1_darwin-universal.tar.gz",
			wantMatch: true,
			wantFields: map[string]string{
				"name":    "hugo",
				"version": "0.151.1",
				"os":      "darwin",
				"arch":    "universal",
			},
			wantTmpl:    "{name}_{version}_{os}-{arch}.tar.gz",
			description: "Hugo with underscore separators",
		},
		{
			name:      "Hugo linux amd64",
			templates: []string{"{name}_{version}_{os}-{arch}.tar.gz"},
			input:     "hugo_0.151.1_linux-amd64.tar.gz",
			wantMatch: true,
			wantFields: map[string]string{
				"name":    "hugo",
				"version": "0.151.1",
				"os":      "linux",
				"arch":    "amd64",
			},
			wantTmpl:    "{name}_{version}_{os}-{arch}.tar.gz",
			description: "Hugo for Linux",
		},
		{
			name:      "Hugo arm64",
			templates: []string{"{name}_{version}_{os}-{arch}.tar.gz"},
			input:     "hugo_0.151.1_linux-arm64.tar.gz",
			wantMatch: true,
			wantFields: map[string]string{
				"name":    "hugo",
				"version": "0.151.1",
				"os":      "linux",
				"arch":    "arm64",
			},
			wantTmpl:    "{name}_{version}_{os}-{arch}.tar.gz",
			description: "Hugo for Linux ARM64",
		},
		{
			name:      "GoReleaser capital case",
			templates: []string{"{name}_{os}_{arch}.tar.gz"},
			input:     "goreleaser_Darwin_arm64.tar.gz",
			wantMatch: true,
			wantFields: map[string]string{
				"name": "goreleaser",
				"os":   "darwin",
				"arch": "arm64",
			},
			wantTmpl:    "{name}_{os}_{arch}.tar.gz",
			description: "GoReleaser with capital case OS",
		},
		{
			name:      "GoReleaser Linux x86_64",
			templates: []string{"{name}_{os}_{arch}.tar.gz"},
			input:     "goreleaser_Linux_x86_64.tar.gz",
			wantMatch: true,
			wantFields: map[string]string{
				"name": "goreleaser",
				"os":   "linux",
				"arch": "amd64",
			},
			wantTmpl:    "{name}_{os}_{arch}.tar.gz",
			description: "GoReleaser Linux with x86_64 normalization",
		},
		{
			name:      "golangci-lint hyphen separator",
			templates: []string{"{name}-{version}-{os}-{arch}.tar.gz"},
			input:     "golangci-lint-1.64.6-darwin-arm64.tar.gz",
			wantMatch: true,
			wantFields: map[string]string{
				"name":    "golangci-lint",
				"version": "1.64.6",
				"os":      "darwin",
				"arch":    "arm64",
			},
			wantTmpl:    "{name}-{version}-{os}-{arch}.tar.gz",
			description: "golangci-lint with hyphens",
		},
		{
			name:      "golangci-lint linux",
			templates: []string{"{name}-{version}-{os}-{arch}.tar.gz"},
			input:     "golangci-lint-1.64.6-linux-amd64.tar.gz",
			wantMatch: true,
			wantFields: map[string]string{
				"name":    "golangci-lint",
				"version": "1.64.6",
				"os":      "linux",
				"arch":    "amd64",
			},
			wantTmpl:    "{name}-{version}-{os}-{arch}.tar.gz",
			description: "golangci-lint for Linux",
		},
		{
			name:      "go-size-analyzer",
			templates: []string{"{name}_{version}_{os}_{arch}.tar.gz"},
			input:     "go-size-analyzer_1.10.0_darwin_arm64.tar.gz",
			wantMatch: true,
			wantFields: map[string]string{
				"name":    "go-size-analyzer",
				"version": "1.10.0",
				"os":      "darwin",
				"arch":    "arm64",
			},
			wantTmpl:    "{name}_{version}_{os}_{arch}.tar.gz",
			description: "go-size-analyzer with underscores",
		},
		{
			name:      "air with extension",
			templates: []string{"{name}_{version}_{os}_{arch}.tar.gz"},
			input:     "air_1.52.3_darwin_amd64.tar.gz",
			wantMatch: true,
			wantFields: map[string]string{
				"name":    "air",
				"version": "1.52.3",
				"os":      "darwin",
				"arch":    "amd64",
			},
			wantTmpl:    "{name}_{version}_{os}_{arch}.tar.gz",
			description: "air with tar.gz",
		},
		{
			name:      "air no extension",
			templates: []string{"{name}_{version}_{os}_{arch}"},
			input:     "air_1.52.3_linux_amd64",
			wantMatch: true,
			wantFields: map[string]string{
				"name":    "air",
				"version": "1.52.3",
				"os":      "linux",
				"arch":    "amd64",
			},
			wantTmpl:    "{name}_{version}_{os}_{arch}",
			description: "air without extension",
		},
		{
			name:      "frankenphp simple",
			templates: []string{"{name}-{os}-{arch}"},
			input:     "frankenphp-linux-x86_64",
			wantMatch: true,
			wantFields: map[string]string{
				"name": "frankenphp",
				"os":   "linux",
				"arch": "amd64",
			},
			wantTmpl:    "{name}-{os}-{arch}",
			description: "frankenphp binary without extension",
		},
		{
			name:      "frankenphp mac",
			templates: []string{"{name}-{os}-{arch}"},
			input:     "frankenphp-mac-arm64",
			wantMatch: true,
			wantFields: map[string]string{
				"name": "frankenphp",
				"os":   "darwin",
				"arch": "arm64",
			},
			wantTmpl:    "{name}-{os}-{arch}",
			description: "frankenphp with 'mac' OS normalization",
		},
		{
			name:      "docker-credential with version prefix",
			templates: []string{"{name}-v{version}.{os}-{arch}"},
			input:     "docker-credential-osxkeychain-v0.9.3.darwin-arm64",
			wantMatch: true,
			wantFields: map[string]string{
				"name":    "docker-credential-osxkeychain",
				"version": "0.9.3",
				"os":      "darwin",
				"arch":    "arm64",
			},
			wantTmpl:    "{name}-v{version}.{os}-{arch}",
			description: "docker-credential with 'v' version prefix",
		},
		{
			name:      "docker-credential linux",
			templates: []string{"{name}-v{version}.{os}-{arch}"},
			input:     "docker-credential-secretservice-v0.9.3.linux-amd64",
			wantMatch: true,
			wantFields: map[string]string{
				"name":    "docker-credential-secretservice",
				"version": "0.9.3",
				"os":      "linux",
				"arch":    "amd64",
			},
			wantTmpl:    "{name}-v{version}.{os}-{arch}",
			description: "docker-credential for Linux",
		},
		{
			name:      "optional field present",
			templates: []string{"{name}-{version?}-{arch}"},
			input:     "myapp-1.2.3-amd64",
			wantMatch: true,
			wantFields: map[string]string{
				"name":    "myapp",
				"version": "1.2.3",
				"arch":    "amd64",
			},
			wantTmpl:    "{name}-{version?}-{arch}",
			description: "optional version field is present",
		},
		{
			name:      "optional field absent",
			templates: []string{"{name}[-{version?}]?-{arch}"},
			input:     "myapp-amd64",
			wantMatch: true,
			wantFields: map[string]string{
				"name": "myapp",
				"arch": "amd64",
			},
			wantTmpl:    "{name}[-{version?}]?-{arch}",
			description: "optional version field is absent",
		},
		{
			name:      "optional segment present",
			templates: []string{"{name}-{variant}_{arch}"},
			input:     "myapp-debug_amd64",
			wantMatch: true,
			wantFields: map[string]string{
				"name":    "myapp",
				"variant": "debug",
				"arch":    "amd64",
			},
			wantTmpl:    "{name}-{variant}_{arch}",
			description: "optional segment with variant present",
		},
		{
			name:      "optional segment absent",
			templates: []string{"{name}[-{variant}]?_{arch}"},
			input:     "myapp_amd64",
			wantMatch: true,
			wantFields: map[string]string{
				"name": "myapp",
				"arch": "amd64",
			},
			wantTmpl:    "{name}[-{variant}]?_{arch}",
			description: "optional segment absent",
		},
		{
			name:      "multiple templates first match",
			templates: []string{"{name}_{version}_{os}_{arch}", "{name}-{os}-{arch}"},
			input:     "app_1.0.0_linux_amd64",
			wantMatch: true,
			wantFields: map[string]string{
				"name":    "app",
				"version": "1.0.0",
				"os":      "linux",
				"arch":    "amd64",
			},
			wantTmpl:    "{name}_{version}_{os}_{arch}",
			description: "first template matches",
		},
		{
			name:      "multiple templates second match",
			templates: []string{"{name}_{version}_{os}_{arch}", "{name}-{os}-{arch}"},
			input:     "app-linux-amd64",
			wantMatch: true,
			wantFields: map[string]string{
				"name": "app",
				"os":   "linux",
				"arch": "amd64",
			},
			wantTmpl:    "{name}-{os}-{arch}",
			description: "second template matches",
		},
		{
			name:      "no match",
			templates: []string{"{name}-{version}-{arch}"},
			input:     "something-completely-different.zip",
			wantMatch: false,
			wantFields: map[string]string{
				"name": "app",
				"os":   "linux",
				"arch": "amd64",
			},
			description: "no template matches",
		},
		{
			name:      "escaped braces",
			templates: []string{"{name}-{{version}}-{arch}"},
			input:     "app-{version}-amd64",
			wantMatch: true,
			wantFields: map[string]string{
				"name": "app",
				"arch": "amd64",
			},
			wantTmpl:    "{name}-{{version}}-{arch}",
			description: "escaped braces treated as literals",
		},
		{
			name:      "complex version with build metadata",
			templates: []string{"{name}-{version}+{arch}"},
			input:     "myapp-1.2.3-beta.1+build.123+amd64",
			wantMatch: true,
			wantFields: map[string]string{
				"name":    "myapp",
				"version": "1.2.3-beta.1+build.123",
				"arch":    "amd64",
			},
			wantTmpl:    "{name}-{version}+{arch}",
			description: "complex semantic version with build metadata",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewReleaseParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			for _, tmpl := range tt.templates {
				if err := p.AddTemplate(tmpl); err != nil {
					t.Fatalf("Failed to add template %q: %v", tmpl, err)
				}
			}

			fields, tmpl, ok := p.TryParse(tt.input)

			if ok != tt.wantMatch {
				t.Errorf("TryParse(%q) match = %v, want %v", tt.input, ok, tt.wantMatch)
			}

			if !tt.wantMatch {
				return
			}

			if tmpl != tt.wantTmpl {
				t.Errorf("TryParse(%q) template = %q, want %q", tt.input, tmpl, tt.wantTmpl)
			}

			for k, wantV := range tt.wantFields {
				gotV, ok := fields[k]
				if !ok {
					t.Errorf("TryParse(%q) missing field %q", tt.input, k)
					continue
				}
				if gotV != wantV {
					t.Errorf("TryParse(%q) field %q = %q, want %q", tt.input, k, gotV, wantV)
				}
			}

			for k := range fields {
				if _, ok := tt.wantFields[k]; !ok {
					t.Errorf("TryParse(%q) unexpected field %q = %q", tt.input, k, fields[k])
				}
			}

			// Also test Parse() API
			result, err := p.Parse(tt.input)
			if err != nil {
				t.Errorf("Parse(%q) unexpected error: %v", tt.input, err)
			}
			if result.Raw != tt.input {
				t.Errorf("Parse(%q) Raw = %q, want %q", tt.input, result.Raw, tt.input)
			}
		})
	}
}

// TestNormalizers specifically tests normalization functions.
func TestNormalizers(t *testing.T) {
	tests := []struct {
		name      string
		fn        func(string) (string, error)
		input     string
		want      string
		wantError bool
	}{
		// Architecture normalizations
		{"arch x86_64", ArchNormalizer, "x86_64", "amd64", false},
		{"arch x86-64", ArchNormalizer, "x86-64", "amd64", false},
		{"arch amd64", ArchNormalizer, "amd64", "amd64", false},
		{"arch arm64", ArchNormalizer, "arm64", "arm64", false},
		{"arch aarch64", ArchNormalizer, "aarch64", "arm64", false},
		{"arch x86_64_v3", ArchNormalizer, "x86_64_v3", "amd64_v3", false},
		{"arch x86-64-v3", ArchNormalizer, "x86-64-v3", "amd64_v3", false},
		{"arch i386", ArchNormalizer, "i386", "386", false},
		{"arch i686", ArchNormalizer, "i686", "386", false},
		{"arch unknown", ArchNormalizer, "unknown", "unknown", false},
		{"arch case insensitive", ArchNormalizer, "AMD64", "amd64", false},

		// OS normalizations
		{"os linux", OSNormalizer, "linux", "linux", false},
		{"os darwin", OSNormalizer, "darwin", "darwin", false},
		{"os mac", OSNormalizer, "mac", "darwin", false},
		{"os macos", OSNormalizer, "macos", "darwin", false},
		{"os osx", OSNormalizer, "osx", "darwin", false},
		{"os windows", OSNormalizer, "windows", "windows", false},
		{"os win32", OSNormalizer, "win32", "windows", false},
		{"os win64", OSNormalizer, "win64", "windows", false},
		{"os freebsd", OSNormalizer, "freebsd", "freebsd", false},
		{"os unknown", OSNormalizer, "unknown", "unknown", false},
		{"os case insensitive", OSNormalizer, "DARWIN", "darwin", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fn(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("%s(%q) error = %v, wantError %v", tt.name, tt.input, err, tt.wantError)
				return
			}
			if got != tt.want {
				t.Errorf("%s(%q) = %q, want %q", tt.name, tt.input, got, tt.want)
			}
		})
	}
}

// TestErrorCases tests various error conditions.
func TestErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*Parser) error
		input     string
		wantError string
	}{
		{
			name: "empty template",
			setup: func(p *Parser) error {
				return p.AddTemplate("")
			},
			wantError: "empty template",
		},
		{
			name: "unclosed brace",
			setup: func(p *Parser) error {
				return p.AddTemplate("{name-{version}")
			},
			wantError: "unclosed '{'",
		},
		{
			name: "unmatched close brace",
			setup: func(p *Parser) error {
				return p.AddTemplate("{name}}")
			},
			wantError: "unmatched '}'",
		},
		{
			name: "empty field name",
			setup: func(p *Parser) error {
				return p.AddTemplate("{}-{version}")
			},
			wantError: "empty field name",
		},
		{
			name: "unclosed bracket",
			setup: func(p *Parser) error {
				return p.AddTemplate("{name}[-{version}")
			},
			wantError: "unclosed '['",
		},
		{
			name: "unmatched close bracket",
			setup: func(p *Parser) error {
				return p.AddTemplate("{name}]-{version}")
			},
			wantError: "unmatched ']'",
		},
		{
			name: "nested brackets",
			setup: func(p *Parser) error {
				return p.AddTemplate("{name}[-[{version}]?]?")
			},
			wantError: "nested optional segments",
		},
		{
			name: "invalid field name starting with digit",
			setup: func(p *Parser) error {
				return p.AddTemplate("{1name}-{version}")
			},
			wantError: "invalid field name",
		},
		{
			name: "template too long",
			setup: func(p *Parser) error {
				longTemplate := "{name}" + strings.Repeat("-", DefaultMaxTemplateLen)
				return p.AddTemplate(longTemplate)
			},
			wantError: "template too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewReleaseParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			err = tt.setup(p)
			if err == nil {
				t.Errorf("Expected error containing %q, got nil", tt.wantError)
				return
			}

			if !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("Error = %q, want to contain %q", err.Error(), tt.wantError)
			}
		})
	}
}

// TestInputValidation tests input validation.
func TestInputValidation(t *testing.T) {
	p, err := NewReleaseParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	if err := p.AddTemplate("{name}-{version}"); err != nil {
		t.Fatalf("Failed to add template: %v", err)
	}

	tests := []struct {
		name      string
		input     string
		wantError string
	}{
		{
			name:      "empty input",
			input:     "",
			wantError: "empty input",
		},
		{
			name:      "whitespace only",
			input:     "   ",
			wantError: "empty input",
		},
		{
			name:      "input too long",
			input:     strings.Repeat("a", DefaultMaxInputLength+1),
			wantError: "input too long",
		},
		{
			name:      "null byte",
			input:     "app-1.0.0\x00-amd64",
			wantError: "null byte",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, ok := p.TryParse(tt.input)
			if ok {
				t.Errorf("TryParse(%q) succeeded, want error", tt.input)
			}

			_, err := p.Parse(tt.input)
			if err == nil {
				t.Errorf("Parse(%q) succeeded, want error", tt.input)
				return
			}

			if !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("Parse(%q) error = %q, want to contain %q", tt.input, err.Error(), tt.wantError)
			}
		})
	}
}

// TestThreadSafety tests concurrent access.
func TestThreadSafety(t *testing.T) {
	p, err := NewReleaseParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	templates := []string{
		"{name}_{version}_{os}_{arch}",
		"{name}-{version}-{os}-{arch}",
		"{name}-{os}-{arch}",
	}

	for _, tmpl := range templates {
		if err := p.AddTemplate(tmpl); err != nil {
			t.Fatalf("Failed to add template: %v", err)
		}
	}

	p.Seal()

	inputs := []string{
		"app_1.0.0_linux_amd64",
		"tool-2.3.4-darwin-arm64",
		"binary-windows-amd64",
		"hugo_0.151.1_linux-amd64.tar.gz",
	}

	// Concurrent parsing
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				input := inputs[j%len(inputs)]
				p.TryParse(input)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestSealedParser tests that sealed parsers reject new templates.
func TestSealedParser(t *testing.T) {
	p, err := NewReleaseParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	if err := p.AddTemplate("{name}-{version}"); err != nil {
		t.Fatalf("Failed to add template: %v", err)
	}

	p.Seal()

	err = p.AddTemplate("{name}-{arch}")
	if err == nil {
		t.Error("Expected error when adding template to sealed parser")
	}

	if !strings.Contains(err.Error(), "sealed") {
		t.Errorf("Error = %q, want to contain 'sealed'", err.Error())
	}
}

// TestBuilderPattern tests the builder pattern.
func TestBuilderPattern(t *testing.T) {
	p, err := NewParserBuilder().
		WithField("name", FieldSpec{Pattern: PatIdent}).
		WithField("version", FieldSpec{Pattern: PatVersion}).
		WithMaxInputLength(256).
		WithParseTimeout(50 * time.Millisecond).
		Build()

	if err != nil {
		t.Fatalf("Failed to build parser: %v", err)
	}

	if p.maxInputLen != 256 {
		t.Errorf("maxInputLen = %d, want 256", p.maxInputLen)
	}

	if p.parseTimeout != 50*time.Millisecond {
		t.Errorf("parseTimeout = %v, want 50ms", p.parseTimeout)
	}
}

// TestLogging tests that logging doesn't cause panics.
func TestLogging(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)

	p, err := NewParserBuilder().
		WithField("name", FieldSpec{Pattern: PatIdent}).
		WithField("version", FieldSpec{Pattern: PatVersion}).
		WithLogger(logger).
		Build()

	if err != nil {
		t.Fatalf("Failed to build parser: %v", err)
	}

	if err := p.AddTemplate("{name}-{version}"); err != nil {
		t.Fatalf("Failed to add template: %v", err)
	}

	p.TryParse("app-1.0.0")
}

// MockMetrics implements MetricsCollector for testing.
type MockMetrics struct {
	parseAttempts      int
	timeouts           int
	compilations       int
	failedCompilations int
}

func (m *MockMetrics) RecordParseAttempt(template string, success bool, duration time.Duration) {
	m.parseAttempts++
}

func (m *MockMetrics) RecordRegexTimeout(template string) {
	m.timeouts++
}

func (m *MockMetrics) RecordTemplateCompilation(template string, success bool) {
	m.compilations++
	if !success {
		m.failedCompilations++
	}
}

// TestMetrics tests metrics collection.
func TestMetrics(t *testing.T) {
	metrics := &MockMetrics{}

	p, err := NewParserBuilder().
		WithField("name", FieldSpec{Pattern: PatIdent}).
		WithField("version", FieldSpec{Pattern: PatVersion}).
		WithMetrics(metrics).
		Build()

	if err != nil {
		t.Fatalf("Failed to build parser: %v", err)
	}

	// Add a template
	if err := p.AddTemplate("{name}-{version}"); err != nil {
		t.Fatalf("Failed to add template: %v", err)
	}

	if metrics.compilations != 1 {
		t.Errorf("compilations = %d, want 1", metrics.compilations)
	}

	// Parse successfully
	p.TryParse("app-1.0.0")

	if metrics.parseAttempts != 1 {
		t.Errorf("parseAttempts = %d, want 1", metrics.parseAttempts)
	}

	// Parse unsuccessfully
	p.TryParse("nomatch")

	if metrics.parseAttempts != 2 {
		t.Errorf("parseAttempts = %d, want 2", metrics.parseAttempts)
	}

	// Try to add a bad template
	_ = p.AddTemplate("{unclosed")

	if metrics.failedCompilations != 1 {
		t.Errorf("failedCompilations = %d, want 1", metrics.failedCompilations)
	}
}

// BenchmarkParse benchmarks the parsing operation.
func BenchmarkParse(b *testing.B) {
	p, err := NewReleaseParser()
	if err != nil {
		b.Fatalf("Failed to create parser: %v", err)
	}

	templates := []string{
		"{name}_{version}_{os}_{arch}.tar.gz",
		"{name}-{version}-{os}-{arch}.tar.gz",
		"{name}-{os}-{arch}",
	}

	for _, tmpl := range templates {
		if err := p.AddTemplate(tmpl); err != nil {
			b.Fatalf("Failed to add template: %v", err)
		}
	}

	inputs := []string{
		"hugo_0.151.1_linux-amd64.tar.gz",
		"golangci-lint-1.64.6-darwin-arm64.tar.gz",
		"frankenphp-linux-x86_64",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := inputs[i%len(inputs)]
		p.TryParse(input)
	}
}

// BenchmarkParseWithNormalization benchmarks parsing with normalization.
func BenchmarkParseWithNormalization(b *testing.B) {
	p, err := NewReleaseParser()
	if err != nil {
		b.Fatalf("Failed to create parser: %v", err)
	}

	if err := p.AddTemplate("{name}-{version}-{os}-{arch}"); err != nil {
		b.Fatalf("Failed to add template: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.TryParse("app-1.0.0-darwin-aarch64")
	}
}
