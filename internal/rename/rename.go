package rename

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"tin/internal/shared/fsx"
)

type Style string

const (
	StyleKebab     Style = "kebab"
	StyleSnake     Style = "snake"
	StyleCamel     Style = "camel"
	StylePascal    Style = "pascal"
	StyleScreaming Style = "screaming"
)

const DefaultStyle = StyleKebab

var styles = map[Style]struct{}{
	StyleKebab:     {},
	StyleSnake:     {},
	StyleCamel:     {},
	StylePascal:    {},
	StyleScreaming: {},
}

func ParseStyle(s string) (Style, error) {
	if s == "" {
		return DefaultStyle, nil
	}
	style := Style(strings.ToLower(s))
	if _, ok := styles[style]; !ok {
		return "", fmt.Errorf("unsupported style %q; use kebab, snake, camel, pascal, or screaming", s)
	}
	return style, nil
}

type Move struct {
	Source      string
	Destination string
}

type Plan struct {
	Moves []Move
	NoOps []string
}

type Result struct {
	Moved int
}

func BuildPlan(paths []string, style Style) (Plan, error) {
	plan := Plan{}
	taken := make(map[string]struct{})

	for _, source := range paths {
		clean := filepath.Clean(source)
		if _, err := os.Lstat(clean); err != nil {
			return Plan{}, err
		}

		dir, base := filepath.Split(clean)
		newBase, err := NormalizeBasename(base, style)
		if err != nil {
			return Plan{}, fmt.Errorf("%s: %w", clean, err)
		}

		if newBase == base {
			plan.NoOps = append(plan.NoOps, clean)
			continue
		}

		dest, err := chooseUniqueDest(dir, newBase, clean, taken)
		if err != nil {
			return Plan{}, err
		}
		taken[dest] = struct{}{}
		plan.Moves = append(plan.Moves, Move{Source: clean, Destination: dest})
	}

	return plan, nil
}

func Apply(plan Plan) (Result, error) {
	moved := 0
	for _, move := range plan.Moves {
		if err := os.Rename(move.Source, move.Destination); err != nil {
			return Result{Moved: moved}, fmt.Errorf("rename %q -> %q: %w", move.Source, move.Destination, err)
		}
		moved++
	}
	return Result{Moved: moved}, nil
}

func NormalizeBasename(base string, style Style) (string, error) {
	dotPrefix := ""
	rest := base
	for strings.HasPrefix(rest, ".") {
		dotPrefix += "."
		rest = rest[1:]
	}
	if rest == "" {
		return base, nil
	}

	ext := filepath.Ext(rest)
	stem := strings.TrimSuffix(rest, ext)

	tokens := tokenize(stem)
	if len(tokens) == 0 {
		return "", fmt.Errorf("name %q has no letters or digits", base)
	}

	return dotPrefix + joinStyle(tokens, style) + strings.ToLower(ext), nil
}

func tokenize(s string) []string {
	runes := []rune(s)
	var tokens []string
	var current []rune

	flush := func() {
		if len(current) > 0 {
			tokens = append(tokens, string(current))
			current = current[:0]
		}
	}

	for i, r := range runes {
		if !isAlnum(r) {
			flush()
			continue
		}
		if len(current) > 0 {
			prev := current[len(current)-1]
			if unicode.IsLower(prev) && unicode.IsUpper(r) {
				flush()
			} else if unicode.IsUpper(prev) && unicode.IsUpper(r) && i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
				flush()
			}
		}
		current = append(current, r)
	}
	flush()
	return tokens
}

func joinStyle(tokens []string, style Style) string {
	switch style {
	case StyleKebab:
		return joinLower(tokens, "-")
	case StyleSnake:
		return joinLower(tokens, "_")
	case StyleScreaming:
		return joinUpper(tokens, "_")
	case StyleCamel:
		return joinCamel(tokens, false)
	case StylePascal:
		return joinCamel(tokens, true)
	default:
		return joinLower(tokens, "-")
	}
}

func joinLower(tokens []string, sep string) string {
	parts := make([]string, len(tokens))
	for i, t := range tokens {
		parts[i] = strings.ToLower(t)
	}
	return strings.Join(parts, sep)
}

func joinUpper(tokens []string, sep string) string {
	parts := make([]string, len(tokens))
	for i, t := range tokens {
		parts[i] = strings.ToUpper(t)
	}
	return strings.Join(parts, sep)
}

func joinCamel(tokens []string, capitalizeFirst bool) string {
	var b strings.Builder
	for i, t := range tokens {
		if i == 0 && !capitalizeFirst {
			b.WriteString(strings.ToLower(t))
			continue
		}
		b.WriteString(title(t))
	}
	return b.String()
}

func title(s string) string {
	runes := []rune(strings.ToLower(s))
	if len(runes) == 0 {
		return ""
	}
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func isAlnum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r)
}

func chooseUniqueDest(dir string, base string, source string, taken map[string]struct{}) (string, error) {
	srcInfo, err := os.Lstat(source)
	if err != nil {
		return "", err
	}
	for i := 0; ; i++ {
		candidate := filepath.Join(dir, fsx.UniqueNameWithExt(base, i))
		if candidate == source {
			continue
		}
		if _, planned := taken[candidate]; planned {
			continue
		}
		dstInfo, err := os.Lstat(candidate)
		if err == nil {
			if os.SameFile(srcInfo, dstInfo) {
				return candidate, nil
			}
			continue
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		return candidate, nil
	}
}
