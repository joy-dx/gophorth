package stringz

import "fmt"

func PadRight(s string, width int) string {
	return fmt.Sprintf("%-*s", width, s)
}
