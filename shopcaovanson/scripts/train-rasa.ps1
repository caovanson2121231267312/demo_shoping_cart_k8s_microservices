# Train Rasa NLU model (Docker) and optionally start the stack.
# Each run saves a timestamped model: rasa/models/chatbot-YYYYMMDD-HHMMSS.tar.gz
param(
    [switch]$Start,
    [switch]$Force
)

$ErrorActionPreference = "Stop"
$Root = Split-Path $PSScriptRoot -Parent
$RasaDir = Join-Path $Root "services\rasa-service"
$ModelsDir = Join-Path $RasaDir "rasa\models"

Push-Location $RasaDir
try {
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        Write-Error "Docker chua cai. Cai Docker Desktop roi chay lai script nay."
    }

    $env:FORCE_RASA_TRAIN = if ($Force) { "true" } else { "false" }

    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $modelName = "chatbot-$timestamp"
    $modelFile = "$modelName.tar.gz"

    Write-Host "Building rasa-actions..." -ForegroundColor Cyan
    docker compose build rasa-actions

    Write-Host "Starting rasa-actions..." -ForegroundColor Cyan
    docker compose up -d rasa-actions

    Write-Host "Training Rasa model ($modelFile)..." -ForegroundColor Cyan
    docker compose run --rm --entrypoint "" rasa rasa train `
        --fixed-model-name $modelName `
        --data data `
        --config config.yml `
        --domain domain.yml `
        --out models

    if (-not (Test-Path (Join-Path $ModelsDir $modelFile))) {
        Write-Error "Train xong nhung khong thay models\$modelFile"
    }

    Set-Content -Path (Join-Path $ModelsDir "latest.txt") -Value $modelFile -NoNewline -Encoding ascii

    Write-Host ""
    Write-Host "Model saved:" -ForegroundColor Green
    Write-Host "  services/rasa-service/rasa/models/$modelFile" -ForegroundColor White
    Write-Host "  latest -> $modelFile" -ForegroundColor DarkGray

    $allModels = @(Get-ChildItem -Path $ModelsDir -Filter "chatbot-*.tar.gz" -ErrorAction SilentlyContinue |
        Sort-Object LastWriteTime -Descending)
    if ($allModels.Count -gt 1) {
        Write-Host "  (${allModels.Count} models in folder)" -ForegroundColor DarkGray
    }

    if ($Start) {
        Write-Host ""
        Write-Host "Starting Rasa server (load $modelFile)..." -ForegroundColor Cyan
        docker compose up -d rasa
        Write-Host "Rasa REST: http://localhost:5005" -ForegroundColor Green
    } else {
        Write-Host "Start server: .\scripts\train-rasa.ps1 -Start  (or docker compose up -d rasa)" -ForegroundColor White
    }
} finally {
    Pop-Location
}
