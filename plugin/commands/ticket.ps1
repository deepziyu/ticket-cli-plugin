$ErrorActionPreference = 'Stop'

$ScriptDir = Split-Path -Parent $PSCommandPath
$BinaryName = "ticket-windows-amd64.exe"
$BinaryPath = Join-Path $ScriptDir $BinaryName
$Repo = "deepziyu/ticket-cli-plugin"

# Check for 0-byte corrupt file and remove it
if (Test-Path -LiteralPath $BinaryPath) {
    $file = Get-Item -LiteralPath $BinaryPath
    if ($file.Length -eq 0) {
        Write-Warning "[ticket] Found 0-byte corrupt binary. Removing..."
        Remove-Item -LiteralPath $BinaryPath -Force
    }
}

if (-not (Test-Path -LiteralPath $BinaryPath)) {
    Write-Host "[ticket] Binary not found. Downloading $BinaryName..." -ForegroundColor Yellow
    $TmpPath = "$BinaryPath.tmp"
    if (Test-Path -LiteralPath $TmpPath) { Remove-Item -LiteralPath $TmpPath -Force }

    $DownloadUrl = "https://github.com/$Repo/releases/latest/download/$BinaryName"
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TmpPath -UseBasicParsing

    if (-not (Test-Path -LiteralPath $TmpPath)) {
        throw "[ticket] Download failed (file not created)."
    }

    $tmpFile = Get-Item -LiteralPath $TmpPath
    if ($tmpFile.Length -eq 0) {
        Remove-Item -LiteralPath $TmpPath -Force
        throw "[ticket] Downloaded file is empty (404 or connection loss)."
    }

    Move-Item -Path $TmpPath -Destination $BinaryPath -Force
    Write-Host "[ticket] Download complete." -ForegroundColor Green
    
    # Brief sleep to allow antivirus/Defender scanning to finish and release file locks
    Start-Sleep -Seconds 1
}

if ($args) {
    $proc = Start-Process -FilePath $BinaryPath -ArgumentList $args -NoNewWindow -Wait -PassThru
} else {
    $proc = Start-Process -FilePath $BinaryPath -NoNewWindow -Wait -PassThru
}
exit $proc.ExitCode
