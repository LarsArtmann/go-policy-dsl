package policydsl_test

// This file is the behaviour suite for the fluent builder. It is deliberately
// written in the stdlib `testing` package — NOT Ginkgo/Gomega — because the
// library's headline contract is "stdlib-only, zero dependencies" and a
// `require github.com/onsi/ginkgo/v2` entry in go.mod would contradict that
// claim even though depguard exempts _test.go. The structure (one
// TestBehavior_* function per behaviour group, with descriptive t.Run names)
// preserves the BDD goal: specs read as sentences top-down and survive
// refactoring because they state what the DSL _does_, not how the setters
// are wired.

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	policydsl "github.com/larsartmann/go-policy-dsl"
)

func TestBehavior_BuildingABan(t *testing.T) {
	t.Parallel()

	t.Run("it reads like prose and preserves every declared intent", func(t *testing.T) {
		t.Parallel()

		spec := policydsl.Ban("gorm").
			Because("ORMs hide N+1 queries; use sqlc").
			WithSeverity(policydsl.SeverityCritical).
			WithCategory(policydsl.CategoryPerformance).
			DetectVia(policydsl.ImportPattern("gorm.io/gorm")).
			Suggest(policydsl.NewReplacement("github.com/yourorg/sqlc-queries", "type-safe SQL")).
			Spec()

		assertSpecField(t, "name", spec.Name, "gorm")
		assertSpecField(t, "reason", spec.Reason, "ORMs hide N+1 queries; use sqlc")
		assertSpecField(t, "severity", string(spec.Severity), string(policydsl.SeverityCritical))
		assertSpecField(t, "category", string(spec.Category), string(policydsl.CategoryPerformance))
		assertSpecField(t, "import pattern", spec.Detection.ImportPatterns[0], "gorm.io/gorm")
		assertSpecField(t, "alternative", spec.Alternatives[0], "github.com/yourorg/sqlc-queries")
	})

	t.Run("it defaults to critical security when no overrides are given", func(t *testing.T) {
		t.Parallel()

		spec := policydsl.Ban("logrus").Spec()

		assertSpecField(t, "severity", string(spec.Severity), string(policydsl.SeverityCritical))
		assertSpecField(t, "category", string(spec.Category), string(policydsl.CategorySecurity))
	})
}

func TestBehavior_SuggestingAReplacement(t *testing.T) {
	t.Parallel()

	t.Run("it documents itself by deriving a description when none is set", func(t *testing.T) {
		t.Parallel()

		spec := policydsl.Ban("gorm").
			Suggest(policydsl.NewReplacement("sqlc", "type-safe SQL")).
			Spec()

		assertSpecField(t, "derived description", spec.Description, "Replace with sqlc: type-safe SQL")
	})

	t.Run("it never overwrites an explicit description", func(t *testing.T) {
		t.Parallel()

		spec := policydsl.Ban("gorm").
			WithDescription("custom: do not use gorm").
			Suggest(policydsl.NewReplacement("sqlc", "type-safe SQL")).
			Spec()

		assertSpecField(t, "description", spec.Description, "custom: do not use gorm")
	})

	t.Run("it accumulates alternatives across repeated suggestions", func(t *testing.T) {
		t.Parallel()

		spec := policydsl.Ban("gorm").
			Suggest(policydsl.NewReplacement("sqlc", "type-safe SQL")).
			Suggest(policydsl.NewReplacement("ent", "typed schema")).
			Spec()

		if len(spec.Alternatives) != 2 {
			t.Fatalf("expected 2 alternatives, got %d: %v", len(spec.Alternatives), spec.Alternatives)
		}

		assertSpecField(t, "alternative[0]", spec.Alternatives[0], "sqlc")
		assertSpecField(t, "alternative[1]", spec.Alternatives[1], "ent")
	})
}

func TestBehavior_RequiringACompanion(t *testing.T) {
	t.Parallel()

	t.Run("it records the companion with its detection pattern and reason", func(t *testing.T) {
		t.Parallel()

		spec := policydsl.Ban("samber/do").
			RequiresCompanion(policydsl.Companion(
				"samber-do-auditlog",
				"github.com/larsartmann/samber-do-auditlog",
				"audit trail for DI lifecycle",
			)).
			Spec()

		if len(spec.Companions) != 1 {
			t.Fatalf("expected 1 companion, got %d", len(spec.Companions))
		}

		companion := spec.Companions[0]
		assertSpecField(t, "companion library", companion.Library, "samber-do-auditlog")
		assertSpecField(t, "companion pattern", companion.DetectionPattern, "github.com/larsartmann/samber-do-auditlog")
		assertSpecField(t, "companion severity", string(companion.Severity), string(policydsl.SeverityModerate))
	})

	t.Run("a companion-only policy sets the mode so the consumer never emits a ban", func(t *testing.T) {
		t.Parallel()

		spec := policydsl.Ban("samber/do").AsCompanionOnly().Spec()

		if spec.Mode != policydsl.ModeCompanionOnly {
			t.Errorf("expected ModeCompanionOnly so the consumer suppresses the ban; got %q", spec.Mode)
		}
	})
}

func TestBehavior_BoundingByLibraryVersion(t *testing.T) {
	t.Parallel()

	t.Run("an unbounded-then-capped range stores nil min and a parsed max", func(t *testing.T) {
		t.Parallel()

		spec := policydsl.Ban("golog").VersionRangeStrings("", "1.99.0").Spec()

		if spec.VersionMin != nil {
			t.Errorf("expected nil (unbounded) min, got %v", spec.VersionMin)
		}

		if spec.VersionMax == nil || spec.VersionMax.String() != "1.99.0" {
			t.Errorf("expected max 1.99.0, got %v", spec.VersionMax)
		}
	})

	t.Run("an inverted range is rejected at construction (fail-fast)", func(t *testing.T) {
		t.Parallel()

		defer expectInversionPanic(t)

		policydsl.Ban("golog").VersionRangeStrings("2.0.0", "1.0.0").Spec()
	})

	t.Run("a spec built via the builder always validates", func(t *testing.T) {
		t.Parallel()

		spec := policydsl.Ban("golog").VersionRangeStrings("1.0.0", "2.0.0").Spec()

		if err := spec.Validate(); err != nil {
			t.Errorf("a builder-produced spec should Validate clean; got %v", err)
		}
	})

	t.Run("direct field assignment that inverts the range fails Validate", func(t *testing.T) {
		t.Parallel()

		high := policydsl.MustParseVersion("2.0.0")
		low := policydsl.MustParseVersion("1.0.0")
		inverted := policydsl.PolicySpec{VersionMin: &high, VersionMax: &low}

		err := inverted.Validate()
		if err == nil {
			t.Fatalf("expected Validate to reject an inverted range, got nil")
		}

		if !errors.Is(err, policydsl.ErrInvertedVersionRange) {
			t.Errorf("expected ErrInvertedVersionRange, got %v", err)
		}
	})
}

// expectInversionPanic is a deferred helper that asserts the surrounding
// function panicked with a message mentioning "inverted". It centralizes the
// recover-and-assert dance so each fail-fast spec reads cleanly.
func expectInversionPanic(t *testing.T) {
	t.Helper()

	recovered := recover()
	if recovered == nil {
		t.Fatalf("expected a panic on inverted VersionRangeStrings, got none")
	}

	message := strings.ToLower(fmt.Sprintf("%v", recovered))
	if !strings.Contains(message, "inverted") {
		t.Errorf("expected panic message to mention inversion, got %q", recovered)
	}
}

// assertSpecField is a tiny stdlib-only equality helper so the behaviour specs
// read as plain English assertions without pulling a matchers library.
func assertSpecField(t *testing.T, label, got, want string) {
	t.Helper()

	if got != want {
		t.Errorf("%s: got %q, want %q", label, got, want)
	}
}
