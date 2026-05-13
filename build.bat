@echo off
chcp 65001 >nul
set VERSION=%date:~0,4%%date:~5,2%%date:~8,2%
echo ========================================
echo  upgrade_poker - 构建工具
echo  版本: %VERSION%
echo ========================================

if not exist release mkdir release

echo.
echo [1/3] 编译 Windows 64位版...
go build -ldflags="-s -w -X main.Version=%VERSION%" -o upgrade.exe .

echo [2/3] 复制到 release 目录...
copy /Y upgrade.exe release\upgrade_%VERSION%_win64.exe >nul

echo [3/3] 编译 Linux 64位版（如需在Linux测试）...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w -X main.Version=%VERSION%" -o release\upgrade_%VERSION%_linux64 .
set GOOS=windows

echo.
echo ========================================
echo  完成！
echo  Windows: upgrade.exe
echo  Release: release\upgrade_%VERSION%_win64.exe
echo  运行: upgrade.exe 或 upgrade.exe -ui=gui
echo ========================================
pause
