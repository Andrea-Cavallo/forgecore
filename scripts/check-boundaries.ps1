$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$violations = New-Object System.Collections.Generic.List[string]

function Test-ImportBoundary {
    param(
        [string]$Layer,
        [string]$Pattern,
        [string[]]$Forbidden
    )

    $files = Get-ChildItem -Path (Join-Path $root "services") -Recurse -File -Filter "*.go" |
        Where-Object { $_.FullName -match $Pattern }

    foreach ($file in $files) {
        $content = Get-Content -LiteralPath $file.FullName -Raw
        foreach ($blocked in $Forbidden) {
            if ($content -match [regex]::Escape($blocked)) {
                $relative = $file.FullName.Substring($root.Path.Length + 1)
                $violations.Add("${Layer} vietato in ${relative}: $blocked")
            }
        }
    }
}

Test-ImportBoundary `
    -Layer "domain" `
    -Pattern "\\internal\\domain\\" `
    -Forbidden @("/internal/infrastructure", "github.com/jackc/pgx", "github.com/redis/go-redis", "github.com/nats-io/nats.go", "net/http")

Test-ImportBoundary `
    -Layer "application" `
    -Pattern "\\internal\\application\\" `
    -Forbidden @("/internal/infrastructure", "github.com/jackc/pgx", "github.com/redis/go-redis")

if ($violations.Count -gt 0) {
    foreach ($violation in $violations) {
        Write-Error $violation
    }
    exit 1
}

Write-Output "Boundary DDD verificate"
