package policydsl

import "testing"

func TestBan_DefaultsToCriticalSecurity(t *testing.T) {
	spec := Ban("gorm").Spec()

	if spec.Name != "gorm" {
		t.Errorf("expected name gorm, got %s", spec.Name)
	}

	if spec.Severity != SeverityCritical {
		t.Errorf("expected critical severity, got %s", spec.Severity)
	}

	if spec.Category != CategorySecurity {
		t.Errorf("expected security category, got %s", spec.Category)
	}
}

func TestBuilder_FluentChain(t *testing.T) {
	spec := Ban("gorm").
		Because("ORMs hide N+1 queries").
		WithSeverity(SeverityHigh).
		WithCategory(CategoryPerformance).
		DetectVia(ImportPattern("gorm.io/gorm")).
		Suggest(NewReplacement("sqlc", "type-safe SQL")).
		Spec()

	if spec.Reason != "ORMs hide N+1 queries" {
		t.Errorf("unexpected reason: %s", spec.Reason)
	}

	if spec.Severity != SeverityHigh {
		t.Errorf("expected high, got %s", spec.Severity)
	}

	if spec.Category != CategoryPerformance {
		t.Errorf("expected performance, got %s", spec.Category)
	}

	if len(spec.Detection.ImportPatterns) != 1 || spec.Detection.ImportPatterns[0] != "gorm.io/gorm" {
		t.Errorf("unexpected import patterns: %v", spec.Detection.ImportPatterns)
	}

	if len(spec.Alternatives) != 1 || spec.Alternatives[0] != "sqlc" {
		t.Errorf("unexpected alternatives: %v", spec.Alternatives)
	}
}

func TestBuilder_RequiresCompanion(t *testing.T) {
	spec := Ban("samber/do").
		Because("DI container").
		RequiresCompanion(Companion("samber-do-auditlog", "auditlog", "audit trail")).
		AsCompanionOnly().
		Spec()

	if !spec.CompanionOnly {
		t.Error("expected companion-only")
	}

	if len(spec.Companions) != 1 {
		t.Fatalf("expected 1 companion, got %d", len(spec.Companions))
	}

	if spec.Companions[0].Library != "samber-do-auditlog" {
		t.Errorf("unexpected companion: %s", spec.Companions[0].Library)
	}
}

func TestImportPattern(t *testing.T) {
	d := ImportPattern("fmt")
	if len(d.ImportPatterns) != 1 || d.ImportPatterns[0] != "fmt" {
		t.Errorf("unexpected: %v", d)
	}
}

func TestCompanion_DefaultSeverity(t *testing.T) {
	c := Companion("lib", "pat", "why")
	if c.Severity != SeverityModerate {
		t.Errorf("expected moderate, got %s", c.Severity)
	}
}

func TestExcludeIfTransitiveFrom(t *testing.T) {
	spec := Ban("testify").
		ExcludeIfTransitiveFrom("ginkgo").
		Spec()

	if len(spec.Detection.ExcludeIfTransitiveFrom) != 1 {
		t.Fatalf("expected 1 transitive exclusion")
	}
}
