# Seed large-scale fake data for load / performance testing
#
# Load test 60M đơn / 1M user / 1M SP:
#   .\scripts\seed-scale.ps1 -Profile load60m -Force
#
# Full scale:
#   .\scripts\seed-scale.ps1 -Profile full
#
# Custom:
#   .\scripts\seed-scale.ps1 -Users 1000000 -Products 2000 -Orders 50000000 -Reviews 0
#
# Dev / quick test:
#   .\scripts\seed-scale.ps1 -Profile dev

param(
    [ValidateSet("dev", "full", "stress", "load60m", "custom")]
    [string]$Profile = "custom",

    [int]$Users = 1000000,
    [int]$Products = 1000000,
    [int]$Orders = 60000000,
    [int]$Reviews = 0,
    [int]$BatchSize = 10000,
    [int]$BcryptCost = 10,

    [switch]$SkipReviews,
    [switch]$SkipOrders,
    [switch]$SkipUsers,
    [switch]$SkipProducts,
    [switch]$Force
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
    "stress" {
        $Users = 1000000
        $Products = 2000
        $Orders = 50000000
        $Reviews = 0
        $BatchSize = 20000
        $SkipReviews = $true
    }
    "load60m" {
        $Users = 1000000
        $Products = 1000000
        $Orders = 60000000
        $Reviews = 0
        $BatchSize = 20000
        $SkipReviews = $true
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

function Invoke-GoSeed($workDir, [string[]]$scripts) {
    $binDir = Join-Path $root ".local\bin"
    New-Item -ItemType Directory -Force -Path $binDir | Out-Null
    $svcName = Split-Path $workDir -Leaf
    $exe = Join-Path $binDir "$svcName-seed.exe"
    Push-Location $workDir
    $prev = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    & go build -o $exe @scripts
    if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
    & $exe 2>&1 | ForEach-Object { Write-Host $_ }
    $code = $LASTEXITCODE
    $ErrorActionPreference = $prev
    Pop-Location
    if ($code -ne 0) { exit $code }
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

if ($Profile -eq "full" -or $Profile -eq "stress" -or $Profile -eq "load60m") {
    Write-Host 'CANH BAO: Scale lon - can nhieu gio va hang chuc GB disk (Postgres).' -ForegroundColor Yellow
    if ($Profile -eq "load60m") {
        Write-Host "  - 1M users:     ~30-60 phut (bcrypt cost=$BcryptCost)"
        Write-Host "  - 1M products:  ~1-3 gio (COPY bulk)"
        Write-Host "  - 60M orders:   ~15-40+ gio"
        Write-Host "  - Disk uoc tinh: 100-200 GB+"
    } elseif ($Profile -eq "stress") {
        Write-Host "  - 1M users:     ~20-40 phut (bcrypt cost=$BcryptCost)"
        Write-Host "  - 2K products:  ~5-10 phut"
        Write-Host "  - 50M orders:   ~10-30+ gio (tuy o cung)"
        Write-Host "  - Reviews:      bo qua (de nhanh hon)"
    } else {
        Write-Host "  - 3M users:     ~30-60 phut (bcrypt cost=$BcryptCost)"
        Write-Host "  - 1K products:  ~2-5 phut"
        Write-Host "  - 50M reviews:  ~2-6 gio (COPY batch)"
        Write-Host "  - 1M orders:    ~20-60 phut"
    }
    Write-Host ""
    if (-not $Force) {
        $confirm = Read-Host "Tiep tuc? (y/N)"
        if ($confirm -notmatch '^[yY]') {
            Write-Host "Da huy." -ForegroundColor DarkGray
            exit 0
        }
    }
}

Write-Host "Checking infra ports..." -ForegroundColor Cyan
$null = Test-InfraPort 5432 "auth postgres"
$null = Test-InfraPort 5433 "product postgres"
$null = Test-InfraPort 5434 "order postgres"

$sw = [System.Diagnostics.Stopwatch]::StartNew()

if (-not $SkipUsers) {
    Write-Host ""
    Write-Host "[1/3] Auth - $Users users" -ForegroundColor Cyan
    Load-Env "$root\services\auth-service\.env"
    if (-not $env:DATABASE_URL) {
        $env:DATABASE_URL = "postgres://auth:authpass@localhost:5432/authdb?sslmode=disable"
    }
    $env:SEED_USERS = "$Users"
    $env:SEED_BATCH_SIZE = "$BatchSize"
    $env:SEED_BCRYPT_COST = "$BcryptCost"
    Invoke-GoSeed "$root\services\auth-service" "scripts/fake_data.go"
}

if (-not $SkipProducts -or -not $SkipReviews) {
    Write-Host ""
    Write-Host "[2/3] Product - $Products products, $Reviews reviews" -ForegroundColor Cyan
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
    if ($Profile -eq "full" -or $Profile -eq "stress" -or $Profile -eq "load60m") {
        $env:SEED_SKIP_MONGO_DETAILS = "true"
        $env:SEED_BULK_PRODUCTS = "true"
        $env:SEED_REFRESH_TEXT = "false"
        $env:SEED_ES_INDEX = "false"
    } elseif ($Profile -eq "dev") {
        $env:SEED_REFRESH_TEXT = "true"
        $env:SEED_ES_INDEX = "true"
    }
    Invoke-GoSeed "$root\services\product-service" @("scripts/fake_data.go", "scripts/es_bulk.go")
}

if (-not $SkipOrders) {
    Write-Host ""
    Write-Host "[3/3] Order - $Orders orders" -ForegroundColor Cyan
    Load-Env "$root\services\order-service\.env"
    $env:SEED_ORDERS = "$Orders"
    $env:SEED_USERS = "$Users"
    $env:SEED_PRODUCTS = "$Products"
    $env:SEED_BATCH_SIZE = "$BatchSize"
    Invoke-GoSeed "$root\services\order-service" "scripts/fake_data.go"
}

$sw.Stop()
$elapsedMin = [math]::Round($sw.Elapsed.TotalMinutes, 1)
Write-Host ""
Write-Host "=== Seed complete ($elapsedMin min) ===" -ForegroundColor Green
Write-Host "  Login: admin@shop.com / Admin@123"
Write-Host "  Customers: user1@shop.com .. user${Users}@shop.com / Customer@123"
Write-Host ""
Write-Host "Test search (vi du):" -ForegroundColor Cyan
Write-Host "  iPhone | Samsung Galaxy | tai nghe Sony | laptop gaming | nuoc hoa Chanel"
Write-Host "  noi com dien | robot hut bui | giay Nike | sach Atomic Habits"
