package policydsl_test

// Example functions are compile-verified by `go test` and rendered on
// pkg.go.dev, so the documented usage can never silently rot. Each Example's
// `// Output:` comment is the contract — if the output drifts, the test fails.

import (
	"errors"
	"fmt"

	policydsl "github.com/larsartmann/go-policy-dsl"
)

// ExampleBan shows the canonical ban: a library banned for a category with a
// suggested replacement. The fluent chain reads top-down as a sentence.
func ExampleBan() {
	spec := policydsl.Ban("gorm").
		Because("ORMs hide N+1 queries; use sqlc for type-safe SQL").
		WithSeverity(policydsl.SeverityCritical).
		WithCategory(policydsl.CategoryPerformance).
		DetectVia(policydsl.ImportPattern("gorm.io/gorm")).
		Suggest(policydsl.NewReplacement("github.com/yourorg/sqlc-queries", "type-safe SQL")).
		Spec()

	fmt.Println(spec.Name, spec.Severity, spec.Category)
	fmt.Println(spec.Alternatives[0])
	// Output:
	// gorm critical performance
	// github.com/yourorg/sqlc-queries
}

// ExampleBan_versionRange shows a version-bounded ban: only library versions
// below 2.x are banned, and the inversion guard means the range is sound.
func ExampleBan_versionRange() {
	spec := policydsl.Ban("golog").
		Because("v1.x has a memory leak fixed in v2").
		VersionRangeStrings("", "1.99.0").
		Spec()

	if spec.VersionMin == nil {
		fmt.Println("min: unbounded")
	}

	fmt.Println("max:", spec.VersionMax)
	// Output:
	// min: unbounded
	// max: 1.99.0
}

// ExampleCompanion shows a companion-only policy: the library is fine, but if
// you use it you must also use the named companion. AsCompanionOnly suppresses
// the ban.
func ExampleCompanion() {
	spec := policydsl.Ban("samber/do").
		Because("DI containers without audit trails are opaque").
		RequiresCompanion(policydsl.Companion(
			"samber-do-auditlog",
			"github.com/larsartmann/samber-do-auditlog",
			"audit trail for DI lifecycle",
		)).
		AsCompanionOnly().
		Spec()

	fmt.Println("companion-only:", spec.Mode == policydsl.ModeCompanionOnly)
	fmt.Println("companion:", spec.Companions[0].Library)
	// Output:
	// companion-only: true
	// companion: samber-do-auditlog
}

// ExampleVersion demonstrates parsing, comparison, and rendering.
func ExampleVersion() {
	v, _ := policydsl.ParseVersion("v1.2.3")
	other := policydsl.MustNewVersion(2, 0, 0)

	fmt.Println(v)
	fmt.Println(v.Before(other))
	fmt.Println(v.Compare(other))
	// Output:
	// 1.2.3
	// true
	// -1
}

// ExamplePolicySpec_Validate shows that an inverted version range — which
// direct field assignment can introduce — is caught by Validate.
func ExamplePolicySpec_Validate() {
	high := policydsl.MustParseVersion("2.0.0")
	low := policydsl.MustParseVersion("1.0.0")
	inverted := policydsl.PolicySpec{VersionMin: &high, VersionMax: &low}

	err := inverted.Validate()
	fmt.Println(errors.Is(err, policydsl.ErrInvertedVersionRange))
	// Output: true
}
