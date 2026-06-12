$ErrorActionPreference = 'Stop'

$ScriptDir = Split-Path -Parent $PSCommandPath
$BinaryName = "ticket-windows-amd64.exe"
$BinaryPath = Join-Path $ScriptDir $BinaryName
$Repo = "deepziyu/ticket-cli-plugin"

if (-not (Test-Path -LiteralPath $BinaryPath)) {
    Write-Host "[ticket] Binary not found. Downloading $BinaryName..." -ForegroundColor Yellow
    $DownloadUrl = "https://github.com/$Repo/releases/latest/download/$BinaryName"
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $BinaryPath -UseBasicParsing
    Write-Host "[ticket] Download complete." -ForegroundColor Green
}

$proc = Start-Process -FilePath $BinaryPath -ArgumentList $args -NoNewWindow -Wait -PassThru
exit $proc.ExitCode
