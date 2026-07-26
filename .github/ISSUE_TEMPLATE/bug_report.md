---
name: Bug report
about: Report something that works incorrectly
labels: bug
---

## What happened?

<!-- A clear description of what you saw. Include the smallest policy + consumer
snippet that reproduces it. -->

## What did you expect?

## Reproduction

```go
// Minimal `policydsl.Ban(...)...Spec()` that triggers the issue.
```

```
# Command run + output (go test, golangci-lint, your consumer CLI, etc.)
```

## Environment

- `go-policy-dsl` version / commit:
- Go version (`go version`):
- Consumer (e.g. library-policy, your own tool):

## Quality gate status

The project gate is `go build ./... && go test ./... -race && go vet ./... && golangci-lint run ./... && golangci-lint fmt ./...`. Did it pass before the bug appeared?
