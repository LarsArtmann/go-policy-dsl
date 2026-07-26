package policydsl

import (
	"errors"
	"fmt"
	"regexp"
)

// CVE is a validated CVE identifier in the canonical MITRE form
// "CVE-YYYY-NNNN..." (e.g. "CVE-2021-44228"). It is a branded string type so
// the DSL stays dependency-free and JSON-serializable while making an
// unvalidated free-form string unrepresentable on a PolicySpec. Construct via
// NewCVE (validated) or MustCVE (panics) so an invalid ID cannot reach a spec.
type CVE string

// ErrInvalidCVE is returned when a string is not a canonical CVE identifier.
var ErrInvalidCVE = errors.New("policydsl: invalid CVE identifier, want \"CVE-YYYY-NNNN\"")

// cvePattern matches the MITRE CVE ID format: "CVE-" + 4-digit year + "-" +
// 4-or-more digit sequence (MITRE allows the sequence to grow beyond 4 digits).
var cvePattern = regexp.MustCompile(`^CVE-\d{4}-\d{4,}$`)

// NewCVE constructs a validated CVE from a string. Returns ErrInvalidCVE
// (wrapping the offending value) when the string is not canonical
// "CVE-YYYY-NNNN" (e.g. rejects "CVE-21-1", "cve-2021-44228", "2021-44228").
func NewCVE(id string) (CVE, error) {
	if !cvePattern.MatchString(id) {
		return "", fmt.Errorf("%w: %q", ErrInvalidCVE, id)
	}

	return CVE(id), nil
}

// MustCVE is the convenience form of NewCVE for package-level policy
// initialization; it panics on an invalid CVE identifier.
func MustCVE(id string) CVE {
	c, err := NewCVE(id)
	if err != nil {
		panic(err)
	}

	return c
}

// String renders the CVE in its canonical form.
func (c CVE) String() string { return string(c) }
