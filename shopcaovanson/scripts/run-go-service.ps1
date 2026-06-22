# Chay mot Go microservice — uu tien WSL de tranh Windows Smart App Control chan .exe
param(
    [Parameter(Mandatory)][string]$Name,
    [Parameter(Mandatory)][string]$Dir,
    [switch]$UseWsl
)

$ErrorActionPreference = 'Continue'

function Write-AppControlHelp {
    Write-Host ""
    Write-Host "Windows Smart App Control dang chan binary Go (.exe)." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Cach 1 (khuyen nghi): Cai Go trong WSL, roi chay lai script:" -ForegroundColor White
    Write-Host "  wsl --install -d Ubuntu" -ForegroundColor DarkGray
    Write-Host "  wsl sudo apt update && sudo apt install -y golang-go" -ForegroundColor DarkGray
    Write-Host "  .\scripts\start-local-services.ps1" -ForegroundColor DarkGray
    Write-Host ""
    Write-Host "Cach 2: Tat Smart App Control" -ForegroundColor White
    Write-Host "  Cai dat > Quyen rieng tu va bao mat > Bao mat Windows" -ForegroundColor DarkGray
    Write-Host "  > Kiem soat ung dung va trinh duyet > Cai dat Kiểm soat Ung dung Thong minh > Tat" -ForegroundColor DarkGray
    Write-Host ""
    Write-Host "Cach 3: Chay thu trong WSL (sau khi cai Go):" -ForegroundColor White
    Write-Host "  wsl --cd `"$Dir`" go run ." -ForegroundColor DarkGray
    Write-Host ""
}

if (-not (Test-Path $Dir)) {
    Write-Host "Khong tim thay thu muc service: $Dir" -ForegroundColor Red
    exit 1
}

$resolvedDir = (Resolve-Path $Dir).Path
Write-Host "[$Name] Ctrl+C de dung" -ForegroundColor Cyan

if ($UseWsl) {
    & wsl --cd $resolvedDir go run .
    exit $LASTEXITCODE
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
    Write-AppControlHelp
}
exit $code
