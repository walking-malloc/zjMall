# 完整的生成流程：生成 protobuf 代码 + 整理文档 + 修复导入
# 使用方式: .\scripts\generate-all.ps1

Write-Host "========================================"
Write-Host "开始生成 Protobuf 代码和文档..."
Write-Host "========================================"
Write-Host ""

# 1. 生成 protobuf 代码和 swagger 文档
Write-Host "📦 步骤 1/3: 运行 buf generate..."
buf generate
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ buf generate 失败"
    exit 1
}
Write-Host "✅ buf generate 完成"
Write-Host ""

# 2. 整理 swagger 文档文件
Write-Host "📁 步骤 2/3: 整理 swagger 文档..."
& .\scripts\organize-swagger.ps1
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 整理文档失败"
    exit 1
}
Write-Host ""

# 3. 修复导入路径
Write-Host "🔧 步骤 3/3: 修复 protobuf 导入路径..."
& .\scripts\fix-protobuf-imports.ps1
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 修复导入路径失败"
    exit 1
}
Write-Host ""

Write-Host "========================================"
Write-Host "✅ 所有步骤完成！"
Write-Host "========================================"
Write-Host ""
Write-Host "生成的文件："
Get-ChildItem -Path "docs\openapi" -Filter "*.swagger.json" | ForEach-Object {
    Write-Host "  - $($_.Name)"
}

