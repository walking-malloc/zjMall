# 简单的头像上传脚本（使用 curl）
param(
    [Parameter(Mandatory=$true)]
    [string]$Token,
    
    [Parameter(Mandatory=$true)]
    [string]$FilePath
)

# 去掉引号
$Token = $Token.Trim('"').Trim()
$FilePath = $FilePath.Trim('"').Trim()

# 检查文件
if (-not (Test-Path $FilePath)) {
    Write-Host "错误: 文件不存在 - $FilePath" -ForegroundColor Red
    exit 1
}

Write-Host "文件路径: $FilePath" -ForegroundColor Green
Write-Host "文件大小: $([math]::Round((Get-Item $FilePath).Length / 1KB, 2)) KB" -ForegroundColor Green

# 确保Token有Bearer前缀
if (-not $Token.StartsWith("Bearer ")) {
    $Token = "Bearer $Token"
}

$uri = "http://localhost:8081/api/v1/users/avatar"

Write-Host "`n正在上传到: $uri" -ForegroundColor Yellow

try {
    # 使用 curl（Windows 10+ 自带）
    $response = curl.exe -X POST $uri `
        -H "Authorization: $Token" `
        -F "avatar=@$FilePath" `
        2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "`n上传成功！" -ForegroundColor Green
        Write-Host "响应:" -ForegroundColor Green
        
        # 尝试解析JSON
        try {
            $json = $response | ConvertFrom-Json
            $json | ConvertTo-Json -Depth 10
            
            if ($json.code -eq 0) {
                Write-Host "`n头像URL: $($json.avatar_url)" -ForegroundColor Cyan
                Write-Host "可以在浏览器中打开查看" -ForegroundColor Cyan
            }
        } catch {
            # 如果不是JSON，直接显示
            $response
        }
    } else {
        Write-Host "`n上传失败！" -ForegroundColor Red
        Write-Host "curl 错误代码: $LASTEXITCODE" -ForegroundColor Red
        Write-Host "响应:" -ForegroundColor Red
        $response
        exit 1
    }
    
} catch {
    Write-Host "`n错误: $_" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    exit 1
}










