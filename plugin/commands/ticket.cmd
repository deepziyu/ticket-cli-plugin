@echo off
setlocal

set "SCRIPT_DIR=%~dp0"
set "BINARY_NAME=ticket-windows-amd64.exe"
set "BINARY_PATH=%SCRIPT_DIR%%BINARY_NAME%"
set "TMP_PATH=%BINARY_PATH%.tmp"
set "REPO=deepziyu/ticket-cli-plugin"
set "DOWNLOAD_URL=https://github.com/%REPO%/releases/latest/download/%BINARY_NAME%"

:: 1. Check for 0-byte corrupt file and remove it
powershell -NoProfile -Command "if (Test-Path -LiteralPath '%BINARY_PATH%') { if ((Get-Item -LiteralPath '%BINARY_PATH%').Length -eq 0) { Remove-Item -LiteralPath '%BINARY_PATH%' -Force } }" >nul 2>&1

if exist "%BINARY_PATH%" goto :run

echo [ticket] Binary not found. Downloading %BINARY_NAME%... 1>&2

:: Cleanup legacy tmp file
del /f /q "%TMP_PATH%" >nul 2>&1

where curl >nul 2>&1
if %errorlevel%==0 (
    curl -fsSL "%DOWNLOAD_URL%" -o "%TMP_PATH%"
) else (
    powershell -NoProfile -Command "Invoke-WebRequest -Uri '%DOWNLOAD_URL%' -OutFile '%TMP_PATH%' -UseBasicParsing"
)

:: 2. Check download command status
if %errorlevel% neq 0 (
    echo [ticket] Error: Download command failed. 1>&2
    del /f /q "%TMP_PATH%" >nul 2>&1
    exit /b 1
)

:: 3. Validate download size is non-zero
powershell -NoProfile -Command "if (-not (Test-Path -LiteralPath '%TMP_PATH%')) { exit 1 } if ((Get-Item -LiteralPath '%TMP_PATH%').Length -eq 0) { exit 1 }" >nul 2>&1
if %errorlevel% neq 0 (
    echo [ticket] Error: Downloaded file is empty or missing [404 or connection loss]. 1>&2
    del /f /q "%TMP_PATH%" >nul 2>&1
    exit /b 1
)

:: 4. Atomic rename and move
move /y "%TMP_PATH%" "%BINARY_PATH%" >nul 2>&1
if %errorlevel% neq 0 (
    echo [ticket] Error: Failed to rename temporary download file. 1>&2
    del /f /q "%TMP_PATH%" >nul 2>&1
    exit /b 1
)

echo [ticket] Download complete. 1>&2
:: Brief sleep to allow antivirus/Defender scanning to finish and release file locks
ping 127.0.0.1 -n 2 >nul 2>&1

:run
"%BINARY_PATH%" %*
exit /b %errorlevel%
