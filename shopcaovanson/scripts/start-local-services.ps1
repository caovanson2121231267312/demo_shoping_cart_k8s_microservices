# shopcaovanson — Start Go microservices locally (infra must already be running)
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot

function Load-EnvFile($path) {
    if (-not (Test-Path $path)) { return }
    Get-Content $path | ForEach-Object {
        if ($_ -match '^\s*([^#=]+)=(.*)$') {
            $name = $matches[1].Trim()
            $value = $matches[2].Trim().Trim('"')
            Set-Item -Path "env:$name" -Value $value
        }
    }
}

$services = @(
    @{ Name = "product-service"; Port = 8082; Dir = "$Root\services\product-service" },
    @{ Name = "order-service";   Port = 8083; Dir = "$Root\services\order-service" },
    @{ Name = "chat-service";   Port = 8084; Dir = "$Root\services\chat-service" },
    @{ Name = "api-gateway";    Port = 8080; Dir = "$Root\services\api-gateway" }
)

foreach ($svc in $services) {
    $envFile = Join-Path $svc.Dir ".env"
    Load-EnvFile $envFile
    Write-Host "Starting $($svc.Name) on port $($svc.Port)..."
    Start-Process -FilePath "go" -ArgumentList "run", "main.go" -WorkingDirectory $svc.Dir -WindowStyle Minimized
    Start-Sleep -Seconds 2
}

Write-Host ""
Write-Host "Services starting. API Gateway: http://localhost:8080"
Write-Host "Health checks:"
Write-Host "  curl http://localhost:8080/health"
Write-Host "  curl http://localhost:8082/health"
Write-Host "  curl http://localhost:8083/health"
Write-Host "  curl http://localhost:8084/health"
