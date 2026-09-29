package material

import (
	"errors"
	"testing"
)

func Test_Query_Normalize(t *testing.T) {
	t.Run("trims the term and applies the default limit", func(t *testing.T) {
		q := Query{Term: "  bopp "}

		q.Normalize()

		if q.Term != "bopp" {
			t.Errorf("Term = %q, want %q", q.Term, "bopp")
		}

		if q.Limit != DefaultLimit {
			t.Errorf("Limit = %d, want default %d", q.Limit, DefaultLimit)
		}
	})

	t.Run("keeps an explicit limit", func(t *testing.T) {
		q := Query{Term: "bopp", Limit: 10}

		q.Normalize()

		if q.Limit != 10 {
			t.Errorf("Limit = %d, want 10", q.Limit)
		}
	})
}

func Test_Query_Validate(t *testing.T) {
	t.Run("accepts a query", func(t *testing.T) {
		q := Query{Term: "bopp"}

		if err := q.Validate(); err != nil {
			t.Fatalf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("normalizes while validating", func(t *testing.T) {
		q := Query{Term: "  bopp "}

		if err := q.Validate(); err != nil {
			t.Fatalf("Validate() error = %v, want nil", err)
		}

		if q.Term != "bopp" || q.Limit != DefaultLimit {
			t.Errorf("Validate() left %+v, want trimmed term and default limit", q)
		}
	})

	t.Run("rejects a blank term", func(t *testing.T) {
		for _, term := range []string{"", "   "} {
			q := Query{Term: term, Limit: DefaultLimit}

			if err := q.Validate(); !errors.Is(err, ErrInvalid) {
				t.Errorf("Validate() term %q error = %v, want ErrInvalid", term, err)
			}
		}
	})

	t.Run("rejects a term below the minimum length", func(t *testing.T) {
		q := Query{Term: "P", Limit: DefaultLimit}

		if err := q.Validate(); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("rejects a limit outside the allowed range", func(t *testing.T) {
		for _, limit := range []int{-1, 1, 9, 51} {
			q := Query{Term: "bopp", Limit: limit}

			if err := q.Validate(); !errors.Is(err, ErrInvalid) {
				t.Errorf("Validate() limit %d error = %v, want ErrInvalid", limit, err)
			}
		}
	})
}

func Test_Query_EscapeTerm(t *testing.T) {
	t.Run("quotes LIKE wildcards", func(t *testing.T) {
		q := Query{Term: `100% cotton_under\lay`}

		if got := q.EscapeTerm(); got != `100\% cotton\_under\\lay` {
			t.Errorf("EscapeTerm() = %q, want %q", got, `100\% cotton\_under\\lay`)
		}
	})

	t.Run("leaves plain text untouched", func(t *testing.T) {
		q := Query{Term: "bopp 20um"}

		if got := q.EscapeTerm(); got != "bopp 20um" {
			t.Errorf("EscapeTerm() = %q, want %q", got, "bopp 20um")
		}
	})
}
