# 修复 protobuf 生成代码中的错误导入路径
# 使用方式: .\scripts\fix-protobuf-imports.ps1

Write-Host "正在修复 protobuf 生成的 .pb.go 文件中的导入路径..."

Get-ChildItem -Path gen\go -Recurse -Filter "*.pb.go" | ForEach-Object {
    $content = Get-Content $_.FullName -Raw
    $originalContent = $content
    
    # 替换错误的导入路径
    $content = $content -replace 'zjMall/api/proto/gen/go/google/api', 'google.golang.org/genproto/googleapis/api/annotations'
    
    if ($content -ne $originalContent) {
        Set-Content $_.FullName $content -NoNewline
        Write-Host "已修复: $($_.FullName)"
    }
}

Write-Host "✅ 修复完成！"

