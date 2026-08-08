# Status Report — 2026-08-08 07:07

## Session scope

Continuation of the public-launch effort. This session picked up the improvement items from the prior status report and executed them: security audit of historical docs, repo topics, CONTRIBUTING.md fixes, README GOWORK=off note, SECURITY.md, CI optimization, and pkg.go.dev verification. This report covers only what this session did and noticed.

---

## a) FULLY DONE

### 1. Historical docs security audit

Scanned all files under `docs/status/` and `docs/reviews/` for secrets, credentials, API keys, tokens, private keys, internal URLs, and IP addresses.

**Result: clean.** No secrets, no credentials, no internal infrastructure references. The only token matches were `go/token` (Go stdlib import) and CSS design tokens in HTML reports. Historical docs reference `library-policy` and `go-linter-sdk` extensively — these are private repos — but the references are factual ("this is the primary consumer") and already present in AGENTS.md and ROADMAP.md. Acceptable for a public SDK repo.

### 2. GitHub repo topics set

Added 7 topics: `go`, `golang`, `policy`, `linting`, `dsl`, `library-governance`, `code-quality`. Verified via `gh api repos/.../topics`.

### 3. CONTRIBUTING.md fixes

Two changes:
- **Fixed a documentation lie:** Line 70 said `VersionRange(min, max *Version) panics on an inverted range (min > max)` — this was removed in v0.2.0 (panic-free design). Replaced with the correct behavior: does NOT panic, detect via `PolicySpec.Validate()`.
- **Added GOWORK=off note:** Contributors with a parent `go.work` need `GOWORK=off` for all Go commands.

Committed by auto-git daemon as `dfade8f`.

### 4. README GOWORK=off note

Added a blockquote under the contributing commands section: "If you have a parent `go.work` that doesn't include this module, prefix commands with `GOWORK=off`." Committed by auto-git daemon as `dfade8f`.

### 5. SECURITY.md created

Responsible disclosure policy for the public repo: email `git@lars.software`, do NOT open public issues, 48-hour acknowledgement SLA. Scoped the attack surface accurately (input validation on constructors, no network/file I/O, stdlib-only). Committed by auto-git daemon as `dfade8f`.

### 6. CI optimized

Replaced `go install golangci-lint` (compiles from source, ~1m overhead) with `golangci/golangci-lint-action@v8` (pre-built binary with built-in caching).

**Quality gate job: 37s (down from 1m41s)** — 2.7x faster.

Also added:
- `concurrency: { group: ci-${{ github.ref }}, cancel-in-progress: true }` — cancels stale runs when new pushes land.
- `timeout-minutes: 10` (quality gate) / `timeout-minutes: 15` (fuzz) — prevents hung jobs.

Committed as `09a8ff3`. CI verified green — both jobs pass.

### 7. pkg.go.dev verified — fully indexed

pkg.go.dev now resolves with the full module page:
- Version: v0.3.0
- License: MIT
- Full API documentation rendered (all types, methods, examples)
- README rendered (with the old version that still references `library-policy` — see below)
- `RequireIfContains` marked `added in v0.3.0`
- 6 imports shown, 0 importers (no consumers yet)

### 8. Quality gate passes locally

`GOWORK=off go test ./...` (0.004s), `GOWORK=off golangci-lint run ./...` (0 issues), `GOWORK=off go build ./...` (clean).

---

## b) PARTIALLY DONE

### pkg.go.dev README is stale

pkg.go.dev rendered the **old** README — the one from commit `f2d4caf` (v0.3.0 tag), not the current rewritten public-facing README from `cb1e2b9`. The old README still references `library-policy` as a consumer and has a different structure. This will self-correct when pkg.go.dev re-indexes (triggered by a new tag or manual request), but right now a public visitor to pkg.go.dev sees the old README. The GitHub repo itself shows the correct README.

---

## c) NOT STARTED

### Branch protection

Not configured. The auto-git daemon commits directly to master, and there are no required status checks. A public repo should have at least CI-required status checks on master. This conflicts with the auto-git workflow and needs a user decision.

---

## d) TOTALLY FUCKED UP

### CONTRIBUTING.md contained a documentation lie until this session

Line 70 said `VersionRange(min, max *Version) panics on an inverted range (min > max)` — a behavior that was **removed in v0.2.0**. The library is explicitly panic-free by design. This lie survived across multiple sessions, a CHANGELOG rewrite, an AGENTS.md rewrite, and a README rewrite. Nobody checked CONTRIBUTING.md because it wasn't in the "hot" files list. A new contributor reading this would expect a panic, write code assuming a panic, and be confused when it doesn't happen.

**Fixed this session** by correcting the line to describe the actual panic-free behavior.

---

## e) WHAT WE SHOULD IMPROVE

1. **CONTRIBUTING.md was never checked during the release/docs work.** The CHANGELOG, README, AGENTS.md, and ROADMAP.md were all rewritten — but CONTRIBUTING.md was left with stale, incorrect content. A docs audit should be comprehensive, not selective.

2. **pkg.go.dev serves the tagged README, not HEAD.** The old README with `library-policy` references is publicly visible on pkg.go.dev. This self-corrects on the next tag, but for a freshly public repo, first impressions matter. A manual re-index request via `go get github.com/larsartmann/go-policy-dsl@latest` + pkg.go.dev request form could speed this up.

3. **CI improvement was 3 commits to get right.** First attempt: version bump (failed — checksum mismatch). Second attempt: `go install` (worked but slow). Third attempt: `golangci-lint-action` (fast, cached). The right answer should have been the first answer — the action is the standard approach recommended by golangci-lint docs.

4. **No branch protection on a public repo.** Any contributor can push directly to master. The auto-git daemon does this routinely. At minimum, CI-required status checks should be enforced.

5. **The Node.js 20 deprecation warning in CI.** `golangci-lint-action@v8` targets Node.js 20 which is deprecated on GitHub Actions runners (forced to Node.js 24). This is a warning, not an error, but will eventually break. The action maintainer needs to update; we can't fix it locally.

---

## f) Up to 50 things to get done next

### Immediate (noticed this session, not yet done)
1. Trigger pkg.go.dev re-index to serve the current README (not the v0.3.0 tagged version)
2. Set up branch protection on master (require CI green before merge) — needs user decision re: auto-git daemon
3. Run `erraudit ./...` to confirm 0 violations (mentioned in AGENTS.md, not run this session)
4. Check if `WithAlternativeStrings` should be documented in the README API table (it exists in the code but is missing from the README constructor table)

### CI / infrastructure
5. Monitor for `golangci-lint-action` Node.js 20 → 24 fix (upstream)
6. Consider adding `govulncheck` job to CI (stdlib-only, but deps-of-deps matter downstream)
7. Consider adding test-coverage reporting (`go test -cover` + threshold gate)
8. Consider adding a Go versions matrix to the quality-gate job (at minimum: 1.26.x)
9. Reconsider `cache: false` on `setup-go` — the golangci-lint-action now handles its own caching, but Go module caching could still help build/vet/test steps
10. Consider adding `actionlint` to CI to catch malformed workflow YAML
11. Consider adding a release GitHub Action (tag → create release from CHANGELOG)

### Documentation consistency
12. Verify `docs/DOMAIN_LANGUAGE.md` doesn't contain stale API references (e.g., `GoVersionRange`, `Must*` functions, `CompanionOnly bool`)
13. Verify no other doc claims `VersionRange` panics (grep found CONTRIBUTING.md this session — check the rest)
14. Do a strict CHANGELOG ↔ release body consistency pass
15. Verify FEATURES.md doesn't claim anything contradicted by the corrected CHANGELOG
16. Check if TODO_LIST.md harvest log `[Unreleased]` references should be `[0.3.0]`
17. Consider adding `WithAlternativeStrings` to the README API table or documenting why it's excluded

### Public presence
18. Verify `go get github.com/larsartmann/go-policy-dsl@v0.3.0` works from a clean environment
19. Consider adding the repo to awesome-go or similar curated lists
20. Consider writing a blog post or announcement for the public launch
21. Consider adding the repo to the LarsArtmann org profile README (if one exists)
22. Cross-link from sibling repos (go-error-family, go-output, etc.)
23. Consider whether a docs website (Astro + Starlight) is worth setting up now
24. Consider adding a code of conduct (`CODE_OF_CONDUCT.md`)

### Code quality
25. Consider whether the `Severity` enum values `recommended`, `deprecated`, `obsolete` are used or over-engineered
26. Consider adding more godoc examples for edge cases (`SuggestExplicit`, `Validate`)
27. Run `golangci-lint fmt ./...` to check for formatting drift after CI version bump
28. Consider whether the fuzz targets need seed corpus expansion

### Release process
29. Write a release checklist doc (verify CI green, CHANGELOG, tag diff, Latest flag)
30. Consider automating release creation from CHANGELOG sections
31. Consider whether v0.3.0 release notes should mention the CI fix (post-release fix)
32. Consider whether a v0.3.1 patch release is warranted for the CONTRIBUTING.md fix and CI fix

### Cleanup
33. Verify all badge links resolve on the public README (Go Reference, CI, License)
34. Consider whether `docs/status/` historical reports should be visible on a public repo
35. Check if any historical reports reference now-fixed issues in a misleading way
36. Consider archiving or annotating old status reports
37. Consider whether the empty-message commit `2c26518` needs documentation
38. Consider adding a version compatibility table (Go version → library version)
39. Consider whether CONTRIBUTING.md needs a section about the release process
40. Consider whether the issue/PR templates are adequate for public consumers

### Consumer readiness
41. Verify `library-policy` can actually `go get` this module now that it's public
42. Prepare a migration plan for `library-policy` to adopt this SDK
43. Consider whether `go-linter-sdk` adoption should be revisited
44. Consider whether the ROADMAP v1.0.0 criteria need updating post-public-launch
45. Consider whether a "used by" section should be added once consumers adopt

### Hardening
46. Add `SECURITY.md` to `.github/` directory (GitHub convention) instead of repo root
47. Consider adding `govulncheck` as a periodic scheduled job
48. Consider adding dependency-review-action for PRs
49. Consider adding SBOM generation on release
50. Consider whether the `dependabot.yml` covers the right ecosystems (currently only `github-actions`)

---

## g) Questions I cannot figure out myself

1. **Should I set up branch protection on master?** The auto-git daemon commits directly to master without PRs. If I enforce "require PR review + CI green before merge," the daemon's commits would be blocked. Options: (a) allow the daemon an admin bypass, (b) make the daemon open PRs instead of pushing directly, (c) only enforce CI-required without PR review, (d) leave it unprotected. Which do you prefer?

2. **Should I trigger a pkg.go.dev re-index now, or wait for the next release tag?** The current pkg.go.dev page shows the old README (references `library-policy`). A re-index via `go get @latest` would update it to the current README. But if you're planning a v0.3.1 soon, that would also trigger a re-index naturally.

3. **Should `WithAlternativeStrings` be in the README API table?** It exists in the code (a string-convenience method), is documented on pkg.go.dev, but is missing from the README constructor table. It might be intentionally omitted (because `WithAlternatives` with `NewReplacement` is the preferred path), or it might be an oversight.

---

## Summary

| Area | State |
|------|-------|
| Historical docs audit | Clean — no secrets, private repo refs acceptable |
| GitHub topics | Set (7 topics) |
| CONTRIBUTING.md | Fixed (GOWORK=off + panic-free correction) |
| README | GOWORK=off note added |
| SECURITY.md | Created |
| CI | Green, optimized (37s quality gate via golangci-lint-action) |
| pkg.go.dev | Fully indexed — but serving old README |
| Branch protection | Not configured (blocked on auto-git daemon decision) |
| Quality gate | Passes (test, lint, build) |
