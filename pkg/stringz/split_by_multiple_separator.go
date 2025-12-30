package stringz

import "strings"

func SplitByMultipleSeparator(text string, separators string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return strings.ContainsRune(separators, r)
	})
}
