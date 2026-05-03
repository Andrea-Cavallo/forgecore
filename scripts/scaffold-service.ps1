param(
    [Parameter(Mandatory = $true)]
    [string]$Name,

    [switch]$DryRun
)

$ErrorActionPreference = "Stop"

if ($Name -notmatch '^forgecore-[a-z0-9-]+$') {
    throw "Il nome deve seguire il pattern forgecore-<bounded-context>"
}

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$serviceDir = Join-Path $root "services\$Name"
if (Test-Path -LiteralPath $serviceDir) {
    throw "Servizio gia' esistente: $Name"
}

New-Item -ItemType Directory -Force -Path `
    "$serviceDir\cmd\server", `
    "$serviceDir\internal\domain", `
    "$serviceDir\internal\application", `
    "$serviceDir\internal\infrastructure", `
    "$serviceDir\internal\transport", `
    "$serviceDir\migrations" | Out-Null

$module = "github.com/Andrea-Cavallo/golang-modules/services/$Name"

if ($DryRun) {
    Write-Output "DRY-RUN servizio valido: $Name"
    Write-Output "DRY-RUN directory: $serviceDir"
    Write-Output "DRY-RUN modulo: $module"
    exit 0
}

Set-Content -LiteralPath "$serviceDir\go.mod" -Value @"
module $module

go 1.26

require github.com/Andrea-Cavallo/golang-modules/shared v0.0.0

replace github.com/Andrea-Cavallo/golang-modules/shared => ../../shared
"@

Set-Content -LiteralPath "$serviceDir\cmd\server\main.go" -Value @"
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("avvio $Name fallito", "errore", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	slog.Info("$Name avviato")
	<-ctx.Done()
	return nil
}
"@

Set-Content -LiteralPath "$serviceDir\Dockerfile" -Value @"
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY . .
WORKDIR /src/services/$Name
RUN go build -o /out/service ./cmd/server

FROM alpine:3.20
COPY --from=build /out/service /service
ENTRYPOINT ["/service"]
"@

Write-Output "Servizio creato: $Name"
