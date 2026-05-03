$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$sdk = Join-Path $root "sdk\go"
$violations = New-Object System.Collections.Generic.List[string]

foreach ($client in @("auth", "permission")) {
    $file = Join-Path $sdk "$client\client.go"
    if (-not (Test-Path -LiteralPath $file)) {
        $violations.Add("client mancante: $client")
        continue
    }
    $content = Get-Content -LiteralPath $file -Raw
    if ($content -notmatch "sdk/go/clientretry") {
        $violations.Add("$client non usa clientretry")
    }
    if ($content -notmatch "sdk/go/clienttransport") {
        $violations.Add("$client non usa clienttransport")
    }
}

if ($violations.Count -gt 0) {
    foreach ($violation in $violations) {
        Write-Error $violation
    }
    exit 1
}

Write-Output "Client SDK verificati"
