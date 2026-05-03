$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$legacyNames = @(
    "api-gateway",
    "auth-service",
    "payment-service",
    "notification-service",
    "admin-service",
    "audit-service",
    "job-service",
    "permission-service",
    "config-service",
    "webhook-service",
    "storage-service",
    "subscription-service"
)

$dockerfiles = Get-ChildItem -Path (Join-Path $root "services") -Recurse -File -Filter "Dockerfile"
foreach ($file in $dockerfiles) {
    $content = Get-Content -LiteralPath $file.FullName -Raw
    if ($content -match "golang:1\.24") {
        throw "Dockerfile usa Go legacy 1.24: $($file.FullName)"
    }
    foreach ($legacy in $legacyNames) {
        if ($content -match [regex]::Escape("services/$legacy")) {
            throw "Dockerfile usa path legacy ${legacy}: $($file.FullName)"
        }
    }
}

Write-Output "Dockerfile ForgeCore verificati"
