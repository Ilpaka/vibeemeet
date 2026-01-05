# ==================================
# Скрипт запуска видеоконференции в локальной сети
# ==================================

Write-Host "==================================" -ForegroundColor Cyan
Write-Host "Video Conference - Локальная сеть" -ForegroundColor Cyan
Write-Host "==================================" -ForegroundColor Cyan
Write-Host ""

# Функция для получения локального IP
function Get-LocalIP {
    $adapters = Get-NetIPAddress -AddressFamily IPv4 | 
        Where-Object { $_.IPAddress -notlike "127.*" -and $_.IPAddress -notlike "169.254.*" } |
        Sort-Object -Property InterfaceIndex
    
    foreach ($adapter in $adapters) {
        $ip = $adapter.IPAddress
        if ($ip -match "^192\.168\." -or $ip -match "^10\." -or $ip -match "^172\.") {
            return $ip
        }
    }
    
    # Fallback
    if ($adapters.Count -gt 0) {
        return $adapters[0].IPAddress
    }
    
    return "localhost"
}

# Получаем IP
$hostIP = Get-LocalIP
Write-Host "Обнаружен IP адрес: $hostIP" -ForegroundColor Green
Write-Host ""

# Устанавливаем переменные окружения
$env:HOST_IP = $hostIP
Write-Host "Установлена переменная HOST_IP=$hostIP" -ForegroundColor Yellow

# Проверяем наличие Docker
$dockerAvailable = $false
try {
    docker --version | Out-Null
    $dockerAvailable = $true
} catch {
    $dockerAvailable = $false
}

if ($dockerAvailable) {
    Write-Host ""
    Write-Host "Docker обнаружен. Выберите режим запуска:" -ForegroundColor Cyan
    Write-Host "1. Запустить через Docker Compose (рекомендуется)"
    Write-Host "2. Запустить только сервер (без Docker)"
    Write-Host ""
    $choice = Read-Host "Ваш выбор (1/2)"
    
    if ($choice -eq "1") {
        Write-Host ""
        Write-Host "Запуск Docker Compose..." -ForegroundColor Green
        Write-Host "HOST_IP будет передан в контейнеры" -ForegroundColor Yellow
        Write-Host ""
        
        docker-compose down 2>$null
        docker-compose up --build -d
        
        Write-Host ""
        Write-Host "==================================" -ForegroundColor Green
        Write-Host "Сервер запущен!" -ForegroundColor Green
        Write-Host "==================================" -ForegroundColor Green
        Write-Host ""
        Write-Host "Откройте в браузере:" -ForegroundColor Cyan
        Write-Host "  На ПК:       http://localhost" -ForegroundColor White
        Write-Host "  На телефоне: http://${hostIP}" -ForegroundColor White
        Write-Host ""
        Write-Host "Для просмотра логов: docker-compose logs -f" -ForegroundColor Yellow
        Write-Host "Для остановки:       docker-compose down" -ForegroundColor Yellow
        exit
    }
}

# Запуск без Docker
Write-Host ""
Write-Host "Запуск сервера напрямую..." -ForegroundColor Green
Write-Host ""
Write-Host "ВАЖНО: Убедитесь, что:" -ForegroundColor Yellow
Write-Host "  1. PostgreSQL запущен на localhost:5432" -ForegroundColor White
Write-Host "  2. Redis запущен на localhost:6379" -ForegroundColor White
Write-Host "  3. LiveKit Server запущен на порту 7880" -ForegroundColor White
Write-Host ""

# Проверяем наличие server.exe
if (Test-Path "server.exe") {
    Write-Host "Найден server.exe, запускаем..." -ForegroundColor Green
    ./server.exe
} elseif (Test-Path "cmd/server/main.go") {
    Write-Host "Компилируем и запускаем Go сервер..." -ForegroundColor Green
    go run ./cmd/server/main.go
} else {
    Write-Host "Ошибка: Не найден исполняемый файл сервера!" -ForegroundColor Red
    Write-Host "Выполните: go build -o server.exe ./cmd/server" -ForegroundColor Yellow
}

