@echo off
REM 一键生成 protobuf 代码 + Swagger 文档并修复导入
REM 直接双击本文件，或在命令行运行：scripts\generate-all.bat

setlocal
set SCRIPT_DIR=%~dp0

REM 调用 PowerShell 脚本，自动执行 buf generate + 整理 swagger + 修复导入
powershell -ExecutionPolicy Bypass -File "%SCRIPT_DIR%generate-all.ps1" %*

endlocal

