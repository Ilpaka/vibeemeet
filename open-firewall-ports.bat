@echo off
:: Этот скрипт должен быть запущен от имени Администратора!
:: Правый клик -> "Запуск от имени администратора"

echo ==================================
echo Открытие портов в Windows Firewall
echo ==================================
echo.

:: Проверка прав администратора
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo ОШИБКА: Запустите этот скрипт от имени Администратора!
    echo.
    echo Правый клик на файле -^> "Запуск от имени администратора"
    echo.
    pause
    exit /b 1
)

echo Добавление правил firewall...

netsh advfirewall firewall add rule name="Video Conference HTTP" dir=in action=allow protocol=TCP localport=80
netsh advfirewall firewall add rule name="Video Conference LiveKit WS" dir=in action=allow protocol=TCP localport=7880
netsh advfirewall firewall add rule name="Video Conference LiveKit UDP" dir=in action=allow protocol=UDP localport=50000-50100
netsh advfirewall firewall add rule name="Video Conference Backend" dir=in action=allow protocol=TCP localport=8081

echo.
echo ==================================
echo Порты успешно открыты!
echo ==================================
echo.
echo Открытые порты:
echo   80    - HTTP (веб-интерфейс)
echo   7880  - LiveKit WebSocket
echo   8081  - Backend API
echo   50000-50100 - LiveKit UDP (медиа)
echo.
pause

