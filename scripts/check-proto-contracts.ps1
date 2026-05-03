$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$protoRoot = Join-Path $root "shared\proto"
$files = Get-ChildItem -Path $protoRoot -File -Filter "*.proto"
$violations = New-Object System.Collections.Generic.List[string]

foreach ($file in $files) {
    $content = Get-Content -LiteralPath $file.FullName -Raw
    if ($content -notmatch 'syntax = "proto3";') {
        $violations.Add("$($file.Name): syntax proto3 mancante")
    }
    if ($content -notmatch 'package [a-z]+\.v1;') {
        $violations.Add("$($file.Name): package v1 mancante")
    }
    if ($content -notmatch 'option go_package = ".+/shared/proto/.+/v1;[a-z]+v1";') {
        $violations.Add("$($file.Name): go_package v1 non coerente")
    }
}

if ($violations.Count -gt 0) {
    foreach ($violation in $violations) {
        Write-Error $violation
    }
    exit 1
}

Write-Output "Contratti proto verificati"
