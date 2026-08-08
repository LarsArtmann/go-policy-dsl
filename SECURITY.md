# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in `go-policy-dsl`, please report it
responsibly:

1. **Do NOT open a public GitHub issue.**
2. Email **git@lars.software** with a description of the vulnerability and,
   if possible, a proof of concept.
3. You will receive an acknowledgement within 48 hours.

## Scope

This library is a declarative DSL for policy definitions — it has no execution
model, no network access, and no file I/O. The attack surface is limited to:

- Input validation on `NewCVE`, `NewVersion`, `ParseVersion` (all return
  errors, never panic).
- The `Validate()` method on `PolicySpec`.

The library is stdlib-only with zero dependencies, so supply-chain risk is
limited to the Go standard library itself.
