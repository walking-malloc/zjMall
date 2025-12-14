# 测试头像上传接口
$token = Read-Host "请输入你的Token"

# 去掉Token中的引号
$token = $token.Trim('"').Trim()

$filePath = Read-Host "请输入图片文件路径（或拖拽文件到这里）"

# 去掉路径中的引号（Windows拖拽文件可能会带引号）
$filePath = $filePath.Trim('"').Trim()

if (-not (Test-Path $filePath)) {
    Write-Host "文件不存在: $filePath" -ForegroundColor Red
    exit
}

Write-Host "文件路径: $filePath" -ForegroundColor Green

$uri = "http://localhost:8081/api/v1/users/avatar"

$headers = @{
    "Authorization" = "Bearer $token"
}

try {
    $fileBytes = [System.IO.File]::ReadAllBytes($filePath)
    $boundary = [System.Guid]::NewGuid().ToString()
    $fileContent = [System.IO.File]::ReadAllBytes($filePath)
    $fileName = [System.IO.Path]::GetFileName($filePath)
    
    $bodyLines = @(
        "--$boundary",
        "Content-Disposition: form-data; name=`"avatar`"; filename=`"$fileName`"",
        "Content-Type: image/jpeg",
        "",
        [System.Text.Encoding]::GetEncoding("ISO-8859-1").GetString($fileContent),
        "--$boundary--"
    )
    
    $body = $bodyLines -join "`r`n"
    $bodyBytes = [System.Text.Encoding]::GetEncoding("ISO-8859-1").GetBytes($body)
    
    $response = Invoke-RestMethod -Uri $uri -Method Post -Headers $headers -ContentType "multipart/form-data; boundary=$boundary" -Body $bodyBytes
    
    Write-Host "上传成功！" -ForegroundColor Green
    Write-Host "响应: $($response | ConvertTo-Json)"
} catch {
    Write-Host "上传失败: $_" -ForegroundColor Red
    Write-Host "状态码: $($_.Exception.Response.StatusCode.value__)" -ForegroundColor Red
}

