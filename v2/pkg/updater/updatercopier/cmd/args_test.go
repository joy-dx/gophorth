package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestSplitArgsString_Golden(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		want    []string
		wantErr bool
	}{
		{
			name: "empty",
			in:   "",
			want: nil,
		},
		{
			name: "simple",
			in:   "a b c",
			want: []string{"a", "b", "c"},
		},
		{
			name: "quotes",
			in:   `"a b" c`,
			want: []string{"a b", "c"},
		},
		{
			name: "escaped_in_quotes",
			in:   `"a\"b" c`,
			want: []string{`a"b`, "c"},
		},
		{
			name:    "unterminated_quote",
			in:      `"a b`,
			wantErr: true,
		},
		{
			name:    "dangling_escape",
			in:      `"a\`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := splitArgsString(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil; out=%v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v want %#v", got, tc.want)
			}
		})
	}
}

func TestParseArgs_Golden(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		argv          []string
		wantOld       string
		wantNew       string
		wantLog       string
		wantLaunch    []string
		wantErrSubstr string
	}{
		// ... your existing cases ...
		{
			name:          "unknown_token",
			argv:          []string{"u", "old", "new", "--nope"},
			wantErrSubstr: "unknown token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			oldP, newP, logP, launch, err := parseArgs(tc.argv)
			if tc.wantErrSubstr != "" {
				if err == nil {
					t.Fatalf(
						"expected err containing %q, got nil",
						tc.wantErrSubstr,
					)
				}
				if !strings.Contains(err.Error(), tc.wantErrSubstr) {
					t.Fatalf(
						"err=%q does not contain %q",
						err.Error(),
						tc.wantErrSubstr,
					)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if oldP != tc.wantOld || newP != tc.wantNew || logP != tc.wantLog {
				t.Fatalf(
					"got old=%q new=%q log=%q; want old=%q new=%q log=%q",
					oldP,
					newP,
					logP,
					tc.wantOld,
					tc.wantNew,
					tc.wantLog,
				)
			}
			if !reflect.DeepEqual(launch, tc.wantLaunch) {
				t.Fatalf("launch got %#v want %#v", launch, tc.wantLaunch)
			}
		})
	}
}
