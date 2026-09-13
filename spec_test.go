package policydsl_test

import (
	"testing"

	policydsl "github.com/larsartmann/go-policy-dsl"
)

func TestBuilderWhenAnyPresent(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("github.com/example/lib").
		WhenAnyPresent("github.com/gin-gonic/gin", "github.com/larsartmann/httputil").
		Spec()

	if len(spec.Detection.WhenAnyPresent) != 2 {
		t.Fatalf("expected 2 WhenAnyPresent signals, got %v", spec.Detection.WhenAnyPresent)
	}

	if spec.Detection.WhenAnyPresent[0] != "github.com/gin-gonic/gin" {
		t.Errorf("expected first signal preserved, got %s", spec.Detection.WhenAnyPresent[0])
	}

	if len(spec.Detection.RequireIfContains) != 0 {
		t.Errorf("WhenAnyPresent must not leak into RequireIfContains, got %v", spec.Detection.RequireIfContains)
	}
}

func TestBuilderWhenAnyPresentAppends(t *testing.T) {
	t.Parallel()

	spec := policydsl.Ban("github.com/example/lib").
		WhenAnyPresent("a").
		WhenAnyPresent("b", "c").
		Spec()

	if len(spec.Detection.WhenAnyPresent) != 3 {
		t.Fatalf("expected appended signals a,b,c — got %v", spec.Detection.WhenAnyPresent)
	}
}
