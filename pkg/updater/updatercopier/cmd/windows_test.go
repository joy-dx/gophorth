//go:build windows

package main

import "testing"

func TestEscapeForCmdLiteral_Golden(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{in: `C:\a\b\c.exe`, want: `C:\a\b\c.exe`},
		{in: `C:\a & b\c.exe`, want: `C:\a ^& b\c.exe`},
		{in: `C:\a^b\c.exe`, want: `C:\a^^b\c.exe`},
		{in: `C:\a(b)\c.exe`, want: `C:\a^(b^)\c.exe`},
		{in: `C:\a!b\c.exe`, want: `C:\a^!b\c.exe`},
		{in: `C:\a"b"\c.exe`, want: `C:\a^"b^"\c.exe`},
	}

	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got := escapeForCmdLiteral(tc.in)
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}
