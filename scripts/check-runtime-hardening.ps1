$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")

$required = @(
    "shared\health\health.go",
    "shared\outbox\outbox.go",
    "shared\idempotency\idempotency.go"
)

foreach ($path in $required) {
    $full = Join-Path $root $path
    if (-not (Test-Path -LiteralPath $full)) {
        throw "Runtime hardening mancante: $path"
    }
}

$services = Get-ChildItem -Path (Join-Path $root "services") -Directory -Filter "forgecore-*" |
    Where-Object { $_.Name -ne "forgecore-example" }
foreach ($service in $services) {
    $files = Get-ChildItem -Path $service.FullName -Recurse -File -Include "*.go"
    $text = ($files | ForEach-Object { Get-Content -LiteralPath $_.FullName -Raw }) -join "`n"
    if ($text -notmatch "/readyz|health.Register|forgecore-jobs") {
        throw "Readiness non rilevata per $($service.Name)"
    }
}

Write-Output "Runtime hardening verificato"
