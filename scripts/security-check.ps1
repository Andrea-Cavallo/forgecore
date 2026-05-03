$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")

if (Get-Command govulncheck -ErrorAction SilentlyContinue) {
    $modules = Get-ChildItem -Path $root -Recurse -File -Filter "go.mod" |
        Where-Object { $_.FullName -notmatch "\\.opencode\\|\\.agents\\|graphify-out" }
    foreach ($mod in $modules) {
        $dir = Split-Path $mod.FullName -Parent
        Push-Location $dir
        try {
            govulncheck ./...
        } finally {
            Pop-Location
        }
    }
} else {
    Write-Output "SKIP govulncheck: installare golang.org/x/vuln/cmd/govulncheck@latest"
}

powershell -ExecutionPolicy Bypass -File (Join-Path $root "scripts\e2e-gateway.ps1")
powershell -ExecutionPolicy Bypass -File (Join-Path $root "scripts\check-rbac-security.ps1")

$testModules = @(
    "shared",
    "sdk\go",
    "services\forgecore-gateway",
    "services\forgecore-auth",
    "services\forgecore-webhooks",
    "services\forgecore-payments"
)
foreach ($module in $testModules) {
    Push-Location (Join-Path $root $module)
    try {
        go test ./...
    } finally {
        Pop-Location
    }
}

Write-Output "Security checks disponibili completati"
