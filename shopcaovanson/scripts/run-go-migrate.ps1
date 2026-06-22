param(
    [Parameter(Mandatory)][string]$Dir,
    [switch]$UseWsl
)

$ErrorActionPreference = 'Continue'
$migrateMain = Join-Path $Dir 'cmd\migrate\main.go'
if (-not (Test-Path $migrateMain)) { exit 0 }

$resolvedDir = (Resolve-Path $Dir).Path

if ($UseWsl) {
    & wsl --cd $resolvedDir go run cmd/migrate/main.go up
    exit $LASTEXITCODE
}

Set-Location $resolvedDir
$cacheRoot = Join-Path $env:LOCALAPPDATA 'shopcaovanson'
$env:GOCACHE = Join-Path $cacheRoot 'gocache'
$env:GOTMPDIR = Join-Path $cacheRoot 'gotmp'
& go run cmd/migrate/main.go up
exit $LASTEXITCODE
