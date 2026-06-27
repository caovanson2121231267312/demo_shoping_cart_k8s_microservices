# Start Rasa action server + trained model server (Docker).
$ErrorActionPreference = "Stop"
$Root = Split-Path $PSScriptRoot -Parent
$RasaDir = Join-Path $Root "services\rasa-service"
$ModelsDir = Join-Path $RasaDir "rasa\models"

function Get-LatestRasaModel {
    $latestFile = Join-Path $ModelsDir "latest.txt"
    if (Test-Path $latestFile) {
        $name = (Get-Content $latestFile -Raw).Trim()
        $path = Join-Path $ModelsDir $name
        if ($name -and (Test-Path $path)) {
            return $path
        }
    }
    $newest = Get-ChildItem -Path $ModelsDir -Filter "chatbot-*.tar.gz" -ErrorAction SilentlyContinue |
        Sort-Object LastWriteTime -Descending |
        Select-Object -First 1
    if ($newest) { return $newest.FullName }
    $legacy = Join-Path $ModelsDir "chatbot.tar.gz"
    if (Test-Path $legacy) { return $legacy }
    return $null
}

Push-Location $RasaDir
try {
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        Write-Warning "Docker khong co - rasa-service se fallback bot_engine (USE_RASA=true van thu Rasa truoc)."
        return
    }

    $modelPath = Get-LatestRasaModel
    if (-not $modelPath) {
        Write-Host "Chua co model - train truoc..." -ForegroundColor Yellow
        & (Join-Path $PSScriptRoot "train-rasa.ps1")
    } else {
        Write-Host "Latest model: $(Split-Path $modelPath -Leaf)" -ForegroundColor DarkGray
    }

    Write-Host "Starting Rasa stack (actions + model server)..." -ForegroundColor Cyan
    docker compose up -d --build rasa-actions rasa

    $deadline = (Get-Date).AddMinutes(3)
    while ((Get-Date) -lt $deadline) {
        try {
            $r = Invoke-WebRequest -Uri "http://localhost:5005/status" -UseBasicParsing -TimeoutSec 3
            if ($r.StatusCode -eq 200) {
                Write-Host "Rasa ready: http://localhost:5005" -ForegroundColor Green
                return
            }
        } catch {}
        Start-Sleep -Seconds 3
    }
    Write-Warning "Rasa chua san sang trong 3 phut - xem log: docker compose logs -f rasa"
} finally {
    Pop-Location
}
