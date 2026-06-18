# Seed large-scale fake data (auth + product + order services)
# Requires running Postgres containers for each service.
#
# Example full scale (may take hours, needs sufficient RAM/disk):
#   .\scripts\seed-scale.ps1 -Users 2000000 -Products 1000 -Orders 1000000
#
# Dev scale (default):
#   .\scripts\seed-scale.ps1

param(
    [int]$Users = 50,
    [int]$Products = 1000,
    [int]$Orders = 300,
    [int]$BatchSize = 5000,
    [int]$BcryptCost = 10
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
if (-not $root) { $root = "d:\k8s\shopcaovanson" }

function Load-Env($path) {
    if (-not (Test-Path $path)) { return }
    Get-Content $path | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
            [Environment]::SetEnvironmentVariable($matches[1].Trim(), $matches[2].Trim(), 'Process')
        }
    }
}

Write-Host "=== Seed scale: users=$Users products=$Products orders=$Orders ===" -ForegroundColor Cyan

# Auth
Load-Env "$root\services\auth-service\.env"
$env:SEED_USERS = "$Users"
$env:SEED_BATCH_SIZE = "$BatchSize"
$env:SEED_BCRYPT_COST = "$BcryptCost"
Push-Location "$root\services\auth-service"
go run scripts/fake_data.go
Pop-Location

# Product
Load-Env "$root\services\product-service\.env"
$env:SEED_PRODUCTS = "$Products"
$env:SEED_REVIEWS = "$([Math]::Min($Products * 2, 50000))"
Push-Location "$root\services\product-service"
go run scripts/fake_data.go
Pop-Location

# Order
Load-Env "$root\services\order-service\.env"
$env:SEED_ORDERS = "$Orders"
$env:SEED_USERS = "$Users"
$env:SEED_PRODUCTS = "$Products"
$env:SEED_BATCH_SIZE = "$BatchSize"
Push-Location "$root\services\order-service"
go run scripts/fake_data.go
Pop-Location

Write-Host "=== Seed complete ===" -ForegroundColor Green
