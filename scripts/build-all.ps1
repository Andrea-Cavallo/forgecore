$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$modules = Get-ChildItem -Path $root -Recurse -File -Filter "go.mod" |
    Where-Object { $_.FullName -notmatch "\\.opencode\\" }

foreach ($mod in $modules) {
    $dir = Split-Path $mod.FullName -Parent
    $relative = $dir.Substring($root.Path.Length + 1)
    Push-Location $dir
    try {
        go build ./...
        Write-Output "PASS $relative"
    } finally {
        Pop-Location
    }
}
