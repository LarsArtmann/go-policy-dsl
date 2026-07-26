package policydsl_test

import (
	"testing"

	policydsl "github.com/larsartmann/go-policy-dsl"
)

// FuzzBuilder_PatternsOpaque pins the contract that detection patterns are
// OPAQUE STRINGS: the DSL stores them verbatim, in declaration order, and never
// interprets them. The DSL deliberately owns NO matching semantics (literal /
// glob / regex is the consumer's job — see docs/DOMAIN_LANGUAGE.md "Detection").
// Therefore the only property the DSL can be fuzzed for is "any string, however
// weird, round-trips unchanged through every pattern-handling entry point".
//
// This resolves TODO [T1]: a pattern-matching fuzz target cannot exist in the
// DSL (there is nothing to match against), but the opaque-storage contract CAN
// and MUST be pinned so consumers can rely on it.
func FuzzBuilder_PatternsOpaque(f *testing.F) {
	// Seed corpus: the interesting classes of string a consumer might declare —
	// a normal import path, regex metacharacters, empty, control/unicode bytes,
	// and a value that looks like another domain concept.
	seeds := []string{
		"gorm.io/gorm",
		".*",
		"[a-z]+",
		"",
		"unicode \x00\n\t weird",
		"CVE-2021-44228",
		"github.com/Azure/go-workflow",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, pattern string) {
		// Every append-style detection helper must store the pattern verbatim,
		// exactly once, in declaration order.
		spec := policydsl.Ban("x").
			ImportPatterns(pattern).
			GoModPatterns(pattern).
			ExcludeIfContains(pattern).
			ExcludeIfTransitiveFrom(pattern).
			Spec()

		assertStoredOnce := func(label string, got []string) {
			t.Helper()

			if len(got) != 1 || got[0] != pattern {
				t.Errorf("%s: expected [%q] stored verbatim once, got %v", label, pattern, got)
			}
		}

		assertStoredOnce("ImportPatterns", spec.Detection.ImportPatterns)
		assertStoredOnce("GoModPatterns", spec.Detection.GoModPatterns)
		assertStoredOnce("ExcludeIfContains", spec.Detection.ExcludeIfContains)
		assertStoredOnce("ExcludeIfTransitiveFrom", spec.Detection.ExcludeIfTransitiveFrom)

		// The single-pattern convenience constructors must round-trip too.
		importDetection := policydsl.ImportPattern(pattern)
		assertStoredOnce("ImportPattern", importDetection.ImportPatterns)

		if len(importDetection.GoModPatterns) != 0 {
			t.Errorf("ImportPattern should not set GoModPatterns, got %v", importDetection.GoModPatterns)
		}

		goModDetection := policydsl.GoModPattern(pattern)
		assertStoredOnce("GoModPattern", goModDetection.GoModPatterns)

		if len(goModDetection.ImportPatterns) != 0 {
			t.Errorf("GoModPattern should not set ImportPatterns, got %v", goModDetection.ImportPatterns)
		}
	})
}
