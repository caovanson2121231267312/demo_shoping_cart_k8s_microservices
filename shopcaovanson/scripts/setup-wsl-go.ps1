# Cai Ubuntu WSL (neu chua co) va golang-go - tranh Windows Smart App Control chan .exe
# Usage: .\scripts\setup-wsl-go.ps1
param(
    [string]$Distro = 'Ubuntu'
)

$ErrorActionPreference = 'Stop'
$Root = Split-Path -Parent $PSScriptRoot
. (Join-Path $PSScriptRoot 'lib\wsl-go.ps1')

Write-Host "shopcaovanson - setup Go trong WSL" -ForegroundColor Cyan
Write-Host ""

if (Test-SmartAppControlOn) {
    Write-Host "Smart App Control: ON (Windows se chan go run .exe)" -ForegroundColor Yellow
} else {
    Write-Host "Smart App Control: OFF hoac khong kha dung" -ForegroundColor DarkGray
}

if (-not (Get-Command wsl -ErrorAction SilentlyContinue)) {
    Write-Host "WSL chua cai. Chay (can quyen Admin, co the can khoi dong lai):" -ForegroundColor Yellow
    Write-Host "  wsl --install -d $Distro" -ForegroundColor White
    exit 1
}

$installed = Get-WslDistroNames
$hasDistro = $installed | Where-Object { $_ -eq $Distro }

if (-not $hasDistro) {
    Write-Host "Chua co WSL '$Distro'. Dang cai..." -ForegroundColor Cyan
    Write-Host 'Can quyen Admin; lan dau co the yeu cau tao user Ubuntu' -ForegroundColor DarkGray
    & wsl --install -d $Distro
    if ($LASTEXITCODE -ne 0) {
        Write-Host ""
        Write-Host "Cai Ubuntu that bai. Thu thu cong:" -ForegroundColor Red
        Write-Host "  wsl --install -d Ubuntu" -ForegroundColor White
        exit 1
    }
    Write-Host ""
    Write-Host "Ubuntu da cai. Neu chua tung mo Ubuntu, hay mo ung dung Ubuntu mot lan de tao user." -ForegroundColor Yellow
    Write-Host "Sau do chay lai: .\scripts\setup-wsl-go.ps1" -ForegroundColor Yellow
    exit 0
}

Write-Host "Cai golang-go trong WSL '$Distro'..." -ForegroundColor Cyan
$installCmd = 'export DEBIAN_FRONTEND=noninteractive; sudo apt-get update -qq && sudo apt-get install -y golang-go'
& wsl -d $Distro bash -lc $installCmd
if ($LASTEXITCODE -ne 0) {
    Write-Host "Cai Go that bai. Thu trong Ubuntu terminal:" -ForegroundColor Red
    Write-Host '  sudo apt update; sudo apt install -y golang-go' -ForegroundColor White
    exit 1
}

$ver = & wsl -d $Distro go version
Write-Host ""
Write-Host "OK  $ver" -ForegroundColor Green
Write-Host ""
Write-Host "Chay services:" -ForegroundColor Cyan
Write-Host "  .\scripts\start-local-services.ps1" -ForegroundColor White
Write-Host ""
Write-Host "Hoac mot service:" -ForegroundColor DarkGray
Write-Host '  .\scripts\run-go-service.ps1 -Name chat-service -Dir services\chat-service -UseWsl' -ForegroundColor DarkGray
