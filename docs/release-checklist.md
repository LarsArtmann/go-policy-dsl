# Release Checklist — go-policy-dsl

Every release tag MUST pass this checklist. The quality gate has been red
on tag push for all three prior releases (v0.1.0–v0.3.0); this checklist
exists to prevent that from recurring.

---

## Pre-tag

- [ ] **Quality gate green locally:**
  ```bash
  GOWORK=off go test ./...
  GOWORK=off golangci-lint run ./...
  GOWORK=off golangci-lint fmt ./...
  GOWORK=off go build ./...
  GOWORK=off go vet ./...
  ```
- [ ] **Working tree clean:** `git status --short` shows no uncommitted changes.
- [ ] **CHANGELOG.md updated:** the `[Unreleased]` section has been renamed to `[x.y.z] - YYYY-MM-DD` with all changes documented.
- [ ] **FEATURES.md consistent:** feature statuses match what the release ships (no FULLY_FUNCTIONAL claims for untested features).
- [ ] **README.md version table updated:** new version row added if user-visible.
- [ ] **ROADMAP.md accurate:** no shipped items listed as future work.
- [ ] **Breaking changes marked:** every breaking change in CHANGELOG has a `**BREAKING**` marker under `Changed (breaking)`.

## Tag

- [ ] **Annotated tag:** `git tag -a vx.y.z -m "vx.y.z"` (NOT a lightweight tag).
- [ ] **Push tag:** `git push origin vx.y.z`.
- [ ] **Verify on remote:** `git ls-remote --tags origin | grep vx.y.z`.

## Post-tag

- [ ] **CI green on the tag:** check GitHub Actions — the tag push triggers CI; wait for green before proceeding.
- [ ] **proxy.golang.org indexed:** the new version appears on `https://proxy.golang.org/github.com/larsartmann/go-policy-dsl/@v/list`.
- [ ] **pkg.go.dev updated:** the README and API docs reflect the new tag (may take a few minutes; force a re-fetch if stale).
- [ ] **GitHub Release created:** body copied from the CHANGELOG section for this version.
- [ ] **Latest flag set:** the release is marked as "Latest" on GitHub.
- [ ] **[Unreleased] section recreated** in CHANGELOG.md with an empty body for the next cycle.

## Recovery (if CI is red on tag)

1. **Do NOT delete the tag** — tags are immutable history. Deleting and re-tagging corrupts the module proxy cache for anyone who already fetched it.
2. **Fix forward:** create a patch release (`vx.y.z+1`) that makes CI green.
3. **Document the broken tag** in the patch release's CHANGELOG: "Note: vx.y.z had CI red on tag; fixed in vx.y.z+1."
