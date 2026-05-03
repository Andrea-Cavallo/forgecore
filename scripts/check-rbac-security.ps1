$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$rbacCode = Get-Content -Path (Join-Path $root "services\forgecore-gateway\internal\middleware\rbac.go") -Raw
$auditCode = Get-Content -Path (Join-Path $root "services\forgecore-gateway\internal\middleware\audit.go") -Raw
$matrix = Get-Content -Path (Join-Path $root "docs\forgecore\rbac-endpoint-matrix.md") -Raw

$requiredPrefixes = @(
    "/v1/payments",
    "/v1/notifications",
    "/v1/admin",
    "/v1/audit",
    "/v1/permissions",
    "/v1/config",
    "/v1/webhooks",
    "/v1/storage",
    "/v1/subscriptions"
)

foreach ($prefix in $requiredPrefixes) {
    if ($rbacCode -notmatch [regex]::Escape($prefix)) {
        throw "RBAC mancante per prefix ${prefix}"
    }
    if ($matrix -notmatch [regex]::Escape($prefix)) {
        throw "Matrice RBAC mancante per prefix ${prefix}"
    }
}

$sensitivePrefixes = @(
    "/v1/admin",
    "/v1/config",
    "/v1/payments",
    "/v1/permissions",
    "/v1/storage",
    "/v1/subscriptions",
    "/v1/webhooks"
)

foreach ($prefix in $sensitivePrefixes) {
    if ($auditCode -notmatch [regex]::Escape($prefix)) {
        throw "Audit obbligatorio mancante per prefix ${prefix}"
    }
}

Write-Output "RBAC e audit endpoint verificati"
