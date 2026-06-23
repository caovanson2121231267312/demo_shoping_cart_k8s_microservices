# Shared helpers: WSL Go runner + Smart App Control detection

function Test-SmartAppControlOn {
    try {
        return (Get-MpComputerStatus).SmartAppControlState -eq 'On'
    } catch {
        return $false
    }
}

function Get-WslDistroNames {
    if (-not (Get-Command wsl -ErrorAction SilentlyContinue)) { return @() }
    $prevEap = $ErrorActionPreference
    $ErrorActionPreference = 'SilentlyContinue'
    try {
        $raw = & wsl -l -q 2>$null
        if (-not $raw) { return @() }
        return @($raw | ForEach-Object { ($_ -replace '\x00', '').Trim() } | Where-Object { $_ })
    } finally {
        $ErrorActionPreference = $prevEap
    }
}

function Invoke-WslWithTimeout {
    param(
        [Parameter(Mandatory)][string]$Distro,
        [Parameter(Mandatory)][string[]]$WslArgs,
        [int]$TimeoutSec = 8
    )
    if (-not (Get-Command wsl -ErrorAction SilentlyContinue)) { return $null }

    $argList = @('-d', $Distro) + $WslArgs
    $psi = New-Object System.Diagnostics.ProcessStartInfo
    $psi.FileName = 'wsl'
    $psi.Arguments = ($argList | ForEach-Object {
        if ($_ -match '\s') { '"' + ($_ -replace '"', '\"') + '"' } else { $_ }
    }) -join ' '
    $psi.RedirectStandardOutput = $true
    $psi.RedirectStandardError = $true
    $psi.UseShellExecute = $false
    $psi.CreateNoWindow = $true

    $proc = [System.Diagnostics.Process]::Start($psi)
    if (-not $proc.WaitForExit($TimeoutSec * 1000)) {
        try { $proc.Kill() } catch { }
        return $null
    }
    return $proc.ExitCode
}

function Test-WslDistroHasGo([string]$Distro) {
    if (-not $Distro) { return $false }
    $code = Invoke-WslWithTimeout -Distro $Distro -WslArgs @('go', 'version') -TimeoutSec 8
    return $code -eq 0
}

function Find-WslDistroWithGo {
    $distros = Get-WslDistroNames
    if ($distros.Count -eq 0) { return $null }

    $ranked = $distros | Sort-Object {
        if ($_ -match '(?i)ubuntu') { 0 }
        elseif ($_ -match '(?i)debian') { 1 }
        elseif ($_ -match '(?i)docker') { 9 }
        else { 5 }
    }

    foreach ($d in $ranked) {
        if (Test-WslDistroHasGo $d) { return $d }
    }
    return $null
}

function Write-SmartAppControlHelp([string]$ServiceDir = '') {
    $setupScript = Join-Path (Split-Path $PSScriptRoot -Parent) 'setup-wsl-go.ps1'

    Write-Host ""
    Write-Host "Windows Smart App Control dang chan binary Go (.exe)." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Cach 1 (khuyen nghi): Cai Ubuntu + Go trong WSL:" -ForegroundColor White
    Write-Host "  .\scripts\setup-wsl-go.ps1" -ForegroundColor DarkGray
    Write-Host "  (Neu moi cai Ubuntu: mo ung dung Ubuntu mot lan de tao user)" -ForegroundColor DarkGray
    Write-Host "  .\scripts\start-local-services.ps1" -ForegroundColor DarkGray
    Write-Host ""
    Write-Host "Cach 2: Tat Smart App Control" -ForegroundColor White
    Write-Host "  Cai dat > Quyen rieng tu va bao mat > Bao mat Windows" -ForegroundColor DarkGray
    Write-Host "  > Kiem soat ung dung va trinh duyet > Cai dat Kiem soat Ung dung Thong minh > Tat" -ForegroundColor DarkGray
    Write-Host ""
    if ($ServiceDir) {
        Write-Host "Cach 3: Chay thu trong WSL (sau khi cai Go):" -ForegroundColor White
        Write-Host "  wsl -d Ubuntu --cd `"$ServiceDir`" go run ." -ForegroundColor DarkGray
        Write-Host ""
    }
}

function Invoke-WslGo {
    param(
        [Parameter(Mandatory)][string]$Distro,
        [Parameter(Mandatory)][string]$Dir,
        [Parameter(Mandatory)][string[]]$GoArgs
    )
    $resolvedDir = (Resolve-Path $Dir).Path
    & wsl -d $Distro --cd $resolvedDir go @GoArgs
    return $LASTEXITCODE
}
