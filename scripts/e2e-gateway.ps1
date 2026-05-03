$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$gateway = Join-Path $root "services\forgecore-gateway"

Push-Location $gateway
try {
    go test ./cmd/server -run TestGatewayFrontendE2E -count=1
} finally {
    Pop-Location
}
