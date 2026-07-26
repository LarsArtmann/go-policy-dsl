# Contributing

Thanks for your interest in contributing to `go-policy-dsl`!

## How to Contribute

1. Fork the repository.
2. Create a feature branch (`git switch -c my-feature`).
3. Make your changes (see the conventions below).
4. Run the full quality gate locally.
5. Submit a pull request.

## Development Setup

The project is a **stdlib-only, zero-dependency** public SDK library. There is
no `flake.nix` (intentional — buildflow handles CI for this tiny module); all
you need is Go 1.26+ and `golangci-lint`.

## The Quality Gate (run all four before pushing)

```bash
go build ./...                 # build check
go test ./... -race            # tests + race detector
go vet ./...                   # vet
golangci-lint run ./...        # lint (uses .golangci.yml, v2 format)
golangci-lint fmt ./...        # apply formatters (gci, goimports, gofumpt, golines@120)
```

`golangci-lint fmt` is the **last** command before declaring done — formatter
drift must never land. If it changes anything, re-review and re-run.

## Conventions

### Stdlib-only contract

The library (`*.go` excluding `_test.go`) must import **only the standard
library**. `depguard` enforces this in `.golangci.yml`. Test files are exempt
(`_test.go` is excluded from depguard) but, by deliberate project choice, the
test suite is ALSO stdlib-only so `go.mod` has zero `require` entries — a
reader of `go.mod` should see the "zero dependencies" claim confirmed
directly. Do not add Ginkgo, testify, or any other test dependency without
discussing it first.

### External test package

Tests live in `package policydsl_test` (the external package), not
`package policydsl`. This exercises only the exported API — the same surface
real consumers use — and keeps the library honest: if a test needs an
unexported symbol, that is a signal the symbol leaks implementation.

### Builder method naming

The `Builder` fluent API follows a naming convention (documented at the
`Builder` type and in `docs/DOMAIN_LANGUAGE.md`):

- `With<X>(...)` — **set / replace** the field wholesale.
- Bare `<X>(...)` — **append** to a slice field.
- `DetectVia(d)` — **replace** the whole `Detection` struct.
- `As<X>()` — **set a mode flag**.
- `Spec()` — **terminate** the chain.

When adding a new method, pick the form that matches its semantics.

### Surprising behaviours (do not break)

These are intentional and pinned by tests; see `AGENTS.md` for full detail:

- `Suggest(r Replacement)` derives `Description` from the replacement when
  `Description` is empty (and never overwrites an explicit one).
- `VersionRange(min, max *Version)` panics on an inverted range (`min > max`).
- `Spec()` performs no validation; structural validation lives in the separate
  `PolicySpec.Validate()` method.

## Reporting Issues

Please use GitHub Issues to report bugs or request features.
