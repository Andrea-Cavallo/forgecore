$ErrorActionPreference = "Stop"

function Invoke-Checked {
    param([string[]]$Command)
    & $Command[0] @($Command | Select-Object -Skip 1)
    if ($LASTEXITCODE -ne 0) {
        throw "Comando fallito: $($Command -join ' ')"
    }
}

Invoke-Checked @("docker", "compose", "--env-file", ".env.example", "up", "-d", "postgres", "redis", "nats")
Invoke-Checked @("docker", "compose", "exec", "-T", "postgres", "pg_isready", "-U", "forgecore")
Invoke-Checked @("docker", "compose", "exec", "-T", "redis", "redis-cli", "ping")
Invoke-Checked @("docker", "compose", "exec", "-T", "nats", "wget", "--spider", "-q", "http://localhost:8222/healthz")

Write-Output "Integration locale infrastruttura completata"
