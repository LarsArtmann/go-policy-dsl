## Summary

<!-- One or two sentences: what does this PR change and why? -->

## Change type

- [ ] Bug fix
- [ ] New feature / builder method / policy type
- [ ] **Breaking change** (pre-1.0: allowed; add a `**BREAKING**` line under `CHANGELOG.md` → `Changed`)
- [ ] Documentation
- [ ] Refactor / cleanup
- [ ] Test / CI

## Quality gate

All four must pass before requesting review (see `CONTRIBUTING.md`):

- [ ] `go build ./...`
- [ ] `go test ./... -race`
- [ ] `go vet ./...`
- [ ] `golangci-lint run ./...` (0 issues)
- [ ] `golangci-lint fmt ./...` (the LAST command — no uncommitted fmt drift)

## Stdlib-only contract

- [ ] Library code (`*.go` excluding `_test.go`) imports **only the standard library**.
- [ ] If test deps were added, this was discussed first (the suite is stdlib-only by choice).

## Documentation

- [ ] Public API changes are reflected in `CHANGELOG.md` `[Unreleased]`.
- [ ] `docs/DOMAIN_LANGUAGE.md` / `FEATURES.md` updated if the domain vocabulary or feature status changed.
- [ ] `AGENTS.md` updated if a surprising behaviour or convention was added/changed.

## Tests

- [ ] New behaviour is covered by a test.
- [ ] Surprising behaviours are pinned by a contract test (so a silent change is caught).
