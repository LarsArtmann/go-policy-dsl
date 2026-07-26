package policydsl_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	policydsl "github.com/larsartmann/go-policy-dsl"
)

// External test package (policydsl_test) exercises only the exported API, the
// same surface real consumers use. This keeps the library honest: if a test
// needs an unexported symbol, that is a signal the symbol leaks implementation.

func TestBan_DefaultsToCriticalSecurity(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("gorm").Spec()

	if spec.Name != "gorm" {
		t.Errorf("expected name gorm, got %q", spec.Name)
	}

	if spec.Severity != policydsl.SeverityCritical {
		t.Errorf("expected critical severity, got %q", spec.Severity)
	}

	if spec.Category != policydsl.CategorySecurity {
		t.Errorf("expected security category, got %q", spec.Category)
	}
}

func TestBuilder_FullFluentChain(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("gorm").
		Because("ORMs hide N+1 queries").
		WithSeverity(policydsl.SeverityHigh).
		WithCategory(policydsl.CategoryPerformance).
		DetectVia(policydsl.ImportPattern("gorm.io/gorm")).
		Suggest(policydsl.NewReplacement("sqlc", "type-safe SQL")).
		VersionRangeStrings("1.0.0", "2.0.0").
		Spec()

	if spec.Reason != "ORMs hide N+1 queries" {
		t.Errorf("unexpected reason: %q", spec.Reason)
	}

	if spec.Severity != policydsl.SeverityHigh {
		t.Errorf("expected high, got %q", spec.Severity)
	}

	if spec.Category != policydsl.CategoryPerformance {
		t.Errorf("expected performance, got %q", spec.Category)
	}

	if len(spec.Detection.ImportPatterns) != 1 || spec.Detection.ImportPatterns[0] != "gorm.io/gorm" {
		t.Errorf("unexpected import patterns: %v", spec.Detection.ImportPatterns)
	}

	if len(spec.Alternatives) != 1 || spec.Alternatives[0] != "sqlc" {
		t.Errorf("unexpected alternatives: %v", spec.Alternatives)
	}

	if spec.VersionMin == nil || spec.VersionMin.String() != "1.0.0" ||
		spec.VersionMax == nil || spec.VersionMax.String() != "2.0.0" {
		t.Errorf("unexpected version range: min=%v max=%v", spec.VersionMin, spec.VersionMax)
	}
}

// TestBuilder_DetectVia_Replaces confirms DetectVia overwrites any prior
// detection state (set semantics), unlike the bare append helpers.
func TestBuilder_DetectVia_Replaces(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("x").
		ImportPatterns("a").
		DetectVia(policydsl.Detection{
			ImportPatterns: []string{"b"},
			GoModPatterns:  []string{"mod"},
		}).
		Spec()

	if len(spec.Detection.ImportPatterns) != 1 || spec.Detection.ImportPatterns[0] != "b" {
		t.Errorf("DetectVia should replace, got %v", spec.Detection.ImportPatterns)
	}

	if len(spec.Detection.GoModPatterns) != 1 || spec.Detection.GoModPatterns[0] != "mod" {
		t.Errorf("unexpected go.mod patterns: %v", spec.Detection.GoModPatterns)
	}
}

// TestBuilder_AppendDetectionHelpers verifies the four append-style detection
// helpers accumulate values across calls rather than replacing.
func TestBuilder_AppendDetectionHelpers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call func(*policydsl.Builder) *policydsl.Builder
		got  func(policydsl.Detection) []string
		want []string
	}{
		{
			name: "ImportPatterns",
			call: func(b *policydsl.Builder) *policydsl.Builder { return b.ImportPatterns("a", "b").ImportPatterns("c") },
			got:  func(d policydsl.Detection) []string { return d.ImportPatterns },
			want: []string{"a", "b", "c"},
		},
		{
			name: "GoModPatterns",
			call: func(b *policydsl.Builder) *policydsl.Builder { return b.GoModPatterns("m1", "m2") },
			got:  func(d policydsl.Detection) []string { return d.GoModPatterns },
			want: []string{"m1", "m2"},
		},
		{
			name: "ExcludeIfContains",
			call: func(b *policydsl.Builder) *policydsl.Builder { return b.ExcludeIfContains("nolint", "justified") },
			got:  func(d policydsl.Detection) []string { return d.ExcludeIfContains },
			want: []string{"nolint", "justified"},
		},
		{
			name: "ExcludeIfTransitiveFrom",
			call: func(b *policydsl.Builder) *policydsl.Builder { return b.ExcludeIfTransitiveFrom("ginkgo") },
			got:  func(d policydsl.Detection) []string { return d.ExcludeIfTransitiveFrom },
			want: []string{"ginkgo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.got(tt.call(policydsl.Ban("x")).Spec().Detection)
			if !equalStrings(got, tt.want) {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

// TestBuilder_Suggest_SetsDescription documents the surprising side effect:
// Suggest appends the library to Alternatives AND, when Description is empty,
// derives Description from the replacement. Intentional and documented in
// AGENTS.md; this test pins it so a silent change is caught.
func TestBuilder_Suggest_SetsDescription(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("gorm").
		Suggest(policydsl.NewReplacement("sqlc", "type-safe SQL")).
		Spec()

	if len(spec.Alternatives) != 1 || spec.Alternatives[0] != "sqlc" {
		t.Fatalf("unexpected alternatives: %v", spec.Alternatives)
	}

	const want = "Replace with sqlc: type-safe SQL"
	if spec.Description != want {
		t.Errorf("Suggest should derive Description when empty:\nwant: %q\ngot:  %q", want, spec.Description)
	}
}

// An explicit Description is never overwritten by a later Suggest.
func TestBuilder_Suggest_PreservesExplicitDescription(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("gorm").
		WithDescription("custom desc").
		Suggest(policydsl.NewReplacement("sqlc", "type-safe SQL")).
		Spec()

	if spec.Description != "custom desc" {
		t.Errorf("expected explicit description preserved, got %q", spec.Description)
	}
}

// Multiple Suggest calls append to Alternatives; only the first derives
// Description (subsequent calls see a non-empty Description and skip).
func TestBuilder_Suggest_MultipleAppendsAlternatives(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("x").
		Suggest(policydsl.NewReplacement("a", "ra")).
		Suggest(policydsl.NewReplacement("b", "rb")).
		Spec()

	if !equalStrings(spec.Alternatives, []string{"a", "b"}) {
		t.Errorf("expected [a b], got %v", spec.Alternatives)
	}

	if want := "Replace with a: ra"; spec.Description != want {
		t.Errorf("expected first suggestion's derived description %q, got %q", want, spec.Description)
	}
}

// WithAlternatives replaces the slice wholesale (set semantics), unlike Suggest
// which appends. The With- prefix denotes set/replace; bare methods append.
func TestBuilder_WithAlternatives_Replaces(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("x").
		Suggest(policydsl.NewReplacement("a", "ra")).
		WithAlternatives("p", "q").
		Spec()

	if !equalStrings(spec.Alternatives, []string{"p", "q"}) {
		t.Errorf("WithAlternatives should replace, got %v", spec.Alternatives)
	}
}

func TestBuilder_WithCVEs(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("log4j").WithCVEs("CVE-2021-44228", "CVE-2021-45046").Spec()

	if !equalStrings(spec.CVEs, []string{"CVE-2021-44228", "CVE-2021-45046"}) {
		t.Errorf("unexpected CVEs: %v", spec.CVEs)
	}
}

func TestBuilder_VersionRangeStrings(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("golog").VersionRangeStrings("", "1.99.0").Spec()

	if spec.VersionMin != nil {
		t.Errorf("expected nil min (unbounded), got %v", spec.VersionMin)
	}

	if spec.VersionMax == nil || spec.VersionMax.String() != "1.99.0" {
		t.Errorf("expected max 1.99.0, got %v", spec.VersionMax)
	}
}

// TestBuilder_VersionRangeStrings_InvertedPanics confirms the string
// convenience panics on a nonsensical inverted range (min > max). This is the
// footgun the typed Version domain exists to eliminate.
func TestBuilder_VersionRangeStrings_InvertedPanics(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic on inverted VersionRangeStrings, got none")
		}

		msg := fmt.Sprintf("%v", r)
		if !strings.Contains(msg, "inverted") {
			t.Errorf("expected panic message to mention inversion, got %q", msg)
		}
	}()

	policydsl.Ban("x").VersionRangeStrings("2.0.0", "1.0.0").Spec()
}

// TestBuilder_VersionRange_Typed exercises the typed *Version signature:
// nil bounds are stored as nil; typed bounds are stored verbatim; an inverted
// typed range panics.
func TestBuilder_VersionRange_Typed(t *testing.T) {
	t.Parallel()

	minVer := policydsl.MustParseVersion("1.0.0")
	maxVer := policydsl.MustParseVersion("2.0.0")

	spec := policydsl.Ban("x").VersionRange(&minVer, &maxVer).Spec()

	if spec.VersionMin == nil || !spec.VersionMin.Equal(minVer) {
		t.Errorf("expected min %s, got %v", minVer, spec.VersionMin)
	}

	if spec.VersionMax == nil || !spec.VersionMax.Equal(maxVer) {
		t.Errorf("expected max %s, got %v", maxVer, spec.VersionMax)
	}

	// Both nil = fully unbounded.
	unbounded := policydsl.Ban("y").VersionRange(nil, nil).Spec()
	if unbounded.VersionMin != nil || unbounded.VersionMax != nil {
		t.Errorf("expected nil bounds, got min=%v max=%v", unbounded.VersionMin, unbounded.VersionMax)
	}
}

func TestBuilder_VersionRange_Typed_InvertedPanics(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on inverted typed VersionRange, got none")
		}
	}()

	high := policydsl.MustParseVersion("2.0.0")
	low := policydsl.MustParseVersion("1.0.0")
	policydsl.Ban("x").VersionRange(&high, &low).Spec()
}

// TestPolicySpec_Validate confirms Validate catches an inverted range that
// direct field assignment can introduce (bypassing the Builder guard).
func TestPolicySpec_Validate(t *testing.T) {
	t.Parallel()

	t.Run("sound_spec_validates", func(t *testing.T) {
		t.Parallel()

		spec := policydsl.Ban("x").VersionRangeStrings("1.0.0", "2.0.0").Spec()
		if err := spec.Validate(); err != nil {
			t.Errorf("expected nil error for sound spec, got %v", err)
		}
	})

	t.Run("inverted_range_returns_error", func(t *testing.T) {
		t.Parallel()

		high := policydsl.MustParseVersion("2.0.0")
		low := policydsl.MustParseVersion("1.0.0")
		spec := policydsl.PolicySpec{
			VersionMin: &high,
			VersionMax: &low,
		}

		err := spec.Validate()
		if err == nil {
			t.Fatalf("expected ErrInvertedVersionRange, got nil")
		}

		if !errors.Is(err, policydsl.ErrInvertedVersionRange) {
			t.Errorf("expected ErrInvertedVersionRange, got %v", err)
		}
	})

	t.Run("nil_bounds_valid", func(t *testing.T) {
		t.Parallel()

		if err := (policydsl.PolicySpec{}).Validate(); err != nil {
			t.Errorf("zero-value spec should validate, got %v", err)
		}
	})
}

func TestBuilder_RequiresCompanionAndAsCompanionOnly(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("samber/do").
		Because("DI container").
		RequiresCompanion(policydsl.Companion("samber-do-auditlog", "auditlog", "audit trail")).
		AsCompanionOnly().
		Spec()

	if !spec.CompanionOnly {
		t.Error("expected companion-only")
	}

	if len(spec.Companions) != 1 {
		t.Fatalf("expected 1 companion, got %d", len(spec.Companions))
	}

	companion := spec.Companions[0]
	if companion.Library != "samber-do-auditlog" {
		t.Errorf("unexpected companion library: %q", companion.Library)
	}

	if companion.DetectionPattern != "auditlog" {
		t.Errorf("unexpected pattern: %q", companion.DetectionPattern)
	}
}

func TestCompanion_DefaultSeverity(t *testing.T) {
	t.Parallel()

	c := policydsl.Companion("lib", "pat", "why")
	if c.Severity != policydsl.SeverityModerate {
		t.Errorf("expected moderate, got %q", c.Severity)
	}
}

func TestCompanionWithSeverity_Overrides(t *testing.T) {
	t.Parallel()

	c := policydsl.CompanionWithSeverity("lib", "pat", "why", policydsl.SeverityCritical)
	if c.Severity != policydsl.SeverityCritical {
		t.Errorf("expected critical, got %q", c.Severity)
	}
}

func TestImportPattern(t *testing.T) {
	t.Parallel()

	detection := policydsl.ImportPattern("fmt")
	if len(detection.ImportPatterns) != 1 || detection.ImportPatterns[0] != "fmt" {
		t.Errorf("unexpected: %v", detection)
	}

	if len(detection.GoModPatterns) != 0 {
		t.Errorf("expected no go.mod patterns, got %v", detection.GoModPatterns)
	}
}

func TestGoModPattern(t *testing.T) {
	t.Parallel()

	detection := policydsl.GoModPattern("github.com/foo/bar")
	if len(detection.GoModPatterns) != 1 || detection.GoModPatterns[0] != "github.com/foo/bar" {
		t.Errorf("unexpected: %v", detection)
	}

	if len(detection.ImportPatterns) != 0 {
		t.Errorf("expected no import patterns, got %v", detection.ImportPatterns)
	}
}

func TestNewReplacement(t *testing.T) {
	t.Parallel()

	replacement := policydsl.NewReplacement("github.com/x/y", "why")
	if replacement.Library != "github.com/x/y" || replacement.Reason != "why" {
		t.Errorf("unexpected replacement: %+v", replacement)
	}
}

// TestBuilder_DetectVia_GoModPattern_Composition pins the common composition
// Ban(...).DetectVia(GoModPattern(...)) — the convenience constructor feeds
// the replace-semantics setter and only the go.mod pattern survives.
func TestBuilder_DetectVia_GoModPattern_Composition(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("foo").
		ImportPatterns("should-be-overwritten").
		DetectVia(policydsl.GoModPattern("github.com/foo/bar")).
		Spec()

	if len(spec.Detection.GoModPatterns) != 1 || spec.Detection.GoModPatterns[0] != "github.com/foo/bar" {
		t.Errorf("expected single go.mod pattern github.com/foo/bar, got %v", spec.Detection.GoModPatterns)
	}

	// DetectVia replaces the whole Detection struct, so the prior ImportPatterns
	// call is discarded.
	if len(spec.Detection.ImportPatterns) != 0 {
		t.Errorf("DetectVia should have replaced prior ImportPatterns, got %v", spec.Detection.ImportPatterns)
	}
}

// TestSeverity_ConstantValues pins the string value of every Severity
// constant. These values are the wire format consumers bridge against, so an
// accidental rename or value change MUST be caught here.
func TestSeverity_ConstantValues(t *testing.T) {
	t.Parallel()

	want := map[policydsl.Severity]string{
		policydsl.SeverityCritical: "critical",
		policydsl.SeverityHigh:     "high",
		policydsl.SeverityModerate: "moderate",
		policydsl.SeverityLow:      "low",
		policydsl.SeverityInfo:     "info",
	}

	for severity, expected := range want {
		if string(severity) != expected {
			t.Errorf("Severity constant %q = %q, want %q", severity, severity, expected)
		}
	}
}

// TestCategory_ConstantValues pins the string value of every Category
// constant. Same rationale as TestSeverity_ConstantValues.
func TestCategory_ConstantValues(t *testing.T) {
	t.Parallel()

	want := map[policydsl.Category]string{
		policydsl.CategorySecurity:        "security",
		policydsl.CategoryPerformance:     "performance",
		policydsl.CategoryMaintainability: "maintainability",
		policydsl.CategoryCorrectness:     "correctness",
		policydsl.CategoryLicensing:       "licensing",
		policydsl.CategoryCompatibility:   "compatibility",
		policydsl.CategoryConfiguration:   "configuration",
	}

	for category, expected := range want {
		if string(category) != expected {
			t.Errorf("Category constant %q = %q, want %q", category, category, expected)
		}
	}
}

// equalStrings is a tiny stdlib-only slice comparison (test-only) so the suite
// stays dependency-free, matching the library's zero-dep contract.
func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}

	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}

	return true
}
