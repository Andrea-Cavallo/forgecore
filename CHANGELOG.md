# Changelog

All notable changes to ForgeCore will be documented here.

The project follows semantic versioning once public releases start.

## Unreleased

### Added

- ForgeCore naming across services.
- Shared config SDK with default, YAML, and ENV precedence.
- Versioned NATS event metadata.
- Tenant/RLS migration checks.
- DDD boundary checks.
- SDK client verification.
- Operational runbooks.
- Security baseline.
- Service scaffold script.

### Changed

- `sdk/go/common` split into `sdk/go/clientretry` and `sdk/go/clienttransport`.

### Fixed

- Shared validation build compatibility.
- Payment webhook application boundary violation.
