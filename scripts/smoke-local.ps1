$ErrorActionPreference = "Stop"

docker compose config --quiet

$checks = @(
    ".\scripts\check-boundaries.ps1",
    ".\scripts\check-proto-contracts.ps1",
    ".\scripts\check-sdk-clients.ps1",
    ".\scripts\check-tenant-migrations.ps1"
)

foreach ($check in $checks) {
    powershell -ExecutionPolicy Bypass -File $check
}

Write-Output "Smoke locale completato"
