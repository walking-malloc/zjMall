# 一键修复 protobuf 生成代码中的 google/api 导入路径（可选顺带执行 buf generate）
# 使用方式（在仓库根目录执行）:
#   powershell -ExecutionPolicy Bypass -File .\scripts\fix-protobuf-imports.ps1
#
# 如果你不在仓库根目录执行，也没关系，脚本会自动切到仓库根目录。

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

chcp 65001 | Out-Null
$OutputEncoding = [System.Text.Encoding]::UTF8

# 切换到仓库根目录（脚本所在目录的上一级）
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot  = Resolve-Path (Join-Path $scriptDir "..")
Set-Location $repoRoot

Write-Host "📁 当前仓库根目录: $repoRoot"

# 1. 如存在 buf.yaml，则先自动执行 buf generate
if (Test-Path ".\buf.yaml") {
    Write-Host "▶ 执行 buf generate..."
    try {
        buf generate
        Write-Host "✅ buf generate 完成"
    } catch {
        Write-Host "⚠️ buf generate 失败: $($_.Exception.Message)"
        Write-Host "   继续尝试修复已存在的 .pb.go 文件导入路径..."
    }
} else {
    Write-Host "ℹ️ 未检测到 buf.yaml，跳过 buf generate"
}

# 2. 修复 gen\go 下所有 .pb.go 文件中的 google/api 导入
Write-Host "🔧 正在修复 protobuf 生成的 .pb.go 文件中的导入路径..."

$pbRoot = Join-Path $repoRoot "gen\go"
if (-not (Test-Path $pbRoot)) {
    Write-Host "ℹ️ 未找到 gen\go 目录，无需修复"
    Write-Host "[OK] 处理结束"
    return
}

$fixedCount = 0

Get-ChildItem -Path $pbRoot -Recurse -Filter "*.pb.go" | ForEach-Object {
    $filePath = $_.FullName
    $content = Get-Content $filePath -Raw -Encoding UTF8
    $originalContent = $content

    # buf generate 生成的错误导入一般是:
    #   _ "zjMall/gen/go/google/api"
    # 早期版本可能是:
    #   _ "zjMall/api/proto/gen/go/google/api"
    #
    # 我们统一替换为标准的 annotations 包:
    #   _ "google.golang.org/genproto/googleapis/api/annotations"

    $content = $content -replace 'zjMall/api/proto/gen/go/google/api', 'google.golang.org/genproto/googleapis/api/annotations'
    $content = $content -replace 'zjMall/gen/go/google/api', 'google.golang.org/genproto/googleapis/api/annotations'

    if ($content -ne $originalContent) {
        [System.IO.File]::WriteAllText($filePath, $content, [System.Text.UTF8Encoding]::new($false))
        Write-Host "✅ Fixed import in: $($_.FullName.Substring($repoRoot.Path.Length + 1))"
        $fixedCount++
    }
}

if ($fixedCount -eq 0) {
    Write-Host "OK: no google/api imports needed fixing in any .pb.go files."
} else {
    Write-Host ("OK: fixed google/api imports in {0} .pb.go files." -f $fixedCount)
}

