@echo off
chcp 65001 >nul
title Video Conference - Локальная сеть

echo ==================================
echo Video Conference - Локальная сеть
echo ==================================
echo.

REM Получаем IP адрес локальной сети
for /f "tokens=2 delims=:" %%a in ('ipconfig ^| findstr /R /C:"IPv4"') do (
    set "IP=%%a"
    goto :found
)
:found
set IP=%IP:~1%

echo Обнаружен IP адрес: %IP%
echo.

REM Устанавливаем переменную окружения
set HOST_IP=%IP%
echo Установлена переменная HOST_IP=%IP%
echo.

REM Проверяем наличие Docker
docker --version >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo Docker обнаружен.
    echo.
    echo Выберите режим запуска:
    echo 1. Запустить через Docker Compose ^(рекомендуется^)
    echo 2. Запустить только сервер ^(без Docker^)
    echo.
    set /p choice="Ваш выбор (1/2): "
    
    if "%choice%"=="1" (
        echo.
        echo Запуск Docker Compose...
        echo.
        docker-compose down >nul 2>&1
        docker-compose up --build -d
        
        echo.
        echo ==================================
        echo Сервер запущен!
        echo ==================================
        echo.
        echo Откройте в браузере:
        echo   На ПК:       http://localhost
        echo   На телефоне: http://%IP%
        echo.
        echo Для просмотра логов: docker-compose logs -f
        echo Для остановки:       docker-compose down
        echo.
        pause
        exit /b
    )
)

echo.
echo Запуск сервера напрямую...
echo.
echo ВАЖНО: Убедитесь, что:
echo   1. PostgreSQL запущен на localhost:5432
echo   2. Redis запущен на localhost:6379
echo   3. LiveKit Server запущен на порту 7880
echo.

if exist "server.exe" (
    echo Найден server.exe, запускаем...
    server.exe
) else (
    echo Компилируем и запускаем Go сервер...
    go run ./cmd/server/main.go
)

pause

