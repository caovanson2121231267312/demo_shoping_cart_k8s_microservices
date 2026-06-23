param(
    [Parameter(Mandatory)][string]$Dir,
    [switch]$UseWsl,
    [string]$WslDistro
)

$ErrorActionPreference = 'Continue'
. (Join-Path $PSScriptRoot 'lib\wsl-go.ps1')

$migrateMain = Join-Path $Dir 'cmd\migrate\main.go'
if (-not (Test-Path $migrateMain)) { exit 0 }

$resolvedDir = (Resolve-Path $Dir).Path
$goArgs = @('run', 'cmd/migrate/main.go', 'up')

$distro = $WslDistro
if ($UseWsl -and -not $distro) {
    $distro = Find-WslDistroWithGo
    if (-not $distro) {
        Write-Host "WARN migration: chua co Go trong WSL" -ForegroundColor Yellow
        exit 1
    }
}

if ($distro) {
    $code = Invoke-WslGo -Distro $distro -Dir $resolvedDir -GoArgs $goArgs
    exit $code
}

Set-Location $resolvedDir
$cacheRoot = Join-Path $env:LOCALAPPDATA 'shopcaovanson'
$env:GOCACHE = Join-Path $cacheRoot 'gocache'
$env:GOTMPDIR = Join-Path $cacheRoot 'gotmp'
& go @goArgs
exit $LASTEXITCODE
