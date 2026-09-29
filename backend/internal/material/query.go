package material

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Bounds for Find: minimum term length, default limit, and the allowed
// limit range. A limit of 0 selects DefaultLimit.
const (
	MinQueryLength = 2
	DefaultLimit   = 20
	MinLimit       = 10
	MaxLimit       = 50
)

// Query describes a find-by-name request. Term is matched as a trimmed,
// case-insensitive substring of the material name.
type Query struct {
	Term  string
	Limit int
}

// Normalize trims the term and applies the default limit.
func (q *Query) Normalize() {
	q.Term = strings.TrimSpace(q.Term)

	if q.Limit == 0 {
		q.Limit = DefaultLimit
	}
}

// Validate normalizes then rejects a blank or too-short term,
// or a limit outside the allowed range.
func (q *Query) Validate() error {
	q.Normalize()

	if q.Term == "" {
		return fmt.Errorf("%w: search term is required", ErrInvalid)
	}

	if strings.ContainsRune(q.Term, 0) {
		return fmt.Errorf("%w: search term contains a NUL byte", ErrInvalid)
	}

	if utf8.RuneCountInString(q.Term) < MinQueryLength {
		return fmt.Errorf(
			"%w: search term is too short, narrow it down",
			ErrInvalid,
		)
	}

	if utf8.RuneCountInString(q.Term) > MaxNameLength {
		return fmt.Errorf(
			"%w: search term exceeds %d characters",
			ErrInvalid,
			MaxNameLength,
		)
	}

	if q.Limit < MinLimit || q.Limit > MaxLimit {
		return fmt.Errorf(
			"%w: limit %d out of range [%d, %d]",
			ErrInvalid,
			q.Limit,
			MinLimit,
			MaxLimit,
		)
	}

	return nil
}

// EscapeTerm quotes LIKE wildcards with a backslash so the term matches
// literally (matches the query's ESCAPE clause).
func (q Query) EscapeTerm() string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

	return replacer.Replace(q.Term)
}
