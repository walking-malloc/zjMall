# 整理 buf generate 生成的 swagger 文档文件
# 使用方式: .\scripts\organize-swagger.ps1

Write-Host "正在整理 swagger 文档文件..."

$tempDir = "docs\openapi\temp"
$outputDir = "docs\openapi"

if (-not (Test-Path $tempDir)) {
    Write-Host "❌ temp 目录不存在，请先运行 buf generate"
    exit 1
}

# 查找所有 swagger.json 文件
$swaggerFiles = Get-ChildItem -Path $tempDir -Recurse -Filter "*.swagger.json"

if ($swaggerFiles.Count -eq 0) {
    Write-Host "❌ 未找到 swagger.json 文件"
    exit 1
}

foreach ($file in $swaggerFiles) {
    # 从文件名中提取服务名称
    # 例如: user.swagger.json -> user
    # BaseName 会返回 "user.swagger"，所以需要去掉 ".swagger"
    $baseName = $file.BaseName  # user.swagger.json -> user.swagger
    $serviceName = $baseName -replace '\.swagger$', ''  # user.swagger -> user
    
    # 目标文件路径
    $targetPath = Join-Path $outputDir "$serviceName.swagger.json"
    
    # 移动文件
    Move-Item -Path $file.FullName -Destination $targetPath -Force
    Write-Host "✅ 已移动: $serviceName.swagger.json"
}

# 清理 temp 目录
Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
Write-Host "✅ 已清理 temp 目录"

Write-Host "`n✅ 整理完成！所有文档已移动到 docs/openapi/"

