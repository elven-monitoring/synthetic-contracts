# synthetic-contracts

Shared, versioned Go contracts used by Synthetic services (control plane and data plane).

This repository contains only types/helpers that must be kept consistent across services.
It should not contain service logic.

## Packages

- `job/`: job payloads, execution messages and helpers.
- `job/streams.go`: Redis Streams naming/helpers used by executors and scheduler.

## Usage

Import from other repos:

```go
import "github.com/elven-monitoring/synthetic-contracts/job"
```

Update dependency (pin to a ref):

```bash
go get github.com/elven-monitoring/synthetic-contracts@<ref>
go mod tidy
```

## Development

Run tests:

```bash
go test ./...
```

## Versioning

Recommendation: tag releases (`v0.x.y`) when changing contracts, and keep changes backwards-compatible whenever possible.

