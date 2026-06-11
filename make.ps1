[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [ValidateSet("build", "build-linux", "build-windows", "build-darwin", "build-linux-amd64", "build-linux-arm64", "build-windows-amd64", "build-darwin-amd64", "build-darwin-arm64", "build-all", "package-windows-amd64", "package-darwin-arm64", "package-darwin-amd64", "package-linux-amd64", "package-all", "clean", "help")]
    [string]$Target = "build"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$Script:RootDir = Split-Path -Parent $PSCommandPath
$Script:CommandDir = Join-Path -Path $Script:RootDir -ChildPath "plugin/commands"

function Ensure-CommandDirectory {
    if (-not (Test-Path -LiteralPath $Script:CommandDir)) {
        New-Item -ItemType Directory -Path $Script:CommandDir | Out-Null
    }
}

function Invoke-Build {
    param(
        [string]$GoOS = $null,
        [string]$GoArch = $null,
        [string]$OutFileName = $null
    )
    Write-Host "🚀 Building ticket-cli-plugin..." -ForegroundColor Cyan
    Ensure-CommandDirectory

    # Backup env vars
    $oldOS = $env:GOOS
    $oldArch = $env:GOARCH

    try {
        if ($GoOS) { $env:GOOS = $GoOS }
        if ($GoArch) { $env:GOARCH = $GoArch }

        # Determine target executable name and suffix
        $targetOS = if ($GoOS) { $GoOS } else { 
            if ([System.Runtime.InteropServices.RuntimeInformation]::IsOSPlatform([System.Runtime.InteropServices.OSPlatform]::Windows)) { "windows" } else { "linux" }
        }
        
        $exeSuffix = ""
        if ($targetOS -eq "windows") {
            $exeSuffix = ".exe"
        }

        $finalFileName = if ($OutFileName) { $OutFileName } else { "ticket$exeSuffix" }
        $outputPath = Join-Path -Path $Script:CommandDir -ChildPath $finalFileName

        Write-Host "   Target OS/Arch: $(if ($GoOS){$GoOS}else{'host'})/$(if ($GoArch){$GoArch}else{'host'})" -ForegroundColor Gray
        Write-Host "   Output Path: $outputPath" -ForegroundColor Gray

        # Build binary
        & go build -ldflags "-s -w" -o $outputPath main.go
        if ($LASTEXITCODE -ne 0) {
            throw "Go build failed!"
        }

        Write-Host "✨ Build complete! Binary written to: $outputPath" -ForegroundColor Green
    }
    finally {
        # Restore env vars
        $env:GOOS = $oldOS
        $env:GOARCH = $oldArch
    }
}

function Invoke-Package {
    param(
        [string]$GoOS,
        [string]$GoArch,
        [string]$PlatformName,
        [string]$OutFormat = "zip"
    )
    Write-Host "📦 Packaging plugin for $GoOS-$GoArch..." -ForegroundColor Cyan
    
    $distDir = Join-Path -Path $Script:RootDir -ChildPath "dist"
    $tempDir = Join-Path -Path $distDir -ChildPath "temp"
    $tempPluginDir = Join-Path -Path $tempDir -ChildPath "plugin"

    # Clean old temp/dist files
    if (Test-Path -LiteralPath $tempDir) {
        Remove-Item -LiteralPath $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
    New-Item -ItemType Directory -Path $tempPluginDir -Force | Out-Null

    # Copy files
    Copy-Item -Path (Join-Path -Path $Script:RootDir -ChildPath "plugin/*") -Destination $tempPluginDir -Recurse -Force

    # Clean the commands dir in the temp copy
    $tempCommandsDir = Join-Path -Path $tempPluginDir -ChildPath "commands"
    if (Test-Path -LiteralPath $tempCommandsDir) {
        Remove-Item -LiteralPath $tempCommandsDir -Recurse -Force -ErrorAction SilentlyContinue
    }
    New-Item -ItemType Directory -Path $tempCommandsDir -Force | Out-Null

    # Compile the target binary inside the temp commands dir
    $exeSuffix = ""
    if ($GoOS -eq "windows") {
        $exeSuffix = ".exe"
    }
    $targetBinaryName = "ticket$exeSuffix"
    $targetOutputPath = Join-Path -Path $tempCommandsDir -ChildPath $targetBinaryName

    Write-Host "   Compiling binary for $GoOS-$GoArch..." -ForegroundColor Gray
    
    # Backup env vars
    $oldOS = $env:GOOS
    $oldArch = $env:GOARCH
    try {
        $env:GOOS = $GoOS
        $env:GOARCH = $GoArch
        & go build -ldflags "-s -w" -o $targetOutputPath main.go
        if ($LASTEXITCODE -ne 0) {
            throw "Go build failed during packaging!"
        }
    }
    finally {
        $env:GOOS = $oldOS
        $env:GOARCH = $oldArch
    }

    # Compress the temp/plugin directory
    $packageName = "ticket-management-plugin-$PlatformName"
    if ($OutFormat -eq "zip") {
        $archivePath = Join-Path -Path $distDir -ChildPath "$packageName.zip"
        if (Test-Path -LiteralPath $archivePath) { Remove-Item -LiteralPath $archivePath -Force }
        Write-Host "   Creating zip archive: $archivePath" -ForegroundColor Gray
        Compress-Archive -Path "$tempPluginDir" -DestinationPath $archivePath -Force
    } else {
        $archivePath = Join-Path -Path $distDir -ChildPath "$packageName.tar.gz"
        if (Test-Path -LiteralPath $archivePath) { Remove-Item -LiteralPath $archivePath -Force }
        Write-Host "   Creating tar.gz archive: $archivePath" -ForegroundColor Gray
        
        # Using native tar which is available on Windows 10+ / macOS / Linux
        Push-Location $tempDir
        try {
            & tar -czf $archivePath plugin
            if ($LASTEXITCODE -ne 0) {
                throw "tar failed!"
            }
        }
        finally {
            Pop-Location
        }
    }

    # Clean up temp
    Remove-Item -LiteralPath $tempDir -Recurse -Force -ErrorAction SilentlyContinue

    Write-Host "✨ Packaging complete! Archive created at: $archivePath" -ForegroundColor Green
}

function Invoke-Clean {
    Write-Host "🧹 Cleaning built binaries and packages..." -ForegroundColor Cyan
    if (Test-Path -LiteralPath $Script:CommandDir) {
        Remove-Item -LiteralPath $Script:CommandDir -Recurse -Force -ErrorAction SilentlyContinue
    }
    $distDir = Join-Path -Path $Script:RootDir -ChildPath "dist"
    if (Test-Path -LiteralPath $distDir) {
        Remove-Item -LiteralPath $distDir -Recurse -Force -ErrorAction SilentlyContinue
    }
    Write-Host "✨ Clean complete!" -ForegroundColor Green
}

function Show-Help {
    Write-Host "Usage: ./make.ps1 <target>"
    Write-Host ""
    Write-Host "Targets:"
    Write-Host "  build                  Build the binary for the host OS and place it in plugin/commands/"
    Write-Host "  build-linux            Build the Linux amd64 binary as 'ticket' in plugin/commands/"
    Write-Host "  build-windows          Build the Windows amd64 binary as 'ticket.exe' in plugin/commands/"
    Write-Host "  build-darwin           Build the macOS arm64 (M1/M2/M3) binary as 'ticket' in plugin/commands/"
    Write-Host "  build-linux-amd64      Build Linux amd64 binary with arch suffix"
    Write-Host "  build-linux-arm64      Build Linux arm64 binary with arch suffix"
    Write-Host "  build-windows-amd64    Build Windows amd64 binary with arch suffix"
    Write-Host "  build-darwin-amd64     Build macOS amd64 binary with arch suffix"
    Write-Host "  build-darwin-arm64     Build macOS arm64 binary with arch suffix"
    Write-Host "  build-all              Build all platforms with arch suffixes"
    Write-Host "  package-windows-amd64  Build and package plugin as ZIP for Windows amd64"
    Write-Host "  package-darwin-arm64    Build and package plugin as ZIP for macOS arm64 (M1/M2/M3)"
    Write-Host "  package-darwin-amd64    Build and package plugin as ZIP for macOS amd64 (Intel)"
    Write-Host "  package-linux-amd64     Build and package plugin as tar.gz for Linux amd64"
    Write-Host "  package-all             Build and package for all platforms"
    Write-Host "  clean                  Remove built binaries and packages"
    Write-Host "  help                   Show this help message"
}

try {
    Push-Location -LiteralPath $Script:RootDir

    switch ($Target) {
        "build" { Invoke-Build }
        "build-linux" { Invoke-Build -GoOS "linux" -GoArch "amd64" -OutFileName "ticket" }
        "build-windows" { Invoke-Build -GoOS "windows" -GoArch "amd64" -OutFileName "ticket.exe" }
        "build-darwin" { Invoke-Build -GoOS "darwin" -GoArch "arm64" -OutFileName "ticket" }
        "build-linux-amd64" { Invoke-Build -GoOS "linux" -GoArch "amd64" -OutFileName "ticket-linux-amd64" }
        "build-linux-arm64" { Invoke-Build -GoOS "linux" -GoArch "arm64" -OutFileName "ticket-linux-arm64" }
        "build-windows-amd64" { Invoke-Build -GoOS "windows" -GoArch "amd64" -OutFileName "ticket-windows-amd64.exe" }
        "build-darwin-amd64" { Invoke-Build -GoOS "darwin" -GoArch "amd64" -OutFileName "ticket-darwin-amd64" }
        "build-darwin-arm64" { Invoke-Build -GoOS "darwin" -GoArch "arm64" -OutFileName "ticket-darwin-arm64" }
        "build-all" {
            Invoke-Build -GoOS "linux" -GoArch "amd64" -OutFileName "ticket-linux-amd64"
            Invoke-Build -GoOS "linux" -GoArch "arm64" -OutFileName "ticket-linux-arm64"
            Invoke-Build -GoOS "windows" -GoArch "amd64" -OutFileName "ticket-windows-amd64.exe"
            Invoke-Build -GoOS "darwin" -GoArch "amd64" -OutFileName "ticket-darwin-amd64"
            Invoke-Build -GoOS "darwin" -GoArch "arm64" -OutFileName "ticket-darwin-arm64"
        }
        "package-windows-amd64" { Invoke-Package -GoOS "windows" -GoArch "amd64" -PlatformName "windows-amd64" -OutFormat "zip" }
        "package-darwin-arm64"   { Invoke-Package -GoOS "darwin"  -GoArch "arm64" -PlatformName "darwin-arm64"   -OutFormat "zip" }
        "package-darwin-amd64"   { Invoke-Package -GoOS "darwin"  -GoArch "amd64" -PlatformName "darwin-amd64"   -OutFormat "zip" }
        "package-linux-amd64"    { Invoke-Package -GoOS "linux"   -GoArch "amd64" -PlatformName "linux-amd64"    -OutFormat "tar.gz" }
        "package-all" {
            Invoke-Package -GoOS "windows" -GoArch "amd64" -PlatformName "windows-amd64" -OutFormat "zip"
            Invoke-Package -GoOS "darwin"  -GoArch "arm64" -PlatformName "darwin-arm64"   -OutFormat "zip"
            Invoke-Package -GoOS "darwin"  -GoArch "amd64" -PlatformName "darwin-amd64"   -OutFormat "zip"
            Invoke-Package -GoOS "linux"   -GoArch "amd64" -PlatformName "linux-amd64"    -OutFormat "tar.gz"
        }
        "clean" { Invoke-Clean }
        "help"  { Show-Help }
    }
}
catch {
    Write-Error -Message $_.Exception.Message
    exit 1
}
finally {
    Pop-Location
}
