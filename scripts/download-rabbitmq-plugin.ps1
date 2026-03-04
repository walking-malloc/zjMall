# 下载 RabbitMQ 延迟消息插件
# 使用方法: .\scripts\download-rabbitmq-plugin.ps1

$pluginUrl = "https://github.com/rabbitmq/rabbitmq-delayed-message-exchange/releases/download/v3.13.0/rabbitmq_delayed_message_exchange-3.13.0.ez"
$outputFile = "rabbitmq_delayed_message_exchange-3.13.0.ez"
$maxRetries = 3

Write-Host "正在下载 RabbitMQ 延迟消息插件..." -ForegroundColor Yellow
Write-Host "URL: $pluginUrl" -ForegroundColor Gray

# 如果文件已存在，询问是否覆盖
if (Test-Path $outputFile) {
    $response = Read-Host "文件已存在，是否重新下载？(Y/N)"
    if ($response -ne "Y" -and $response -ne "y") {
        Write-Host "跳过下载，使用现有文件" -ForegroundColor Green
        exit 0
    }
    Remove-Item $outputFile -Force
}

# 重试下载
for ($i = 1; $i -le $maxRetries; $i++) {
    Write-Host "尝试下载 (第 $i/$maxRetries 次)..." -ForegroundColor Cyan
    
    try {
        # 使用 Invoke-WebRequest 下载文件，添加超时和重试
        $ProgressPreference = 'SilentlyContinue'  # 禁用进度条，避免输出过多
        Invoke-WebRequest -Uri $pluginUrl -OutFile $outputFile -UseBasicParsing -TimeoutSec 60 -ErrorAction Stop
        
        if (Test-Path $outputFile) {
            $fileSize = (Get-Item $outputFile).Length
            if ($fileSize -gt 0) {
                Write-Host "✅ 插件下载成功: $outputFile ($([math]::Round($fileSize/1KB, 2)) KB)" -ForegroundColor Green
                Write-Host "现在可以运行: docker-compose build rabbitmq" -ForegroundColor Cyan
                exit 0
            } else {
                Write-Host "⚠️ 下载的文件为空，重试中..." -ForegroundColor Yellow
                Remove-Item $outputFile -Force -ErrorAction SilentlyContinue
            }
        }
    } catch {
        Write-Host "⚠️ 第 $i 次下载失败: $($_.Exception.Message)" -ForegroundColor Yellow
        if ($i -lt $maxRetries) {
            Write-Host "等待 3 秒后重试..." -ForegroundColor Gray
            Start-Sleep -Seconds 3
        }
        Remove-Item $outputFile -Force -ErrorAction SilentlyContinue
    }
}

# 所有重试都失败
Write-Host "`n❌ 下载失败：已尝试 $maxRetries 次" -ForegroundColor Red
Write-Host "`n请手动下载插件文件：" -ForegroundColor Yellow
Write-Host "1. 访问以下链接：" -ForegroundColor Cyan
Write-Host "   $pluginUrl" -ForegroundColor White
Write-Host "`n2. 下载后，将文件重命名为: rabbitmq_delayed_message_exchange-3.13.0.ez" -ForegroundColor Yellow
Write-Host "3. 放置在项目根目录（与 docker-compose.yml 同级）" -ForegroundColor Yellow
Write-Host "`n或者使用其他下载工具（如浏览器、IDM 等）" -ForegroundColor Gray
exit 1
