$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$files = Get-ChildItem -Path (Join-Path $root "services") -Recurse -File -Filter "*.up.sql"
$violations = New-Object System.Collections.Generic.List[string]

foreach ($file in $files) {
    $content = Get-Content -LiteralPath $file.FullName -Raw
    $tables = [regex]::Matches($content, 'CREATE TABLE(?: IF NOT EXISTS)?\s+([a-z_]+)')
    foreach ($match in $tables) {
        $table = $match.Groups[1].Value
        if ($content -notmatch "tenant_id\s+UUID\s+NOT\s+NULL") {
            $violations.Add("$($file.Name): tenant_id mancante")
        }
        if ($content -notmatch "CREATE INDEX .*${table}.*tenant|CREATE INDEX .*tenant.*${table}|CREATE INDEX ON ${table}\(tenant_id\)") {
            $violations.Add("$($file.Name): indice tenant mancante per $table")
        }
        if ($content -notmatch "ALTER TABLE ${table} ENABLE ROW LEVEL SECURITY") {
            $violations.Add("$($file.Name): RLS mancante per $table")
        }
        if ($content -notmatch "CREATE POLICY tenant_isolation ON ${table}") {
            $violations.Add("$($file.Name): policy tenant_isolation mancante per $table")
        }
    }
}

if ($violations.Count -gt 0) {
    foreach ($violation in $violations) {
        Write-Error $violation
    }
    exit 1
}

Write-Output "Migrazioni tenant/RLS verificate"
