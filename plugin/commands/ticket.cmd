@echo off
setlocal

set "SCRIPT_DIR=%~dp0"
set "BINARY_NAME=ticket-windows-amd64.exe"
set "BINARY_PATH=%SCRIPT_DIR%%BINARY_NAME%"
set "REPO=deepziyu/ticket-cli-plugin"

if exist "%BINARY_PATH%" goto :run

echo [ticket] Binary not found. Downloading %BINARY_NAME%... 1>&2
set "DOWNLOAD_URL=https://github.com/%REPO%/releases/latest/download/%BINARY_NAME%"

where curl >nul 2>&1
if %errorlevel%==0 (
    curl -fsSL "%DOWNLOAD_URL%" -o "%BINARY_PATH%"
) else (
    powershell -NoProfile -Command "Invoke-WebRequest -Uri '%DOWNLOAD_URL%' -OutFile '%BINARY_PATH%' -UseBasicParsing"
)

if not exist "%BINARY_PATH%" (
    echo [ticket] Download failed. 1>&2
    exit /b 1
)

echo [ticket] Download complete. 1>&2

:run
"%BINARY_PATH%" %*
exit /b %errorlevel%
