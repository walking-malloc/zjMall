# 简单的头像上传测试脚本
param(
    [Parameter(Mandatory=$true)]
    [string]$Token,
    
    [Parameter(Mandatory=$true)]
    [string]$FilePath
)

# 去掉路径和Token中的引号
$Token = $Token.Trim('"').Trim()
$FilePath = $FilePath.Trim('"').Trim()

# 检查文件是否存在
if (-not (Test-Path $FilePath)) {
    Write-Host "错误: 文件不存在 - $FilePath" -ForegroundColor Red
    exit 1
}

Write-Host "文件路径: $FilePath" -ForegroundColor Green
Write-Host "文件大小: $((Get-Item $FilePath).Length / 1KB) KB" -ForegroundColor Green

# 确保Token有Bearer前缀
if (-not $Token.StartsWith("Bearer ")) {
    $Token = "Bearer $Token"
}

$uri = "http://localhost:8081/api/v1/users/avatar"

try {
    Write-Host "`n正在上传..." -ForegroundColor Yellow
    
    # 使用 curl（如果可用）或 .NET 的 HttpClient
    $fileItem = Get-Item $FilePath
    $fileName = $fileItem.Name
    
    # 方法1: 尝试使用 curl（Windows 10+ 自带）
    $curlCmd = "curl"
    if (Get-Command $curlCmd -ErrorAction SilentlyContinue) {
        Write-Host "使用 curl 上传..." -ForegroundColor Cyan
        
        $headers = @(
            "-H", "Authorization: $Token"
        )
        
        $response = & curl -X POST $uri $headers -F "avatar=@$FilePath" 2>&1
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "`n上传成功！" -ForegroundColor Green
            $response
            exit 0
        } else {
            Write-Host "`n上传失败 (curl)" -ForegroundColor Red
            $response
            exit 1
        }
    }
    
    # 方法2: 使用 .NET HttpClient（备用方案）
    Write-Host "使用 .NET HttpClient 上传..." -ForegroundColor Cyan
    
    Add-Type -AssemblyName System.Net.Http
    
    $httpClient = New-Object System.Net.Http.HttpClient
    $httpClient.DefaultRequestHeaders.Add("Authorization", $Token)
    
    $multipartContent = New-Object System.Net.Http.MultipartFormDataContent
    $fileStream = [System.IO.File]::OpenRead($FilePath)
    $streamContent = New-Object System.Net.Http.StreamContent($fileStream)
    
    # 根据文件扩展名设置Content-Type
    $ext = [System.IO.Path]::GetExtension($FilePath).ToLower()
    $contentType = switch ($ext) {
        ".jpg" { "image/jpeg" }
        ".jpeg" { "image/jpeg" }
        ".png" { "image/png" }
        ".gif" { "image/gif" }
        ".webp" { "image/webp" }
        default { "application/octet-stream" }
    }
    
    $streamContent.Headers.ContentType = New-Object System.Net.Http.Headers.MediaTypeHeaderValue($contentType)
    $multipartContent.Add($streamContent, "avatar", $fileName)
    
    $response = $httpClient.PostAsync($uri, $multipartContent).Result
    
    $fileStream.Close()
    $httpClient.Dispose()
    
    $responseBody = $response.Content.ReadAsStringAsync().Result
    
    if ($response.IsSuccessStatusCode) {
        Write-Host "`n上传成功！" -ForegroundColor Green
        Write-Host "HTTP状态: $($response.StatusCode)" -ForegroundColor Green
        Write-Host "响应内容:" -ForegroundColor Green
        $responseBody | ConvertFrom-Json | ConvertTo-Json -Depth 10
    } else {
        Write-Host "`n上传失败！" -ForegroundColor Red
        Write-Host "HTTP状态: $($response.StatusCode)" -ForegroundColor Red
        Write-Host "响应内容: $responseBody" -ForegroundColor Red
        exit 1
    }
    
} catch {
    Write-Host "`n错误: $_" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    if ($_.Exception.InnerException) {
        Write-Host "内部错误: $($_.Exception.InnerException.Message)" -ForegroundColor Red
    }
    exit 1
}











