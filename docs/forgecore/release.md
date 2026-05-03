# ForgeCore Release Strategy

ForgeCore uses semantic versioning for public releases.

## Versioning

- `MAJOR`: breaking SDK, proto, event, or schema contract changes
- `MINOR`: backward-compatible functionality
- `PATCH`: fixes and internal improvements

## Breaking Changes

A change is breaking when it:

- removes or renames a public shared SDK API
- removes an event field
- changes event semantics without a new event name
- reuses or removes a proto field number
- changes a migration contract incompatibly

## Release Checklist

- Update `CHANGELOG.md`.
- Update compatibility matrix.
- Run all verification scripts.
- Build every Go module.
- Run available tests.
- Tag release after CI passes.
