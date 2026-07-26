---
name: Feature request
about: Suggest a new policy type, builder method, or behaviour
labels: enhancement
---

## The problem

<!-- What can't you express today, or what footgun did you hit? -->

## Proposed shape

```go
// Sketch the DSL call site you wish existed.
policydsl.Ban("...")...
```

## Consumers affected

<!-- Which consumer(s) would use this? (library-policy, go-linter-sdk, your own) -->

## Alternatives considered

## Stdlib-only contract check

This library is **stdlib-only with zero dependencies** (including test deps, by
choice). Does the proposal pull in any non-std import? If so, how should it be
avoided?
