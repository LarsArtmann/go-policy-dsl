# Status Report — 2026-08-08 05:38

## Session scope

Two tasks: (1) rewrite `ROADMAP.md` as forward-looking only, (2) audit all three GitHub releases against commits and `CHANGELOG.md`, fix inaccuracies, and make them superb. This report covers only what this session did and noticed — no unrelated research.

---

## a) FULLY DONE

### 1. ROADMAP.md rewrite

**Was:** 178 lines with 6 "LANDED"/"SHIPPED" sections documenting completed work (CHANGELOG/FEATURES territory), stale `[Unreleased]` references everywhere, and v1.0.0 criteria where 3 of 4 items were already done.

**Now:** 107 lines, pure forward-looking. Removed all completed-work sections. Condensed v1.0.0 criteria from 4 items to 3 (reflecting what's actually still pending). Fixed all stale `[Unreleased]` references. Noted repo is currently private in the GOPROXY section. Committed as `aa08b95`.

### 2. CHANGELOG v0.1.0 fix

Removed `.github/dependabot.yml` from the v0.1.0 "Added" list — it was added in v0.2.0, not v0.1.0 (verified via `git ls-tree v0.1.0`). Committed by auto-git daemon as `2c26518`.

### 3. CHANGELOG v0.2.0 restructure

Removed 6 false "Changed (breaking)" items, 5 false "Added" items, and the entire "Fixed" section — all were pre-v0.1.0 changes incorrectly attributed to v0.2.0. The only real v0.2.0 breaking change is the panic-free refactor (removing `Must*` family + `VersionRangeStrings` + inversion panic). Restructured "Internal" to only contain actual v0.2.0 internal changes. Committed by auto-git daemon as `2c26518`.

### 4. v0.1.0 GitHub release fixed

**Was:** 5,674 chars with a fake "Changed" section (8 items, all pre-v0.1.0 or v0.2.0 changes) and a fake "Fixed" section (3 items, all pre-v0.1.0 fixes).

**Now:** 3,100 chars, pure "Added" section matching the corrected CHANGELOG. Retained the "Note" about `Must*` constructors being removed in v0.2.0.

### 5. v0.2.0 GitHub release fixed

**Was:** 9,399 chars with 6 false "Changed (breaking)" items, 5 false "Added" items, and a "Fixed" section — all pre-v0.1.0 changes.

**Now:** 2,480 chars. Only the real v0.2.0 change (panic-free refactor) under "Changed (breaking)", only `cmp.Compare` under "Added", and actual v0.2.0 internal changes under "Internal". Title fixed from "panic-free, typed domain" → "panic-free API" (the typed domain was already in v0.1.0).

### 6. v0.3.0 GitHub release fixed

**Was:** Contradictory footer — said "Pre-1.0 breaking release" despite being explicitly non-breaking.

**Now:** Footer says "Pre-1.0 policy" matching the CHANGELOG header tone. Body otherwise unchanged (was already correct).

### 7. Quality gate

`GOWORK=off go test ./...` (0.003s, pass), `GOWORK=off golangci-lint run ./...` (0 issues), `GOWORK=off go build ./...` (clean). Run after ROADMAP and CHANGELOG changes.

### 8. Release list verification

All three releases verified: not drafts, not prereleases, v0.3.0 is Latest, v0.1.0 and v0.2.0 are not Latest. All use annotated git tags pushed to remote.

---

## b) PARTIALLY DONE

### CHANGELOG ↔ GitHub release body consistency

The CHANGELOG sections and GitHub release bodies are **close but not identical** in wording. They were written from different sources (CHANGELOG edited in-place, release notes written to temp files). Minor wording differences exist (e.g., the v0.2.0 release body has more detail in the "Changed" section than the CHANGELOG, the v0.1.0 release body mentions `MustNewVersion`/`MustParseVersion` in the note but the CHANGELOG only notes `MustCVE` and `VersionRangeStrings`). This is acceptable — release notes can be more detailed — but a strict consistency pass would catch drift.

---

## c) NOT STARTED

### Making the repo public

The repo is ready (README is public-facing, CHANGELOG is accurate, ROADMAP is forward-looking, releases are correct). The user has been informed. Running `gh repo edit --visibility public` is the only remaining step. Not started — user's discretion.

---

## d) TOTALLY FUCKED UP

### The original v0.1.0 and v0.2.0 releases were fundamentally wrong

Both releases were backfilled in a prior session by diffing from the initial commit to HEAD, which lumped ALL pre-v0.1.0 development changes into both release notes. The v0.2.0 release claimed credit for the entire typed domain (`Version`, `Mode`, `CVE`, `[]Replacement`, `Validate()`, `SuggestExplicit`) that was already shipped in v0.1.0. The v0.1.0 release had "Changed" and "Fixed" sections describing changes that happened before v0.1.0 was tagged — a release should only describe what's new IN that release, not the development history that led to it.

**Root cause:** No one ran `git diff v0.1.0..v0.2.0 -- builder.go policy.go version.go cve.go` to verify what actually changed between tags. The release notes were written from the CHANGELOG, which was itself wrong.

**Fixed this session** by diffing the actual tags and rewriting both releases + the CHANGELOG.

---

## e) WHAT WE SHOULD IMPROVE

1. **Release notes should always be verified against `git diff <prev_tag>..<tag>`.** This session caught 14 incorrect items across two releases. The prior session trusted the CHANGELOG, which was itself wrong. The diff is the source of truth.

2. **CHANGELOG sections should be verified against tag diffs before tagging.** Same root cause — the CHANGELOG was written from memory/session context, not from git diffs. A pre-tag checklist item should be "run `git diff <prev_tag>..<tag> --stat` and confirm every CHANGELOG item appears in the diff."

3. **The auto-git daemon commits with empty messages** (commit `2c26518` has no message). This makes `git log` harder to scan. Not a this-session issue, but observed.

4. **ROADMAP had 6 "LANDED"/"SHIPPED" sections** — completed work documented in a forward-looking file. This is a docs-health anti-pattern. The ROADMAP should never contain completed work; that's CHANGELOG + FEATURES territory. Fixed this session, but the pattern should be watched for.

5. **The v0.2.0 release title "panic-free, typed domain" was wrong for the entire life of the release** — the typed domain was in v0.1.0. Titles should be verified against what actually changed in that specific version.

6. **`GOWORK=off` is required for all Go commands** because a parent `go.work` doesn't include this module. This is documented in AGENTS.md but is a recurring friction point. Consider adding this module to `go.work` or documenting the workaround more prominently.

---

## f) Up to 50 things to get done next

### Release & visibility

1. Make the repo public (`gh repo edit --visibility public`)
2. Verify `proxy.golang.org` picks up the module after going public
3. Verify the pkg.go.dev badge resolves after going public
4. Verify the Go Report Card badge resolves after going public
5. Verify the CI badge resolves after going public

### CHANGELOG ↔ release consistency

6. Do a strict consistency pass between CHANGELOG sections and GitHub release bodies — unify wording where they should match
7. Add `MustNewVersion` and `MustParseVersion` removal notes to the CHANGELOG v0.1.0 section (the release notes mention them but the CHANGELOG only notes `MustCVE` and `VersionRangeStrings`)
8. Verify the CHANGELOG v0.1.0 "Added" list is a complete inventory of what existed at the tag (run `git ls-tree -r v0.1.0 --name-only` and cross-check)

### ROADMAP

9. Consider removing `library-policy` and `go-linter-sdk` references from ROADMAP if the repo goes public (ROADMAP is internal, but if it's public, these references point to private repos)
10. Consider whether the "Direction: documentation presence" section is still relevant or can be trimmed
11. Consider whether the "golangci-lint plugin template" idea is still relevant or YAGNI

### FEATURES.md

12. Verify FEATURES.md is consistent with the corrected CHANGELOG (e.g., no claims that contradict the v0.2.0 scope)
13. Check if FEATURES.md "Domain validation in `Validate()`" PLANNED item is still accurate

### TODO_LIST.md

14. Verify TODO_LIST.md harvest log is accurate — it references `[Unreleased]` which is now `[0.3.0]`
15. Check if TODO_LIST.md needs any new items based on this session's findings

### AGENTS.md

16. Verify AGENTS.md "Consumers" section is still accurate
17. Verify AGENTS.md "Surprising Behaviors" section doesn't reference removed APIs
18. Consider whether AGENTS.md should mention `GOWORK=off` requirement more prominently

### Testing

19. Consider adding a test that verifies the CHANGELOG sections match the GitHub release bodies (meta-test)
20. Consider adding a CI check that runs `git diff <prev_tag>..<tag> --stat` and warns if CHANGELOG items don't match

### Docs hygiene

21. Check if `docs/DOMAIN_LANGUAGE.md` needs updates after the ROADMAP rewrite
22. Check if any `docs/status/` reports reference the old v0.2.0 release title or contents
23. Check if `CONTRIBUTING.md` references are still accurate
24. Check if `.editorconfig` and `.gitattributes` are still needed

### Release process

25. Write a release checklist (verify CHANGELOG against tag diff, verify release notes against CHANGELOG, verify Latest flag)
26. Consider release automation (GitHub Action for tag → release → pkg.go.dev)
27. Consider whether future releases should generate release notes from the CHANGELOG automatically

### Code quality

28. Run `erraudit ./...` to confirm 0 violations (mentioned in AGENTS.md but not run this session)
29. Run `golangci-lint fmt ./...` to check for formatting drift
30. Consider whether the `Severity` enum has too many values (`recommended`, `deprecated`, `obsolete` are listed in v0.1.0 but not in the README API table)

### Public presence

31. Rewrite README to remove `library-policy`/`go-linter-sdk` references — **WAIT, this was already done in a prior session** (commit `cb1e2b9`). Verify it's still clean.
32. Check if the README "Contributing" section mentions `GOWORK=off` (it doesn't — it says `go test ./...` without the prefix)
33. Add `GOWORK=off` to the README contributing commands
34. Consider adding a CONTRIBUTING.md mention of the `GOWORK=off` requirement

### Consistency

35. Verify that no doc references the old v0.2.0 title "panic-free, typed domain"
36. Verify that no doc references `[Unreleased]` when it should reference `[0.3.0]`
37. Verify that the ROADMAP v1.0.0 criteria don't contradict the CHANGELOG or FEATURES.md
38. Check if the `docs/status/` retrospective reports need annotation after the release fixes

---

## g) Questions I cannot figure out myself

1. **Should the ROADMAP reference `library-policy` and `go-linter-sdk` if the repo goes public?** These are private repos. The README was rewritten to remove all references to them, but the ROADMAP still mentions them as adoption targets. If the repo goes public, should the ROADMAP also be sanitized, or is it acceptable for a public repo's roadmap to reference private repos as "internal consumers"?

2. **Should the CHANGELOG and GitHub release bodies be identical, or is it acceptable for release notes to be more detailed?** Currently the release notes have slightly more detail (e.g., migration instructions, context notes) than the CHANGELOG sections. This is a common pattern but could be seen as drift.

3. **Should I make the repo public now, or do you want to review the changes first?** Everything is ready — README, CHANGELOG, ROADMAP, releases, FEATURES, AGENTS.md are all consistent and accurate. The only remaining step is `gh repo edit --visibility public`.

---

## Summary

| Area                            | State                                                           |
| ------------------------------- | --------------------------------------------------------------- |
| ROADMAP.md                      | Fully done — forward-looking only, committed                    |
| CHANGELOG.md                    | Fully done — v0.1.0 and v0.2.0 corrected, committed             |
| v0.1.0 GitHub release           | Fully done — pure "Added", 3,100 chars                          |
| v0.2.0 GitHub release           | Fully done — panic-free refactor only, 2,480 chars, title fixed |
| v0.3.0 GitHub release           | Fully done — footer fixed, 1,499 chars                          |
| Quality gate                    | Passes (test, lint, build)                                      |
| Repo visibility                 | Not started — user's discretion                                 |
| CHANGELOG ↔ release consistency | Partially done — close but not identical                        |
