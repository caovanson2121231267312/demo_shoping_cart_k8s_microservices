# Reset infra volumes + migrations + fake data (from scratch)
# Usage:
#   .\scripts\reset-and-seed.ps1              # dev profile (50 users, 200 products, ...)
#   .\scripts\reset-and-seed.ps1 -Profile stress  # 1M users, 2K products, 50M orders

param(
    [ValidateSet("dev", "full", "stress")]
    [string]$Profile = "dev",
    [switch]$SkipReset
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot

function Test-Docker {
    try {
        docker info 2>$null | Out-Null
        return ($LASTEXITCODE -eq 0)
    } catch { return $false }
}

function Load-Env($path) {
    if (-not (Test-Path $path)) { return }
    Get-Content $path | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
            [Environment]::SetEnvironmentVariable($matches[1].Trim(), $matches[2].Trim().Trim('"'), 'Process')
        }
    }
}

Write-Host ""
Write-Host "=== Reset & seed fake data (profile: $Profile) ===" -ForegroundColor Cyan
Write-Host ""

if (-not (Test-Docker)) {
    Write-Host "Docker chua chay. Mo Docker Desktop roi chay lai:" -ForegroundColor Red
    Write-Host "  .\scripts\reset-and-seed.ps1 -Profile $Profile" -ForegroundColor Yellow
    exit 1
}

if (-not $SkipReset) {
    Write-Host "[1/5] Xoa volumes DB (reset tu dau)..." -ForegroundColor Cyan
    Push-Location "$root\services\auth-service"
    docker compose down -v postgres redis zookeeper kafka 2>&1 | Out-Host
    Pop-Location
    Push-Location "$root\services\product-service"
    docker compose down -v postgres mongo elasticsearch 2>&1 | Out-Host
    Pop-Location
    Push-Location "$root\services\order-service"
    docker compose down -v postgres 2>&1 | Out-Host
    Pop-Location
    Push-Location "$root\services\chat-service"
    if (Test-Path "docker-compose.yml") {
        docker compose down -v mongo redis 2>&1 | Out-Host
    }
    Pop-Location
    Write-Host "  OK volumes cleared" -ForegroundColor Green
}

Write-Host ""
Write-Host "[2/5] Khoi dong infra..." -ForegroundColor Cyan
& "$root\scripts\start-local-infra.ps1"
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Write-Host "[3/5] Chay migrations..." -ForegroundColor Cyan
Load-Env "$root\services\auth-service\.env"
if (-not $env:DATABASE_URL) {
    $env:DATABASE_URL = "postgres://auth:authpass@localhost:5432/authdb?sslmode=disable"
}
Push-Location "$root\services\auth-service"
go run cmd/migrate/main.go up
if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
Pop-Location

Push-Location "$root\services\product-service"
go run cmd/migrate/main.go
if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
Pop-Location

Push-Location "$root\services\order-service"
go run cmd/migrate/main.go
if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
Pop-Location
Write-Host "  OK migrations" -ForegroundColor Green

Write-Host ""
Write-Host "[4/5] Seed fake data (auth, product, order)..." -ForegroundColor Cyan
& "$root\scripts\seed-scale.ps1" -Profile $Profile
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Write-Host "[5/5] Seed chat (Mongo)..." -ForegroundColor Cyan
Load-Env "$root\services\chat-service\.env"
Push-Location "$root\services\chat-service"
go run scripts/fake_data.go
if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
Pop-Location

Write-Host ""
Write-Host "=== Hoan tat reset & seed ===" -ForegroundColor Green
Write-Host "  Login admin: admin@shop.com / Admin@123"
Write-Host "  Customers:   user1@shop.com .. / Customer@123"
Write-Host ""
Write-Host "Chay services: .\scripts\start-local-services.ps1" -ForegroundColor Cyan
