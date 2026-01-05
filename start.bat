@echo off
chcp 65001 >nul
echo ==================================
echo Video Conference - Запуск
echo ==================================
echo.

REM Проверка Docker
where docker >nul 2>&1
if %errorlevel% neq 0 (
    echo [ОШИБКА] Docker не установлен. Установите Docker Desktop и повторите попытку.
    pause
    exit /b 1
)

echo [OK] Docker найден
echo.

REM Получаем IP адрес локальной сети для подключения с телефона
for /f "tokens=2 delims=:" %%a in ('ipconfig ^| findstr /R /C:"IPv4"') do (
    set "LOCALIP=%%a"
    goto :found
)
:found
set LOCALIP=%LOCALIP:~1%

echo Локальный IP: %LOCALIP%
set HOST_IP=%LOCALIP%
echo HOST_IP=%HOST_IP%
echo.

REM Создаём .env файл с HOST_IP для docker-compose
echo HOST_IP=%HOST_IP%> .env.local
echo LIVEKIT_API_KEY=devkey>> .env.local
echo LIVEKIT_API_SECRET=secret>> .env.local

REM Остановка существующих контейнеров
echo Остановка существующих контейнеров...
docker-compose --env-file .env.local down

echo.
echo Сборка и запуск контейнеров...
docker-compose --env-file .env.local up --build -d

echo.
echo Ожидание запуска сервисов...
timeout /t 10 /nobreak >nul

REM Проверка статуса
echo.
echo Статус сервисов:
docker-compose --env-file .env.local ps

echo.
echo ==================================
echo Приложение запущено!
echo ==================================
echo.
echo Откройте в браузере:
echo   На ПК:       http://localhost
echo   На телефоне: http://%LOCALIP%
echo.
echo Для видеоконференции между ПК и телефоном:
echo   1. Откройте http://localhost на ПК
echo   2. Создайте комнату
echo   3. Скопируйте ID комнаты
echo   4. На телефоне откройте http://%LOCALIP%
echo   5. Присоединитесь по ID комнаты
echo.
echo Просмотр логов: docker-compose --env-file .env.local logs -f
echo Остановка:      docker-compose --env-file .env.local down
echo.
pause

