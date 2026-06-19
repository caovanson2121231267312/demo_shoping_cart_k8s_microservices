# Seed large-scale fake data for load / performance testing
#
# Full scale (user request):
#   .\scripts\seed-scale.ps1 -Profile full
#
# Custom:
#   .\scripts\seed-scale.ps1 -Users 3000000 -Products 1000 -Orders 1000000 -Reviews 50000000
#
# Dev / quick test:
#   .\scripts\seed-scale.ps1 -Profile dev

param(
    [ValidateSet("dev", "full", "custom")]
    [string]$Profile = "custom",

    [int]$Users = 3000000,
    [int]$Products = 1000,
    [int]$Orders = 1000000,
    [int]$Reviews = 50000000,
    [int]$BatchSize = 10000,
    [int]$BcryptCost = 10,

    [switch]$SkipReviews,
    [switch]$SkipOrders,
    [switch]$SkipUsers,
    [switch]$SkipProducts
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot

switch ($Profile) {
    "dev" {
        $Users = 50
        $Products = 200
        $Orders = 300
        $Reviews = 2000
        $BatchSize = 5000
    }
    "full" {
        $Users = 3000000
        $Products = 1000
        $Orders = 1000000
        $Reviews = 50000000
        $BatchSize = 20000
    }
}

function Load-Env($path) {
    if (-not (Test-Path $path)) { return }
    Get-Content $path | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
            [Environment]::SetEnvironmentVariable($matches[1].Trim(), $matches[2].Trim().Trim('"'), 'Process')
        }
    }
}

function Test-InfraPort($port, $label) {
    try {
        $tcp = New-Object System.Net.Sockets.TcpClient
        $tcp.Connect("localhost", $port)
        $tcp.Close()
        return $true
    } catch {
        Write-Host "  WARN $label (port $port) chua mo" -ForegroundColor Yellow
        return $false
    }
}

Write-Host ""
Write-Host "=== Seed scale profile: $Profile ===" -ForegroundColor Cyan
Write-Host "  Users:    $Users"
Write-Host "  Products: $Products"
Write-Host "  Orders:   $Orders"
Write-Host "  Reviews:  $Reviews"
Write-Host "  Batch:    $BatchSize"
Write-Host ""

if ($Profile -eq "full") {
    Write-Host "CANH BAO: Full scale can nhieu gio va ~15-25GB disk (Postgres)." -ForegroundColor Yellow
    Write-Host "  - 3M users:     ~30-60 phut (bcrypt cost=$BcryptCost)"
    Write-Host "  - 1K products:  ~2-5 phut"
    Write-Host "  - 50M reviews:  ~2-6 gio (COPY batch)"
    Write-Host "  - 1M orders:    ~20-60 phut"
    Write-Host ""
    $confirm = Read-Host "Tiep tuc? (y/N)"
    if ($confirm -notmatch '^[yY]') {
        Write-Host "Da huy." -ForegroundColor DarkGray
        exit 0
    }
}

Write-Host "Checking infra ports..." -ForegroundColor Cyan
$null = Test-InfraPort 5432 "auth postgres"
$null = Test-InfraPort 5433 "product postgres"
$null = Test-InfraPort 5434 "order postgres"

$sw = [System.Diagnostics.Stopwatch]::StartNew()

if (-not $SkipUsers) {
    Write-Host ""
    Write-Host "[1/3] Auth — $Users users" -ForegroundColor Cyan
    Load-Env "$root\services\auth-service\.env"
    if (-not $env:DATABASE_URL) {
        $env:DATABASE_URL = "postgres://auth:authpass@localhost:5432/authdb?sslmode=disable"
    }
    $env:SEED_USERS = "$Users"
    $env:SEED_BATCH_SIZE = "$BatchSize"
    $env:SEED_BCRYPT_COST = "$BcryptCost"
    Push-Location "$root\services\auth-service"
    go run scripts/fake_data.go
    if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
    Pop-Location
}

if (-not $SkipProducts -or -not $SkipReviews) {
    Write-Host ""
    Write-Host "[2/3] Product — $Products products, $Reviews reviews" -ForegroundColor Cyan
    Load-Env "$root\services\product-service\.env"
    $env:SEED_PRODUCTS = "$Products"
    $env:SEED_ARTICLES = "120"
    $env:SEED_USERS = "$Users"
    $env:SEED_BATCH_SIZE = "$BatchSize"
    if ($SkipReviews) {
        $env:SEED_REVIEWS = "0"
    } else {
        $env:SEED_REVIEWS = "$Reviews"
    }
    if ($Profile -eq "full") {
        $env:SEED_SKIP_MONGO_DETAILS = "false"
        $env:SEED_REFRESH_TEXT = "true"
        $env:SEED_ES_INDEX = "true"
    } elseif ($Profile -eq "dev") {
        $env:SEED_REFRESH_TEXT = "true"
        $env:SEED_ES_INDEX = "true"
    }
    Push-Location "$root\services\product-service"
    go run scripts/fake_data.go
    if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
    Pop-Location
}

if (-not $SkipOrders) {
    Write-Host ""
    Write-Host "[3/3] Order — $Orders orders" -ForegroundColor Cyan
    Load-Env "$root\services\order-service\.env"
    $env:SEED_ORDERS = "$Orders"
    $env:SEED_USERS = "$Users"
    $env:SEED_PRODUCTS = "$Products"
    $env:SEED_BATCH_SIZE = "$BatchSize"
    Push-Location "$root\services\order-service"
    go run scripts/fake_data.go
    if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
    Pop-Location
}

$sw.Stop()
Write-Host ""
Write-Host "=== Seed complete ($([math]::Round($sw.Elapsed.TotalMinutes, 1)) min) ===" -ForegroundColor Green
Write-Host "  Login: admin@shop.com / Admin@123"
Write-Host "  Customers: user1@shop.com .. user${Users}@shop.com / Customer@123"
Write-Host ""
Write-Host "Test search (vi du):" -ForegroundColor Cyan
Write-Host "  iPhone | Samsung Galaxy | tai nghe Sony | laptop gaming | nuoc hoa Chanel"
Write-Host "  noi com dien | robot hut bui | giay Nike | sach Atomic Habits"
