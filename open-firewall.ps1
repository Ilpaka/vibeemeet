# Запустите этот скрипт от имени Администратора!
# PowerShell -> Правый клик -> "Запуск от имени администратора"

Write-Host "Открытие портов для видеоконференции..." -ForegroundColor Green

# HTTP
New-NetFirewallRule -DisplayName "VideoConf HTTP" -Direction Inbound -LocalPort 80 -Protocol TCP -Action Allow

# LiveKit WebSocket
New-NetFirewallRule -DisplayName "VideoConf LiveKit WS" -Direction Inbound -LocalPort 7880 -Protocol TCP -Action Allow

# LiveKit TCP fallback
New-NetFirewallRule -DisplayName "VideoConf LiveKit TCP" -Direction Inbound -LocalPort 7881 -Protocol TCP -Action Allow

# LiveKit UDP WebRTC
New-NetFirewallRule -DisplayName "VideoConf LiveKit UDP" -Direction Inbound -LocalPort 7882 -Protocol UDP -Action Allow

# RTP Media порты (UDP)
New-NetFirewallRule -DisplayName "VideoConf RTP UDP" -Direction Inbound -LocalPort 50000-50100 -Protocol UDP -Action Allow

Write-Host "`nПорты успешно открыты!" -ForegroundColor Green
Write-Host "Теперь подключение с телефона/ноутбука должно работать." -ForegroundColor Yellow

