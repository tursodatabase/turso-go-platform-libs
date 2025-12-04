@echo off
setlocal ENABLEEXTENSIONS ENABLEDELAYEDEXPANSION

:: ---------------------------
:: Default env variables
:: ---------------------------
if "%TURSO_RS_REPO%"=="" set TURSO_RS_REPO=https://github.com/tursodatabase/turso.git
if "%TURSO_RS_BUILD_PROFILE%"=="" set TURSO_RS_BUILD_PROFILE=lib-release
if "%TURSO_RS_BUILD_DIR%"=="" set TURSO_RS_BUILD_DIR=turso-rs
if "%TURSO_RS_PACKAGE%"=="" set TURSO_RS_PACKAGE=turso_sync_sdk_kit
if "%TURSO_RS_LIBC_VARIANT%"=="" set TURSO_RS_LIBC_VARIANT=
if "%TURSO_GO_LIB_DIR%"=="" set TURSO_GO_LIB_DIR=libs

if "%TURSO_RS_BUILD_REF%"=="" (
    echo TURSO_RS_BUILD_REF env var must be set
    exit /b 1
)

:: --------------------------------
:: OS/ARCH detection via PowerShell
:: --------------------------------
for /f "delims=" %%i in ('powershell -NoProfile -Command "[System.Environment]::OSVersion.Platform"') do set __PLAT=%%i
for /f "delims=" %%i in ('powershell -NoProfile -Command "$env:PROCESSOR_ARCHITECTURE"') do set UNAME_M=%%i

:: Normalize OS
set OS=windows

:: Normalize ARCH
if /I "%UNAME_M%"=="ARM64" (
    set ARCH=arm64
) else (
    set ARCH=amd64
)

:: Windows never uses musl/glibc variants; ignore
:: TURSO_RS_LIBC_VARIANT remains empty

:: --------------------------------
:: Platform + output file name
:: --------------------------------
set PLATFORM=%OS%_%ARCH%%TURSO_RS_LIBC_VARIANT%
set TURSO_GO_LIB_PATH=%TURSO_GO_LIB_DIR%\%PLATFORM%

if /I "%OS%"=="windows" (
    set OUTPUT_NAME=%TURSO_RS_PACKAGE%.dll
)

:: --------------------------------
:: musl logic skipped (not used on Windows)
set RUST_TARGET=

:: Determine cargo output directory (no --target here on Windows)
set CARGO_OUT_DIR=%TURSO_RS_BUILD_DIR%\target\%TURSO_RS_BUILD_PROFILE%
set CARGO_LIB_PATH=%CARGO_OUT_DIR%\%OUTPUT_NAME%

echo TURSO_RS_REPO: %TURSO_RS_REPO%
echo TURSO_RS_BUILD_REF: %TURSO_RS_BUILD_REF%
echo TURSO_RS_BUILD_DIR: %TURSO_RS_BUILD_DIR%
echo TURSO_RS_PACKAGE: %TURSO_RS_PACKAGE%
echo PLATFORM: %PLATFORM%
echo OUTPUT_NAME: %OUTPUT_NAME%
echo CARGO_OUT_DIR: %CARGO_OUT_DIR%
echo CARGO_LIB_PATH: %CARGO_LIB_PATH%
echo TURSO_GO_LIB_PATH: %TURSO_GO_LIB_PATH%

:: --------------------------------
:: Clone repo
:: --------------------------------
git clone --single-branch --depth 1 --branch "%TURSO_RS_BUILD_REF%" "%TURSO_RS_REPO%" "%TURSO_RS_BUILD_DIR%"
if errorlevel 1 (
    echo Failed to clone repo
    exit /b 1
)

:: --------------------------------
:: Build with cargo
:: --------------------------------
pushd "%TURSO_RS_BUILD_DIR%"
echo Building %TURSO_RS_PACKAGE% (%TURSO_RS_BUILD_PROFILE%) for %PLATFORM%
cargo build --profile "%TURSO_RS_BUILD_PROFILE%" --package "%TURSO_RS_PACKAGE%"
if errorlevel 1 (
    echo Cargo build failed
    popd
    exit /b 1
)
popd

:: --------------------------------
:: Check artifact
:: --------------------------------
if not exist "%CARGO_LIB_PATH%" (
    echo Expected artifact not found: %CARGO_LIB_PATH%
    echo Contents of %CARGO_OUT_DIR%:
    dir "%CARGO_OUT_DIR%"
    exit /b 1
)

:: --------------------------------
:: Copy library
:: --------------------------------
if not exist "%TURSO_GO_LIB_PATH%" mkdir "%TURSO_GO_LIB_PATH%"
copy /y "%CARGO_LIB_PATH%" "%TURSO_GO_LIB_PATH%\"

echo Wrote %CD%\%TURSO_GO_LIB_PATH%\%OUTPUT_NAME%

exit /b 0
