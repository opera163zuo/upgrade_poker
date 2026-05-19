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
if errorlevel 1 goto :fail

echo [2/3] 复制到 release 目录...
copy /Y upgrade.exe release\upgrade_%VERSION%_win64.exe >nul
if errorlevel 1 goto :fail

echo [3/3] 生成 Windows zip 包...
powershell -NoProfile -Command "Compress-Archive -LiteralPath 'release\upgrade_%VERSION%_win64.exe' -DestinationPath 'release\upgrade_%VERSION%_win64.zip' -Force"
if errorlevel 1 goto :fail

echo.
echo ========================================
echo  完成
echo  Windows: upgrade.exe
echo  Release: release\upgrade_%VERSION%_win64.exe
echo  Zip: release\upgrade_%VERSION%_win64.zip
echo  运行: upgrade.exe
echo ========================================
pause
goto :eof

:fail
echo.
echo 构建失败，请检查上面的错误输出。
pause
exit /b 1
