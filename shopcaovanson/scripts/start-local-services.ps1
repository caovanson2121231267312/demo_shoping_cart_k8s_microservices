# shopcaovanson - Start Go microservices locally (infra must already be running)
# Usage:
#   .\scripts\start-local-services.ps1              # mac dinh: Go tren Windows
#   .\scripts\start-local-services.ps1 -Background  # chay nen, ghi log vao logs/
#   .\scripts\start-local-services.ps1 -Wsl         # chay Go trong WSL (tranh Smart App Control)
param(
    [switch]$Background,
    [switch]$Wsl,
    [switch]$Windows
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
$LogDir = Join-Path $Root "logs"
$GoServiceLauncher = Join-Path $PSScriptRoot "run-go-service.ps1"
$GoMigrateLauncher = Join-Path $PSScriptRoot "run-go-migrate.ps1"
. (Join-Path $PSScriptRoot "lib\wsl-go.ps1")
foreach ($d in @($LogDir)) {
    if (-not (Test-Path $d)) {
        New-Item -ItemType Directory -Path $d | Out-Null
    }
}

$WslDistro = $null
$UseWsl = $false
if ($Windows -and $Wsl) {
    throw "Khong the dung dong thoi -Wsl va -Windows"
}
if ($Wsl) {
    $WslDistro = Find-WslDistroWithGo
    if (-not $WslDistro) {
        throw "Wsl duoc yeu cau nhung chua co Go trong WSL. Chay: .\scripts\setup-wsl-go.ps1"
    }
    $UseWsl = $true
}

if ($UseWsl) {
    Write-Host "Go services: WSL ($WslDistro) - tranh Windows App Control" -ForegroundColor Cyan
} else {
    Write-Host "Go services: Windows (go run)" -ForegroundColor DarkGray
    if (Test-SmartAppControlOn) {
        Write-Host "  CANH BAO: Smart App Control dang bat - neu bi chan .exe, dung -Wsl hoac tat SAC" -ForegroundColor Yellow
    }
    Write-Host ""
}

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

function Ensure-EnvFile($dir) {
    $envFile = Join-Path $dir ".env"
    $example = Join-Path $dir ".env.example"
    if (-not (Test-Path $envFile) -and (Test-Path $example)) {
        Copy-Item $example $envFile
        Write-Host "  + Tao .env tu .env.example: $dir" -ForegroundColor DarkGray
    }
}

function Ensure-PythonDeps($dir) {
    $req = Join-Path $dir "requirements.txt"
    if (-not (Test-Path $req)) { return }
    $name = Split-Path $dir -Leaf
    $prevEap = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    try {
        & pip install -q httpx python-dotenv 2>&1 | Out-Null
        & pip install -q -r $req 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) {
            Write-Host "  WARN pip install that bai: $name (thu: pip install -r services\$name\requirements.txt)" -ForegroundColor Yellow
        }
    } catch {
        Write-Host "  WARN pip install: $name - $_" -ForegroundColor Yellow
    } finally {
        $ErrorActionPreference = $prevEap
    }
}

function Stop-PortListener($Port) {
    $pids = @(Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue |
        Select-Object -ExpandProperty OwningProcess -Unique |
        Where-Object { $_ -and $_ -ne 0 })
    foreach ($procId in $pids) {
        Stop-Process -Id $procId -Force -ErrorAction SilentlyContinue
    }
    if ($pids.Count -gt 0) {
        Start-Sleep -Seconds 1
    }
}

function Start-ServiceProcess($name, $dir, $runCommand) {
    $logFile = Join-Path $LogDir "$name.log"
    if (-not $Background) {
        $inner = "Set-Location '$dir'; Write-Host '[$name] Ctrl+C de dung' -ForegroundColor Cyan; $runCommand"
        Start-Process powershell -ArgumentList @("-NoExit", "-NoProfile", "-Command", $inner) -WindowStyle Normal | Out-Null
        Write-Host "  + terminal: $name" -ForegroundColor DarkGray
        return $true
    }

    $stamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $header = "`n===== $name started $stamp =====`n"
    Add-Content -Path $logFile -Value $header -Encoding utf8
    $inner = @"
Set-Location '$dir'
`$ErrorActionPreference = 'Continue'
& { $runCommand } *>> '$logFile'
"@
    Start-Process powershell -ArgumentList @("-NoProfile", "-WindowStyle", "Hidden", "-Command", $inner) -WindowStyle Hidden | Out-Null
    Write-Host "  + log: logs\$name.log" -ForegroundColor DarkGray
    return $true
}

function Start-GoService($dir, $name) {
    $wslArg = if ($UseWsl) { "-UseWsl -WslDistro '$WslDistro'" } else { "" }
    $runCommand = "& '$GoServiceLauncher' -Name '$name' -Dir '$dir' $wslArg"
    try {
        return Start-ServiceProcess $name $dir $runCommand
    } catch {
        Write-Host "  FAIL start $name`: $_" -ForegroundColor Red
        return $false
    }
}

function Start-PythonService($dir, $name, $port) {
    $cmd = "python -m uvicorn main:app --host 0.0.0.0 --port $port"
    try {
        return Start-ServiceProcess $name $dir $cmd
    } catch {
        Write-Host "  FAIL start $name`: $_" -ForegroundColor Red
        return $false
    }
}

function Start-FrontendDev() {
    $frontendDir = Join-Path $Root "frontend\web"
    if (-not (Test-Path (Join-Path $frontendDir "package.json"))) { return }

    $port = 3000
    if (Test-PortOpen $port) {
        Write-Host "  skip frontend (port $port dang dung)" -ForegroundColor DarkGray
        return
    }

    $inner = @"
Set-Location '$frontendDir'
Write-Host '[frontend] npm run dev - http://localhost:3000' -ForegroundColor Cyan
npm run dev
"@
    Start-Process powershell -ArgumentList @("-NoExit", "-NoProfile", "-Command", $inner) -WindowStyle Normal | Out-Null
    Write-Host "  + terminal: frontend (http://localhost:3000)" -ForegroundColor DarkGray
}

function Test-PortOpen($Port) {
    try {
        $tcp = New-Object System.Net.Sockets.TcpClient
        $tcp.Connect("localhost", $Port)
        $tcp.Close()
        return $true
    } catch {
        return $false
    }
}

function Test-Health($Url) {
    try {
        $null = Invoke-RestMethod -Uri $Url -TimeoutSec 8
        return $true
    } catch {
        return $false
    }
}

function Wait-Health($Name, $Url, $MaxAttempts = 20) {
    for ($i = 1; $i -le $MaxAttempts; $i++) {
        if (Test-Health $Url) {
            Write-Host "  OK  $Name" -ForegroundColor Green
            return $true
        }
        Start-Sleep -Seconds 2
    }
    Write-Host "  FAIL $Name - $Url" -ForegroundColor Red
    return $false
}

function Run-GoMigration($dir, $name) {
    $migrateMain = Join-Path $dir "cmd\migrate\main.go"
    if (-not (Test-Path $migrateMain)) { return }
    Ensure-EnvFile $dir
    Load-EnvFile (Join-Path $dir ".env")
    if (-not $env:DATABASE_URL) { return }
    $prevEap = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    try {
        if ($UseWsl) {
            $out = & powershell -NoProfile -File $GoMigrateLauncher -Dir $dir -UseWsl -WslDistro $WslDistro 2>&1
        } else {
            $out = & powershell -NoProfile -File $GoMigrateLauncher -Dir $dir 2>&1
        }
        if ($LASTEXITCODE -ne 0) {
            Write-Host "  WARN migration $name`: $out" -ForegroundColor Yellow
        } else {
            Write-Host "  OK  migration $name" -ForegroundColor DarkGray
        }
    } catch {
        Write-Host "  WARN migration $name`: $_" -ForegroundColor Yellow
    } finally {
        $ErrorActionPreference = $prevEap
    }
}

$requiredPorts = @(
    @{ Port = 5432; Label = "auth postgres" },
    @{ Port = 5433; Label = "product postgres" },
    @{ Port = 5434; Label = "order postgres" },
    @{ Port = 6379; Label = "redis" },
    @{ Port = 9092; Label = "kafka" }
)

Write-Host "Checking infra..." -ForegroundColor Cyan
$infraOk = $true
foreach ($p in $requiredPorts) {
    if (Test-PortOpen $p.Port) {
        Write-Host "  OK  $($p.Label) (port $($p.Port))" -ForegroundColor Green
    } else {
        Write-Host "  --  $($p.Label) (port $($p.Port)) chua mo" -ForegroundColor Yellow
        $infraOk = $false
    }
}

if (-not $infraOk) {
    Write-Host ""
    Write-Host "Infra chua san sang. Chay truoc:" -ForegroundColor Yellow
    Write-Host '  .\scripts\start-local-infra.ps1' -ForegroundColor White
    Write-Host ""
}

$services = @(
    @{ Name = "auth-service";    Port = 8081; Health = "/health"; Dir = "$Root\services\auth-service"; Warmup = 8;  HealthAttempts = 30 },
    @{ Name = "product-service"; Port = 8082; Health = "/health"; Dir = "$Root\services\product-service"; Warmup = 8;  HealthAttempts = 25 },
    @{ Name = "order-service";   Port = 8083; Health = "/health"; Dir = "$Root\services\order-service"; Warmup = 4;  HealthAttempts = 20 },
    @{ Name = "chat-service";    Port = 8084; Health = "/health"; Dir = "$Root\services\chat-service"; Warmup = 4;  HealthAttempts = 20 },
    @{ Name = "rasa-service"; Port = 8090; Health = "/health"; Dir = "$Root\services\rasa-service"; Warmup = 4; HealthAttempts = 15; Python = $true },
    @{ Name = "notification-service"; Port = 8085; Health = "/health"; Dir = "$Root\services\notification-service"; Warmup = 5; HealthAttempts = 15; Python = $true },
    @{ Name = "analytics-service"; Port = 8086; Health = "/health"; Dir = "$Root\services\analytics-service"; Warmup = 5; HealthAttempts = 15; Python = $true },
    @{ Name = "api-gateway";     Port = 8080; Health = "/health"; Dir = "$Root\services\api-gateway"; Warmup = 3;  HealthAttempts = 20 }
)

Write-Host ""
if ($infraOk) {
    Write-Host "Running DB migrations..." -ForegroundColor Cyan
    Run-GoMigration "$Root\services\auth-service" "auth-service"
    Run-GoMigration "$Root\services\product-service" "product-service"
    Run-GoMigration "$Root\services\order-service" "order-service"
    Run-GoMigration "$Root\services\chat-service" "chat-service"
}

Write-Host ""
Write-Host "Stopping existing services..." -ForegroundColor Cyan
foreach ($svc in ($services | Sort-Object Port -Descending)) {
    if (Test-PortOpen $svc.Port) {
        Write-Host "  stop $($svc.Name) (port $($svc.Port))" -ForegroundColor DarkGray
        Stop-PortListener $svc.Port
    }
}

Write-Host ""
Write-Host "Starting services..." -ForegroundColor Cyan
foreach ($svc in $services) {
    Ensure-EnvFile $svc.Dir

    $envFile = Join-Path $svc.Dir ".env"
    Load-EnvFile $envFile
    Write-Host "  -> $($svc.Name) (port $($svc.Port))"
    if ($svc.Python) {
        Ensure-PythonDeps $svc.Dir
        if (-not (Start-PythonService $svc.Dir $svc.Name $svc.Port)) { continue }
    } else {
        if (-not (Start-GoService $svc.Dir $svc.Name)) { continue }
    }
    Start-Sleep -Seconds $svc.Warmup
}

Write-Host ""
Write-Host "Health checks..." -ForegroundColor Cyan
$allOk = $true
$failed = @()
foreach ($svc in $services) {
    $url = "http://localhost:$($svc.Port)$($svc.Health)"
    $ok = Wait-Health $svc.Name $url $svc.HealthAttempts
    if (-not $ok) { $failed += $svc.Name }
    $allOk = $ok -and $allOk
}

Write-Host ""
if ($allOk) {
    Write-Host "All services ready." -ForegroundColor Green
    if (-not $Background) {
        Write-Host ""
        Write-Host "Starting frontend..." -ForegroundColor Cyan
        Start-FrontendDev
    }
    Write-Host "  API Gateway:  http://localhost:8080" -ForegroundColor White
    Write-Host "  Auth:         http://localhost:8081" -ForegroundColor White
    Write-Host "  Product:      http://localhost:8082" -ForegroundColor White
    Write-Host "  Chatbot:      http://localhost:8090" -ForegroundColor White
    Write-Host "  Email (SMTP): notification-service :8085 - xem Mailtrap inbox" -ForegroundColor White
    Write-Host "  Login test:   admin@shop.com / Admin@123" -ForegroundColor White
    if (-not $Background) {
        Write-Host "  Frontend:     http://localhost:3000" -ForegroundColor White
    }
    if (-not $Background) {
        Write-Host "  Terminals:    moi service mot cua so PowerShell rieng" -ForegroundColor White
    } else {
        Write-Host "  Logs:         $LogDir" -ForegroundColor White
        Write-Host "  Xem log:      Get-Content logs\product-service.log -Wait" -ForegroundColor DarkGray
        Write-Host "  Terminal mode: .\scripts\start-local-services.ps1" -ForegroundColor DarkGray
    }
} else {
    Write-Host "Mot so service chua len: $($failed -join ', ')" -ForegroundColor Yellow
    if ($Background) {
        Write-Host "Xem log service loi:" -ForegroundColor Yellow
        foreach ($name in $failed) {
            Write-Host "  Get-Content logs\$name.log -Tail 40" -ForegroundColor White
        }
        Write-Host "Hoac chay khong -Background de mo terminal tung service:" -ForegroundColor Yellow
        Write-Host '  .\scripts\start-local-services.ps1' -ForegroundColor White
    }
    if ($failed -contains "auth-service") {
        Write-Host ""
        Write-Host "Login 500 (connection refused :8081) = auth-service chua chay." -ForegroundColor Red
        Write-Host "Chay thu trong terminal rieng de xem loi:" -ForegroundColor Yellow
        if ($UseWsl) {
            Write-Host '  .\scripts\run-go-service.ps1 -Name auth-service -Dir services\auth-service -UseWsl' -ForegroundColor White
        } else {
            Write-Host '  .\scripts\run-go-service.ps1 -Name auth-service -Dir services\auth-service' -ForegroundColor White
            Write-Host '  Hoac cai Go trong WSL roi chay lai script (tu dong dung WSL)' -ForegroundColor DarkGray
        }
        Write-Host ""
        Write-Host "Neu thieu bang DB, chay migration:" -ForegroundColor Yellow
        Write-Host '  cd services\auth-service; go run cmd/migrate/main.go up' -ForegroundColor White
        Write-Host '  go run scripts/fake_data.go' -ForegroundColor White
    } else {
        Write-Host "Kiem tra cua so service hoac chay thu service bi loi." -ForegroundColor Yellow
        foreach ($svcName in @('product-service', 'api-gateway', 'order-service', 'chat-service')) {
            if ($failed -contains $svcName) {
                $w = if ($UseWsl) { ' -UseWsl' } else { '' }
                Write-Host "  .\scripts\run-go-service.ps1 -Name $svcName -Dir services\$svcName$w" -ForegroundColor White
            }
        }
        Write-Host ""
        Write-Host "Neu gap Application Control policy has blocked:" -ForegroundColor Yellow
        Write-Host "  .\scripts\setup-wsl-go.ps1" -ForegroundColor White
        Write-Host "  .\scripts\start-local-services.ps1" -ForegroundColor White
    }
}
