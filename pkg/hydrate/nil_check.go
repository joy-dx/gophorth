package hydrate

import "strings"

func NilCheck(serviceName string, toCheck map[string]interface{}) error {
	var sb strings.Builder
	var nilFound bool
	for key, toCheckItem := range toCheck {
		if toCheckItem == nil {
			nilFound = true
			sb.WriteString(key)
			sb.WriteString(" ")
		}
	}
	if nilFound {
		return &ServiceHydrateError{
			Service: serviceName,
			Problem: sb.String(),
		}
	}
	return nil
}
