@echo off
chcp 65001 >nul
setlocal EnableExtensions

set "WSL_DISTRO=Ubuntu-24.04"
set "PROJECT_DIR=/home/haowu/photoprism"
set "OPEN_URL=http://127.0.0.1:2342"

echo.
echo [PhotoPrism] Starting...

set "DOCKER_EXE=%ProgramFiles%\Docker\Docker\Docker Desktop.exe"
if exist "%DOCKER_EXE%" (
  tasklist /FI "IMAGENAME eq Docker Desktop.exe" 2>nul | find /I "Docker Desktop.exe" >nul
  if errorlevel 1 (
    echo [PhotoPrism] Launching Docker Desktop...
    start "" "%DOCKER_EXE%"
  ) else (
    echo [PhotoPrism] Docker Desktop already running.
  )
) else (
  echo [WARN] Docker Desktop not found.
)

set /a WAIT=0
:wait_docker
wsl -d %WSL_DISTRO% docker info >nul 2>&1
if %errorlevel%==0 goto docker_ready
set /a WAIT+=1
if %WAIT% GEQ 60 (
  echo [ERROR] Docker not ready after 3 minutes.
  pause
  exit /b 1
)
echo [PhotoPrism] Waiting for Docker... (%WAIT%/60)
timeout /t 3 /nobreak >nul
goto wait_docker

:docker_ready
echo [PhotoPrism] Docker is ready.

echo [PhotoPrism] docker compose up -d ...
wsl -d %WSL_DISTRO% bash -lc "cd %PROJECT_DIR%; docker compose up -d"
if errorlevel 1 (
  echo [ERROR] compose up failed. Try: wsl --shutdown then restart Docker Desktop.
  pause
  exit /b 1
)

echo [PhotoPrism] Starting photoprism in container...
wsl -d %WSL_DISTRO% bash -lc "cd %PROJECT_DIR%; docker compose exec -T photoprism bash -lc 'cd /go/src/github.com/photoprism/photoprism; nohup ./photoprism start >/tmp/photoprism.log 2>&1 &'"
timeout /t 2 /nobreak >nul

echo.
echo [PhotoPrism] Done. Open: %OPEN_URL%
echo.
set /p OPEN=Open browser now? [Y/n] 
if /I "%OPEN%"=="n" goto end
start "" %OPEN_URL%

:end
echo.
pause
endlocal
