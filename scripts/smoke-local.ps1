$ErrorActionPreference = "Stop"

docker compose config --quiet

$checks = @(
    ".\scripts\check-boundaries.ps1",
    ".\scripts\check-proto-contracts.ps1",
    ".\scripts\check-sdk-clients.ps1",
    ".\scripts\check-tenant-migrations.ps1",
    ".\scripts\check-runtime-hardening.ps1",
    ".\scripts\check-dockerfiles.ps1",
    ".\scripts\check-rbac-security.ps1"
)

foreach ($check in $checks) {
    & powershell -NoProfile -ExecutionPolicy Bypass -File $check
    if ($LASTEXITCODE -ne 0) {
        throw "Check locale fallito: $check"
    }
}

Write-Output "Smoke locale completato"
