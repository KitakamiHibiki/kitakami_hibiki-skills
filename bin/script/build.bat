@echo off
setlocal enabledelayedexpansion

set ROOT=%~dp0..
set OUT=%ROOT%\build

:: --- version info ---
for /f %%i in ('git describe --tags --always --dirty 2^>nul') do set VERSION=%%i
if "%VERSION%"=="" set VERSION=v0.0.0

for /f %%i in ('git rev-parse --short HEAD 2^>nul') do set COMMIT=%%i
if "%COMMIT%"=="" set COMMIT=unknown

:: Strip leading 'v' from version for the Go var
set LDVERSION=%VERSION%
if "%LDVERSION:~0,1%"=="v" set LDVERSION=%LDVERSION:~1%

set DATE=%DATE:/=-%
set LDFLAGS=-X skills/bin/internal/version.Version=%LDVERSION% -X skills/bin/internal/version.Commit=%COMMIT% -X skills/bin/internal/version.Date=%DATE%

echo === Building skills ===
echo   Version: %VERSION%
echo   Commit:  %COMMIT%
echo   Date:    %DATE%
echo.

echo [clean] Removing old build artifacts...
if exist "%OUT%" (
    rmdir /s /q "%OUT%"
)
mkdir "%OUT%"
echo OK

echo [fetch] go mod tidy...
cd /d "%ROOT%"
go mod tidy
if %ERRORLEVEL% neq 0 exit /b %ERRORLEVEL%
echo OK

:: Build each subcommand (each directory with a main.go).
for /d %%D in ("%ROOT%\*") do (
    set NAME=%%~nD
    if exist "%%D\main.go" (
        echo [build] !NAME! (windows amd64^)...
        set GOOS=windows
        set GOARCH=amd64
        go build -ldflags "!LDFLAGS!" -o "%OUT%\!NAME!.exe" "%%D"
        if !ERRORLEVEL! neq 0 exit /b !ERRORLEVEL!
        echo   OK - %OUT%\!NAME!.exe

        echo [build] !NAME! (linux amd64^)...
        set GOOS=linux
        set GOARCH=amd64
        go build -ldflags "!LDFLAGS!" -o "%OUT%\!NAME!" "%%D"
        if !ERRORLEVEL! neq 0 exit /b !ERRORLEVEL!
        echo   OK - %OUT%\!NAME!
    )
)

echo.
echo Done. Outputs:
dir "%OUT%"
