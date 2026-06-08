@echo off
chcp 65001 >nul
setlocal EnableExtensions

rem Always use the Linux project path (never rely on /mnt/e/... as WSL cwd).
set "WSL_DISTRO=Ubuntu-24.04"
set "PROJECT_DIR=/home/haowu/photoprism"
set "ORIGINALS_MOUNT=/mnt/e/pg/demo/images/mg/_sc/_res/fs/__i2i_fd"
set "OPEN_URL=http://127.0.0.1:2342"
rem Set to 0 to skip WSL restart on each run (faster, but may keep stale bind mounts).
set "RESTART_WSL=1"

echo.
echo [PhotoPrism] Starting...

set "DOCKER_EXE=%ProgramFiles%\Docker\Docker\Docker Desktop.exe"

if "%RESTART_WSL%"=="1" (
  echo [PhotoPrism] Restarting WSL ^(wsl --shutdown^)...
  wsl --shutdown 2>nul
  echo [PhotoPrism] Waiting for WSL to stop...
  timeout /t 8 /nobreak >nul
  if exist "%DOCKER_EXE%" (
    echo [PhotoPrism] Relaunching Docker Desktop after WSL restart...
    start "" "%DOCKER_EXE%"
    timeout /t 15 /nobreak >nul
  )
) else (
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
)

set /a WAIT=0
:wait_docker
wsl -d %WSL_DISTRO% --cd %PROJECT_DIR% docker info >nul 2>&1
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

rem Check E: originals mount when docker-compose.override.yml is used.
wsl -d %WSL_DISTRO% --cd %PROJECT_DIR% bash -lc "test -d '%ORIGINALS_MOUNT%'"
if errorlevel 1 (
  echo.
  echo [WARN] Originals path not available in WSL: %ORIGINALS_MOUNT%
  echo        This usually means drive E: is not mounted after reconnecting the disk.
  echo        Fix options:
  echo          1. Open File Explorer and confirm E: is visible, then retry.
  echo          2. In WSL run:  ls /mnt/e/pg
  echo          3. Temporarily rename docker-compose.override.yml and use storage/originals.
  echo.
  set /p CONT=Continue anyway? [y/N]
  if /I not "%CONT%"=="y" (
    pause
    exit /b 1
  )
)

echo [PhotoPrism] Cleaning stale containers (compose down)...
wsl -d %WSL_DISTRO% --cd %PROJECT_DIR% docker compose down >nul 2>&1

echo [PhotoPrism] docker compose up -d ...
wsl -d %WSL_DISTRO% --cd %PROJECT_DIR% docker compose up -d
if errorlevel 1 (
  echo.
  echo [ERROR] compose up failed.
  echo        Ensure E: is mounted:  wsl -d %WSL_DISTRO% ls /mnt/e/pg
  echo        Or set RESTART_WSL=1 at top of this script and run again.
  pause
  exit /b 1
)

echo [PhotoPrism] Starting photoprism in container (daemon mode)...
wsl -d %WSL_DISTRO% --cd %PROJECT_DIR% docker compose exec -T photoprism bash -lc "cd /go/src/github.com/photoprism/photoprism && ./photoprism stop 2>/dev/null; rm -f storage/photoprism.pid; ./photoprism start -d"
timeout /t 3 /nobreak >nul

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
