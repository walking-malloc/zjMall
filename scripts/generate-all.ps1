# Protobuf + swagger generation pipeline
# Usage: .\scripts\generate-all.ps1

chcp 65001 | Out-Null
$OutputEncoding = [System.Text.Encoding]::UTF8

Write-Host "========================================"
Write-Host "Protobuf + swagger generation"
Write-Host "========================================"
Write-Host ""

Write-Host "[1/3] buf generate..."
buf generate
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERR] buf generate failed"
    exit 1
}
Write-Host "[OK] buf generate done"
Write-Host ""

Write-Host "[2/3] Organize swagger..."
& .\scripts\organize-swagger.ps1
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERR] organize swagger failed"
    exit 1
}
Write-Host ""

Write-Host "[3/3] Fix protobuf imports..."
& .\scripts\fix-protobuf-imports.ps1
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERR] fix imports failed"
    exit 1
}
Write-Host ""

Write-Host "========================================"
Write-Host "[OK] All steps done"
Write-Host "========================================"
Write-Host ""
Write-Host "Generated:"
Get-ChildItem -Path docs\openapi -Filter '*.swagger.json' | ForEach-Object { Write-Host ('  - ' + $_.Name) }
