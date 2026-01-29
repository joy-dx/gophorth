package parser

import (
	"errors"
	"testing"
)

func mustCompile(t *testing.T, pattern string) *FilenameParser {
	t.Helper()

	p, err := CompileFilenamePattern(pattern)
	if err != nil {
		t.Fatalf("CompileFilenamePattern(%q) error: %v", pattern, err)
	}
	return p
}

func TestCompileFilenamePattern_InvalidPatterns(t *testing.T) {
	t.Parallel()

	golden := []struct {
		name    string
		pattern string
		wantErr error
	}{
		{
			name:    "empty pattern",
			pattern: "",
			wantErr: ErrInvalidPattern,
		},
		{
			name:    "whitespace pattern",
			pattern: "   ",
			wantErr: ErrInvalidPattern,
		},
		{
			name:    "unclosed tag",
			pattern: "tool-{:version",
			wantErr: ErrInvalidPattern,
		},
		{
			name:    "unsupported tag",
			pattern: "tool-{:tool}-{:version}",
			wantErr: ErrInvalidPattern,
		},
		{
			name:    "variant empty list",
			pattern: `tool-{:variant[]}`,
			wantErr: ErrInvalidPattern,
		},
		{
			name:    "variant values must be backtick-quoted",
			pattern: `tool-{:variant[static,dynamic]}`,
			wantErr: ErrInvalidPattern,
		},
		{
			name:    "variant missing comma",
			pattern: "tool-{:variant[`a` `b`]}",
			wantErr: ErrInvalidPattern,
		},
		{
			name:    "unclosed optional group",
			pattern: "tool{?-{:platform}",
			wantErr: ErrInvalidPattern,
		},
		{
			name:    "nested optional groups not supported",
			pattern: "tool{?x{?y}}",
			wantErr: ErrInvalidPattern,
		},
		{
			name:    "variant unterminated quote",
			pattern: "tool-{:variant[`a]}",
			wantErr: ErrInvalidPattern,
		},
	}

	for _, tc := range golden {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := CompileFilenamePattern(tc.pattern)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected errors.Is(err, %v)=true, got err=%v", tc.wantErr, err)
			}
		})
	}
}

func TestOptionalVariantEnum(t *testing.T) {
	t.Parallel()

	p := mustCompile(t, "tool-{:variant[`a`,`b`]?}-{:platform}")

	cases := []struct {
		name     string
		filename string
		wantVar  string
		wantPlat string
		wantErr  error
	}{
		{
			name:     "variant present",
			filename: "tool-a-linux",
			wantVar:  "a",
			wantPlat: "linux",
		},
		{
			name:     "variant absent (double dash)",
			filename: "tool--linux",
			wantVar:  "",
			wantPlat: "linux",
		},
		{
			name:     "variant not allowed",
			filename: "tool-c-linux",
			wantErr:  ErrNoMatch,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := p.ParseFilename(tc.filename)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.wantErr)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf(
						"expected errors.Is(err, %v)=true, got %v",
						tc.wantErr,
						err,
					)
				}
				return
			}

			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if got.Artifact.Variant != tc.wantVar {
				t.Fatalf(
					"Variant: want %q, got %q",
					tc.wantVar,
					got.Artifact.Variant,
				)
			}
			if got.Artifact.Platform != tc.wantPlat {
				t.Fatalf(
					"Platform: want %q, got %q",
					tc.wantPlat,
					got.Artifact.Platform,
				)
			}
		})
	}
}

func TestParseFilename_NoMatchAndInputHygiene(t *testing.T) {
	t.Parallel()

	parser := mustCompile(t, `tool-{:version}-{:platform}-{:arch}.{:format}`)

	golden := []struct {
		name     string
		filename string
		wantErr  error
	}{
		{
			name:     "empty filename",
			filename: "",
			wantErr:  ErrNoMatch,
		},
		{
			name:     "whitespace filename",
			filename: "   ",
			wantErr:  ErrNoMatch,
		},
		{
			name:     "newline attack LF",
			filename: "tool-1.2.3-linux-amd64.tar.gz\njunk",
			wantErr:  ErrNoMatch,
		},
		{
			name:     "newline attack CRLF",
			filename: "tool-1.2.3-linux-amd64.tar.gz\r\njunk",
			wantErr:  ErrNoMatch,
		},
		{
			name:     "pattern mismatch",
			filename: "tool-1.2.3.tar.gz",
			wantErr:  ErrNoMatch,
		},
		{
			name:     "extra suffix should not match (anchored)",
			filename: "tool-1.2.3-linux-amd64.tar.gz.sig",
			wantErr:  ErrNoMatch,
		},
		{
			name:     "unknown platform",
			filename: "tool-1.2.3-sunos-amd64.tar.gz",
			wantErr:  ErrNoMatch, // won't match regex alternation => no match, not unknown
		},
		{
			name:     "unknown arch",
			filename: "tool-1.2.3-linux-riscv64.tar.gz",
			wantErr:  ErrNoMatch, // won't match regex alternation => no match, not unknown
		},
		{
			name:     "unknown format",
			filename: "tool-1.2.3-linux-amd64.7z",
			wantErr:  ErrNoMatch, // won't match regex alternation => no match, not unknown
		},
	}

	for _, tc := range golden {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := parser.ParseFilename(tc.filename)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected errors.Is(err, %v)=true, got err=%v", tc.wantErr, err)
			}
		})
	}
}

func TestParseFilename_HappyPaths(t *testing.T) {
	t.Parallel()

	golden := []struct {
		name     string
		pattern  string
		filename string
		want     Artifact
	}{
		{
			name:     "node semver with v prefix and tar.gz",
			pattern:  `node-v{:version}-{:platform}-{:arch}.{:format}`,
			filename: `node-v20.1.0-darwin-arm64.tar.gz`,
			want: Artifact{
				Version:  "20.1.0",
				Platform: "darwin",
				Arch:     "arm64",
				Format:   "tar.gz",
			},
		},
		{
			name:     "node semver without v prefix and tar.xz",
			pattern:  `node-v{:version}-{:platform}-{:arch}.{:format}`,
			filename: `node-v18.20.8-linux-x64.tar.xz`,
			want: Artifact{
				Version:  "18.20.8",
				Platform: "linux",
				Arch:     "amd64", // x64 normalized to amd64
				Format:   "tar.xz",
			},
		},
		{
			name:     "windows platform normalization (win32) and arch normalization (x86)",
			pattern:  `node-v{:version}-{:platform}-{:arch}.{:format}`,
			filename: `node-v22.1.0-win32-x86.zip`,
			want: Artifact{
				Version:  "22.1.0",
				Platform: "windows",
				Arch:     "386",
				Format:   "zip",
			},
		},
		{
			name:     "go release style (calver-like segments) and tgz",
			pattern:  `go{:version}.{:platform}-{:arch}.{:format}`,
			filename: `go1.25.5.darwin-arm64.tgz`,
			want: Artifact{
				Version:  "1.25.5",
				Platform: "darwin",
				Arch:     "arm64",
				Format:   "tgz",
			},
		},
		{
			name:     "semver prerelease and build metadata",
			pattern:  `tool-{:version}-{:platform}-{:arch}.{:format}`,
			filename: `tool-v3.0.0-beta.11+build.7-linux-amd64.tar.zst`,
			want: Artifact{
				Version:  "v3.0.0-beta.11+build.7",
				Platform: "linux",
				Arch:     "amd64",
				Format:   "tar.zst",
			},
		},
		{
			name:     "ext tag standalone lowercased",
			pattern:  `file-{:platform}.{:ext}`,
			filename: `file-macOS.GZ`,
			want: Artifact{
				Platform: "darwin",
				Ext:      "gz",
			},
		},
		{
			name:     "variant enum basic",
			pattern:  "jq-{:platform}-{:variant[`static`,`dynamic`]}",
			filename: `jq-linux-static`,
			want: Artifact{
				Platform: "linux",
				Variant:  "static",
			},
		},
		{
			name:     "variant enum is case-sensitive",
			pattern:  "jq-{:platform}-{:variant[`static`,`dynamic`]}",
			filename: `jq-linux-Static`,
			want:     Artifact{}, // should fail; asserted in separate test below
		},
	}

	for _, tc := range golden {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Special-case the “should fail” row to keep the table compact.
			if tc.name == "variant enum is case-sensitive" {
				p := mustCompile(t, tc.pattern)
				_, err := p.ParseFilename(tc.filename)
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !errors.Is(err, ErrNoMatch) {
					t.Fatalf("expected ErrNoMatch, got %v", err)
				}
				return
			}

			p := mustCompile(t, tc.pattern)
			got, err := p.ParseFilename(tc.filename)
			if err != nil {
				t.Fatalf("ParseFilename(%q) error: %v", tc.filename, err)
			}

			if got.Artifact.Platform != tc.want.Platform {
				t.Fatalf("Platform: want %q, got %q", tc.want.Platform, got.Artifact.Platform)
			}
			if got.Artifact.Arch != tc.want.Arch {
				t.Fatalf("Arch: want %q, got %q", tc.want.Arch, got.Artifact.Arch)
			}
			if got.Artifact.Format != tc.want.Format {
				t.Fatalf("Format: want %q, got %q", tc.want.Format, got.Artifact.Format)
			}
			if got.Artifact.Version != tc.want.Version {
				t.Fatalf("Version: want %q, got %q", tc.want.Version, got.Artifact.Version)
			}
			if got.Artifact.Ext != tc.want.Ext {
				t.Fatalf("Ext: want %q, got %q", tc.want.Ext, got.Artifact.Ext)
			}
			if got.Artifact.Variant != tc.want.Variant {
				t.Fatalf("Variant: want %q, got %q", tc.want.Variant, got.Artifact.Variant)
			}
		})
	}
}

func TestOptionalFields(t *testing.T) {
	t.Parallel()

	golden := []struct {
		name     string
		pattern  string
		filename string
		want     Artifact
	}{
		{
			name:     "optional version present",
			pattern:  `tool-{:version?}-{:platform}`,
			filename: `tool-1.2.3-linux`,
			want: Artifact{
				Version:  "1.2.3",
				Platform: "linux",
			},
		},
		{
			name:     "optional version absent",
			pattern:  `tool-{:version?}-{:platform}`,
			filename: `tool--linux`,
			want: Artifact{
				Platform: "linux",
			},
		},
		{
			name:     "optional platform absent in middle",
			pattern:  `tool-{:platform?}-{:arch}`,
			filename: `tool--amd64`,
			want: Artifact{
				Arch: "amd64",
			},
		},
		{
			name:     "optional arch absent at end",
			pattern:  `tool-{:platform}-{:arch?}`,
			filename: `tool-linux-`,
			want: Artifact{
				Platform: "linux",
			},
		},
		{
			name:     "all optional fields absent (still matches literals)",
			pattern:  `x{:version?}-{:platform?}-{:arch?}`,
			filename: `x--`,
			want:     Artifact{},
		},
	}

	for _, tc := range golden {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := mustCompile(t, tc.pattern)
			got, err := p.ParseFilename(tc.filename)
			if err != nil {
				t.Fatalf("ParseFilename(%q) error: %v", tc.filename, err)
			}

			if got.Artifact.Version != tc.want.Version {
				t.Fatalf("Version: want %q, got %q", tc.want.Version, got.Artifact.Version)
			}
			if got.Artifact.Platform != tc.want.Platform {
				t.Fatalf("Platform: want %q, got %q", tc.want.Platform, got.Artifact.Platform)
			}
			if got.Artifact.Arch != tc.want.Arch {
				t.Fatalf("Arch: want %q, got %q", tc.want.Arch, got.Artifact.Arch)
			}
		})
	}
}

func TestCaseSensitivityOption(t *testing.T) {
	t.Parallel()

	golden := []struct {
		name          string
		opts          CompileOptions
		pattern       string
		filename      string
		wantErr       error
		wantPlatform  string
		wantArch      string
		wantFormat    string
		wantVersion   string
		wantIsNoMatch bool
	}{
		{
			name:         "default is case-insensitive for platform/arch tokens",
			opts:         DefaultOptions(),
			pattern:      `tool-{:platform}-{:arch}.{:format}`,
			filename:     `tool-macOS-ARM64.tar.gz`,
			wantErr:      nil,
			wantPlatform: "darwin",
			wantArch:     "arm64",
			wantFormat:   "tar.gz",
		},
		{
			name:          "case-sensitive option rejects different casing for platform",
			opts:          CompileOptions{CaseSensitive: true},
			pattern:       `tool-{:platform}-{:arch}.{:format}`,
			filename:      `tool-macOS-arm64.tar.gz`,
			wantErr:       ErrNoMatch,
			wantIsNoMatch: true,
		},
		{
			name:          "case-sensitive option rejects different casing for arch",
			opts:          CompileOptions{CaseSensitive: true},
			pattern:       `tool-{:platform}-{:arch}.{:format}`,
			filename:      `tool-darwin-ARM64.tar.gz`,
			wantErr:       ErrNoMatch,
			wantIsNoMatch: true,
		},
		{
			name:         "format matching is always case-insensitive (extensions)",
			opts:         CompileOptions{CaseSensitive: true},
			pattern:      `tool-{:platform}-{:arch}.{:format}`,
			filename:     `tool-darwin-arm64.TAR.GZ`,
			wantErr:      nil,
			wantPlatform: "darwin",
			wantArch:     "arm64",
			wantFormat:   "tar.gz",
		},
	}

	for _, tc := range golden {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p, err := CompileFilenamePatternWithOptions(tc.pattern, tc.opts)
			if err != nil {
				t.Fatalf("compile error: %v", err)
			}

			got, err := p.ParseFilename(tc.filename)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.wantErr)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected errors.Is(err, %v)=true, got %v", tc.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			if got.Artifact.Platform != tc.wantPlatform {
				t.Fatalf("Platform: want %q, got %q", tc.wantPlatform, got.Artifact.Platform)
			}
			if got.Artifact.Arch != tc.wantArch {
				t.Fatalf("Arch: want %q, got %q", tc.wantArch, got.Artifact.Arch)
			}
			if got.Artifact.Format != tc.wantFormat {
				t.Fatalf("Format: want %q, got %q", tc.wantFormat, got.Artifact.Format)
			}
			if tc.wantVersion != "" && got.Artifact.Version != tc.wantVersion {
				t.Fatalf("Version: want %q, got %q", tc.wantVersion, got.Artifact.Version)
			}
		})
	}
}

func TestFormatLongestFirst(t *testing.T) {
	t.Parallel()

	// Ensures tar.gz beats gz if both could appear.
	// Since {:format} only matches known formats, "tar.gz" must be captured as-is.
	p := mustCompile(t, `tool-{:platform}-{:arch}.{:format}`)

	golden := []struct {
		name     string
		filename string
		wantFmt  string
	}{
		{
			name:     "tar.gz captured",
			filename: "tool-linux-amd64.tar.gz",
			wantFmt:  "tar.gz",
		},
		{
			name:     "gz captured",
			filename: "tool-linux-amd64.gz",
			wantFmt:  "gz",
		},
		{
			name:     "tar.zst captured",
			filename: "tool-linux-amd64.tar.zst",
			wantFmt:  "tar.zst",
		},
		{
			name:     "zst captured",
			filename: "tool-linux-amd64.zst",
			wantFmt:  "zst",
		},
	}

	for _, tc := range golden {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := p.ParseFilename(tc.filename)
			if err != nil {
				t.Fatalf("ParseFilename error: %v", err)
			}
			if got.Artifact.Format != tc.wantFmt {
				t.Fatalf("Format: want %q, got %q", tc.wantFmt, got.Artifact.Format)
			}
		})
	}
}

func TestParseFilenamesFromLines(t *testing.T) {
	t.Parallel()

	p := mustCompile(t, `tool-{:version}-{:platform}-{:arch}.{:format}`)

	input := "" +
		"tool-1.2.3-linux-amd64.tar.gz\n" +
		"tool-9.9.9-darwin-arm64.zip, some extra columns\n" +
		"not-a-match\n" +
		"\n" +
		"tool-1.0.0-win64-x86.zip\n"

	got := p.ParseFilenamesFromLines(input)

	if len(got.Artifacts) != 3 {
		t.Fatalf("Artifacts: want %d, got %d", 3, len(got.Artifacts))
	}
	if len(got.Errors) != 1 {
		t.Fatalf("Errors: want %d, got %d", 1, len(got.Errors))
	}

	// Golden expectations for extracted artifacts
	want := []Artifact{
		{Version: "1.2.3", Platform: "linux", Arch: "amd64", Format: "tar.gz"},
		{Version: "9.9.9", Platform: "darwin", Arch: "arm64", Format: "zip"},
		{Version: "1.0.0", Platform: "windows", Arch: "386", Format: "zip"},
	}

	for i := range want {
		if got.Artifacts[i].Version != want[i].Version {
			t.Fatalf("Artifacts[%d].Version: want %q, got %q",
				i, want[i].Version, got.Artifacts[i].Version)
		}
		if got.Artifacts[i].Platform != want[i].Platform {
			t.Fatalf("Artifacts[%d].Platform: want %q, got %q",
				i, want[i].Platform, got.Artifacts[i].Platform)
		}
		if got.Artifacts[i].Arch != want[i].Arch {
			t.Fatalf("Artifacts[%d].Arch: want %q, got %q",
				i, want[i].Arch, got.Artifacts[i].Arch)
		}
		if got.Artifacts[i].Format != want[i].Format {
			t.Fatalf("Artifacts[%d].Format: want %q, got %q",
				i, want[i].Format, got.Artifacts[i].Format)
		}
	}

	// Error details (line numbers are 1-based)
	if got.Errors[0].Line != 3 {
		t.Fatalf("Errors[0].Line: want %d, got %d", 3, got.Errors[0].Line)
	}
	if got.Errors[0].Filename != "not-a-match" {
		t.Fatalf("Errors[0].Filename: want %q, got %q",
			"not-a-match", got.Errors[0].Filename)
	}
	if !errors.Is(got.Errors[0].Err, ErrNoMatch) {
		t.Fatalf("Errors[0].Err: want ErrNoMatch, got %v", got.Errors[0].Err)
	}
}

func TestOptionalGroups_VariantAndFormat(t *testing.T) {
	t.Parallel()

	// Accept:
	// - joydx-darwin-amd64
	// - joydx-darwin-amd64.zip
	// - joydx-darwin-amd64-webkit241
	// - joydx-darwin-amd64-webkit241.zip
	p := mustCompile(
		t,
		"joydx-{:platform}-{:arch}{?-{:variant[`webkit241`]}}{?.{:format}}",
	)

	cases := []struct {
		name     string
		filename string
		wantVar  string
		wantFmt  string
		wantErr  error
	}{
		{
			name:     "no variant, no format",
			filename: "joydx-darwin-amd64",
			wantVar:  "",
			wantFmt:  "",
		},
		{
			name:     "no variant, format",
			filename: "joydx-darwin-amd64.zip",
			wantVar:  "",
			wantFmt:  "zip",
		},
		{
			name:     "variant, no format",
			filename: "joydx-darwin-amd64-webkit241",
			wantVar:  "webkit241",
			wantFmt:  "",
		},
		{
			name:     "variant and format",
			filename: "joydx-darwin-amd64-webkit241.zip",
			wantVar:  "webkit241",
			wantFmt:  "zip",
		},
		{
			name:     "wrong variant rejected",
			filename: "joydx-darwin-amd64-webkit999.zip",
			wantErr:  ErrNoMatch,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := p.ParseFilename(tc.filename)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.wantErr)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected errors.Is(err, %v)=true, got %v", tc.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if got.Artifact.Platform != "darwin" {
				t.Fatalf("Platform: want %q, got %q", "darwin", got.Artifact.Platform)
			}
			if got.Artifact.Arch != "amd64" {
				t.Fatalf("Arch: want %q, got %q", "amd64", got.Artifact.Arch)
			}
			if got.Artifact.Variant != tc.wantVar {
				t.Fatalf("Variant: want %q, got %q", tc.wantVar, got.Artifact.Variant)
			}
			if got.Artifact.Format != tc.wantFmt {
				t.Fatalf("Format: want %q, got %q", tc.wantFmt, got.Artifact.Format)
			}
		})
	}
}
