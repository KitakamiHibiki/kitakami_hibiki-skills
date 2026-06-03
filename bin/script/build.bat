@echo off
setlocal enabledelayedexpansion

set ROOT=%~dp0..
set OUT=%ROOT%\build
set APP=skills-cli

echo === Building %APP% ===
echo.

echo [clean] Removing old build artifacts...
if exist "%OUT%" (
    del /f /q "%OUT%\%APP%.exe" 2>nul
    del /f /q "%OUT%\%APP%" 2>nul
    echo OK
) else (
    mkdir "%OUT%"
)

echo [1/3] Fetching dependencies...
cd /d "%ROOT%"
go mod tidy
if %ERRORLEVEL% neq 0 (
    echo FAILED
    exit /b %ERRORLEVEL%
)
echo OK

echo [2/3] Building Windows amd64...
set GOOS=windows
set GOARCH=amd64
go build -o "%OUT%\%APP%.exe" "%ROOT%\main.go"
if %ERRORLEVEL% neq 0 (
    echo FAILED
    exit /b %ERRORLEVEL%
)
echo OK - %OUT%\%APP%.exe

echo [3/3] Building Linux amd64...
set GOOS=linux
set GOARCH=amd64
go build -o "%OUT%\%APP%" "%ROOT%\main.go"
if %ERRORLEVEL% neq 0 (
    echo FAILED
    exit /b %ERRORLEVEL%
)
echo OK - %OUT%\%APP%

echo.
echo Done. Outputs:
dir "%OUT%"
