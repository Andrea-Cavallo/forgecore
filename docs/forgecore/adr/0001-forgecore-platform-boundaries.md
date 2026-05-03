# ADR-0001: ForgeCore Platform Boundaries

## Status

Accepted

## Context

ForgeCore is both a backend SDK and a reference microservices platform. Without clear boundaries, shared code can become a dumping ground and services can duplicate infrastructure logic.

## Decision

- `shared/` contains stable SDK primitives.
- `sdk/go/` contains reusable clients for consumers.
- `services/forgecore-*` contains reference bounded contexts.
- DDD boundaries are enforced by `scripts/check-boundaries.ps1`.
- Configuration is centralized through `shared/config*` and `forgecore-config`.

## Consequences

- New services must reuse SDK primitives instead of local helpers.
- Generic package names are rejected.
- Runtime service code can evolve independently as long as shared contracts remain compatible.
