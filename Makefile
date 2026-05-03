.PHONY: build test-shared test-sdk test-auth test-e2e verify boundaries proto-check sdk-check tenant-check runtime-check dockerfile-check rbac-check security-check integration-check smoke scaffold scaffold-dryrun

build:
	powershell -ExecutionPolicy Bypass -File ./scripts/build-all.ps1

test-shared:
	cd shared && go test ./...

test-sdk:
	cd sdk/go && go test ./...

test-auth:
	cd services/forgecore-auth && go test ./...

test-e2e:
	powershell -ExecutionPolicy Bypass -File ./scripts/e2e-gateway.ps1

boundaries:
	powershell -ExecutionPolicy Bypass -File ./scripts/check-boundaries.ps1

proto-check:
	powershell -ExecutionPolicy Bypass -File ./scripts/check-proto-contracts.ps1

sdk-check:
	powershell -ExecutionPolicy Bypass -File ./scripts/check-sdk-clients.ps1

tenant-check:
	powershell -ExecutionPolicy Bypass -File ./scripts/check-tenant-migrations.ps1

runtime-check:
	powershell -ExecutionPolicy Bypass -File ./scripts/check-runtime-hardening.ps1

dockerfile-check:
	powershell -ExecutionPolicy Bypass -File ./scripts/check-dockerfiles.ps1

rbac-check:
	powershell -ExecutionPolicy Bypass -File ./scripts/check-rbac-security.ps1

security-check:
	powershell -ExecutionPolicy Bypass -File ./scripts/security-check.ps1

integration-check:
	powershell -ExecutionPolicy Bypass -File ./scripts/integration-local.ps1

verify: boundaries proto-check sdk-check tenant-check runtime-check dockerfile-check rbac-check test-shared test-sdk test-auth test-e2e build

smoke:
	powershell -ExecutionPolicy Bypass -File ./scripts/smoke-local.ps1

scaffold:
	powershell -ExecutionPolicy Bypass -File ./scripts/scaffold-service.ps1 -Name $(name)

scaffold-dryrun:
	powershell -ExecutionPolicy Bypass -File ./scripts/scaffold-service.ps1 -Name $(name) -DryRun
