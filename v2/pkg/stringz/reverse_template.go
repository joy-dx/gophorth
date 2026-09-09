package stringz

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type ReverseTemplateOptions struct {
	Pattern           string
	AllowAnyExtension bool
	RequireVersion    bool
}

func CompileReverseTemplate(opts ReverseTemplateOptions) (*regexp.Regexp, error) {
	if strings.TrimSpace(opts.Pattern) == "" {
		return nil, errors.New("pattern must not be empty")
	}

	var versionRule string
	if opts.RequireVersion {
		versionRule = `-[0-9]+(?:\.[0-9A-Za-z]+)*(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?`
	} else {
		versionRule = `(?:-[0-9]+(?:\.[0-9A-Za-z]+)*(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?)?`
	}

	rules := map[string]string{
		"platform": `[a-z0-9]+`,
		"arch":     `[A-Za-z0-9_]+`,
		"variant":  `(?:-[A-Za-z0-9._]+)?`,
		"version":  versionRule,
	}

	var b strings.Builder
	b.WriteString("^")

	for i := 0; i < len(opts.Pattern); {
		if opts.Pattern[i] != '{' {
			j := strings.IndexByte(opts.Pattern[i:], '{')
			if j == -1 {
				j = len(opts.Pattern)
			} else {
				j = i + j
			}
			b.WriteString(regexp.QuoteMeta(opts.Pattern[i:j]))
			i = j
			continue
		}

		end := strings.IndexByte(opts.Pattern[i:], '}')
		if end == -1 {
			return nil, fmt.Errorf("unclosed placeholder starting at index %d", i)
		}
		end = i + end

		name := opts.Pattern[i+1 : end]
		if name == "" {
			return nil, fmt.Errorf("empty placeholder at index %d", i)
		}

		rule, ok := rules[name]
		if !ok {
			return nil, fmt.Errorf("unknown placeholder {%s}", name)
		}

		b.WriteString("(?P<")
		b.WriteString(name)
		b.WriteString(">")
		b.WriteString(rule)
		b.WriteString(")")

		i = end + 1
	}

	if opts.AllowAnyExtension {
		// Allow ".zip", ".tar.gz", etc. after the pattern.
		// .asc/.asc.sig are filtered by caller since Go regex lacks lookarounds.
		b.WriteString(`(?:\.[A-Za-z0-9]+(?:\.[A-Za-z0-9]+)*)?`)
	}

	b.WriteString("$")

	re, err := regexp.Compile(b.String())
	if err != nil {
		return nil, fmt.Errorf("compile regex from pattern %q: %w", opts.Pattern, err)
	}
	return re, nil
}
