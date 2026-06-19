# shopcaovanson — Start Docker infra for local Go services (Postgres, Redis, Mongo, ES, Kafka)

$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot



function Test-Docker {

    try {

        docker info 2>$null | Out-Null

        return $true

    } catch {

        return $false

    }

}



function Wait-Port($HostName, $Port, $Label, $TimeoutSec = 60) {

    $deadline = (Get-Date).AddSeconds($TimeoutSec)

    while ((Get-Date) -lt $deadline) {

        try {

            $tcp = New-Object System.Net.Sockets.TcpClient

            $tcp.Connect($HostName, $Port)

            $tcp.Close()

            Write-Host "  OK  $Label (port $Port)" -ForegroundColor Green

            return $true

        } catch {

            Start-Sleep -Seconds 2

        }

    }

    Write-Host "  FAIL $Label (port $Port) — timeout" -ForegroundColor Red

    return $false

}



if (-not (Test-Docker)) {

    Write-Host "Docker chua chay. Hay mo Docker Desktop roi chay lai script nay." -ForegroundColor Red

    exit 1

}



Write-Host "Starting infra containers..." -ForegroundColor Cyan



Push-Location "$Root\services\auth-service"

docker compose up -d postgres redis zookeeper kafka 2>&1 | Out-Host

Pop-Location



Push-Location "$Root\services\product-service"

docker compose up -d postgres mongo elasticsearch 2>&1 | Out-Host

Pop-Location



Push-Location "$Root\services\order-service"

docker compose up -d postgres 2>&1 | Out-Host

Pop-Location



Push-Location "$Root\services\chat-service"

if (Test-Path "docker-compose.yml") {

    docker compose up -d mongo redis 2>&1 | Out-Host

}

Pop-Location



Write-Host ""

Write-Host "Waiting for ports..." -ForegroundColor Cyan

$ok = $true

$ok = (Wait-Port localhost 5432 "auth postgres") -and $ok

$ok = (Wait-Port localhost 5433 "product postgres") -and $ok

$ok = (Wait-Port localhost 5434 "order postgres") -and $ok

$ok = (Wait-Port localhost 6379 "redis") -and $ok

$ok = (Wait-Port localhost 27017 "mongodb") -and $ok

$null = Wait-Port localhost 9200 "elasticsearch" 90

$null = Wait-Port localhost 9092 "kafka" 90



Write-Host ""

if ($ok) {

    Write-Host "Infra ready. Chay migrations (neu chua): LOCAL=true bash scripts/migrate-all.sh" -ForegroundColor Green

    Write-Host "Sau do: .\scripts\start-local-services.ps1" -ForegroundColor Green

} else {

    Write-Host "Mot so port chua san sang — kiem tra: docker ps" -ForegroundColor Yellow

}

