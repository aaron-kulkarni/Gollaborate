@echo off
setlocal enabledelayedexpansion

set APP_NAME=gollaborate.exe

echo Building Gollaborate (Decentralized Collaborative Editor)...

REM Clean previous builds
echo Cleaning previous builds...
if exist %APP_NAME% del %APP_NAME%

REM Build main executable
echo Building executable...
go build -o %APP_NAME% peer\main.go

if errorlevel 1 (
    echo Build failed!
    exit /b 1
)

echo Build complete!
echo Executable: %APP_NAME%

echo.
echo To run (listening on port 49874):
echo   %APP_NAME% -listen 127.0.0.1:49874
echo.
echo To connect to other instances:
echo   %APP_NAME% -listen 127.0.0.1:49875 -peers 127.0.0.1:49874
echo.
echo You can open multiple terminals and run on different ports, connecting them as desired.
echo.
echo To run all integration/unit tests:
echo   go test -v
echo.
echo Happy collaborating!

endlocal
