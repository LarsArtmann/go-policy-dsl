package policydsl_test

import (
	"testing"

	policydsl "github.com/larsartmann/go-policy-dsl"
)

// parseVersionOrFatal parses s, failing the test immediately on a parse error.
// It is a test-only convenience (using t.Fatal, not panic) for fixtures where
// the literal is known valid; it deliberately is NOT part of the exported API.
func parseVersionOrFatal(t *testing.T, s string) policydsl.Version {
	t.Helper()

	v, err := policydsl.ParseVersion(s)
	if err != nil {
		t.Fatalf("unexpected ParseVersion(%q) error: %v", s, err)
	}

	return v
}
