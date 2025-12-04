@echo off
setlocal enabledelayedexpansion

echo SHA256 checksums:

REM Find all .so, .a, .dylib, .dll files recursively in ./libs
for /r "libs" %%f in (*.so *.a *.dylib *.dll) do (
    call :calc_hash "%%f"
)

echo Wrote per-file .sha256 sidecars.
goto :eof

:calc_hash
setlocal enabledelayedexpansion
set "filepath=%~1"

REM Calculate SHA256 using certutil (built into Windows)
for /f "skip=1 delims=" %%h in ('certutil -hashfile "%filepath%" SHA256 2^>nul') do (
    set "line=%%h"
    REM Skip the "CertUtil: -hashfile command completed successfully" line
    if not "!line:CertUtil=!"=="!line!" goto :done_hash
    REM Remove spaces from hash
    set "hash=!line: =!"
    
    REM Get relative path from libs folder
    set "relpath=!filepath:*libs\=!"
    
    echo !hash!  !relpath!
    
    REM Write hash to sidecar file
    echo !hash!> "%filepath%.sha256"
    goto :done_hash
)
:done_hash
endlocal
goto :eof
