# Chay mot Go microservice - uu tien WSL de tranh Windows Smart App Control chan .exe
param(
    [Parameter(Mandatory)][string]$Name,
    [Parameter(Mandatory)][string]$Dir,
    [switch]$UseWsl,
    [string]$WslDistro
)

$ErrorActionPreference = 'Continue'
. (Join-Path $PSScriptRoot 'lib\wsl-go.ps1')

if (-not (Test-Path $Dir)) {
    Write-Host "Khong tim thay thu muc service: $Dir" -ForegroundColor Red
    exit 1
}

$resolvedDir = (Resolve-Path $Dir).Path
Write-Host "[$Name] Ctrl+C de dung" -ForegroundColor Cyan

$distro = $WslDistro
if ($UseWsl -and -not $distro) {
    $distro = Find-WslDistroWithGo
    if (-not $distro) {
        Write-Host "WSL duoc yeu cau nhung chua co Go trong bat ky distro nao." -ForegroundColor Red
        Write-SmartAppControlHelp $resolvedDir
        Write-Host "Chay: .\scripts\setup-wsl-go.ps1" -ForegroundColor White
        exit 1
    }
}

if ($distro) {
    if ($distro -ne 'default') {
        Write-Host "[$Name] WSL ($distro): go run ." -ForegroundColor DarkGray
    }
    $code = Invoke-WslGo -Distro $distro -Dir $resolvedDir -GoArgs @('run', '.')
    exit $code
}

if (Test-SmartAppControlOn) {
    Write-Host "[$Name] CANH BAO: Smart App Control dang bat - co the chan .exe" -ForegroundColor Yellow
}

Set-Location $resolvedDir
$cacheRoot = Join-Path $env:LOCALAPPDATA 'shopcaovanson'
foreach ($sub in @('gocache', 'gotmp')) {
    $p = Join-Path $cacheRoot $sub
    if (-not (Test-Path $p)) {
        New-Item -ItemType Directory -Path $p -Force | Out-Null
    }
}
$env:GOCACHE = Join-Path $cacheRoot 'gocache'
$env:GOTMPDIR = Join-Path $cacheRoot 'gotmp'

& go run .
$code = $LASTEXITCODE
if ($code -ne 0) {
    Write-SmartAppControlHelp $resolvedDir
}
exit $code
