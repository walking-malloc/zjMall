# 从 Git 中移除不应该被跟踪的文件（但保留本地文件）
# 使用方法: .\scripts\clean-git-tracked-files.ps1

Write-Host "正在从 Git 中移除不应该被跟踪的文件..." -ForegroundColor Yellow

# 移除可执行文件
Write-Host "`n移除可执行文件..." -ForegroundColor Cyan
git rm --cached cart-service.exe 2>$null
git rm --cached inventory.exe 2>$null
git rm --cached order.exe 2>$null
git rm --cached payment.exe 2>$null
git rm --cached product-service.exe 2>$null
git rm --cached user-service.exe 2>$null

# 移除缓存目录
Write-Host "移除缓存目录..." -ForegroundColor Cyan
git rm -r --cached cache/ 2>$null

# 移除生成的代码目录
Write-Host "移除生成的代码目录..." -ForegroundColor Cyan
git rm -r --cached gen/ 2>$null

# 移除日志目录（如果存在）
Write-Host "移除日志目录..." -ForegroundColor Cyan
git rm -r --cached log/ 2>$null
git rm -r --cached logs/ 2>$null

# 移除前端构建产物和依赖（如果存在）
Write-Host "移除前端构建产物和依赖..." -ForegroundColor Cyan
git rm -r --cached frontend/dist/ 2>$null
git rm -r --cached frontend/node_modules/ 2>$null

Write-Host "`n完成！这些文件已从 Git 索引中移除，但本地文件仍然保留。" -ForegroundColor Green
Write-Host "请运行 'git status' 查看更改，然后提交这些更改。" -ForegroundColor Yellow
Write-Host "`n注意: 如果 configs/config.yaml 包含敏感信息，建议也移除它：" -ForegroundColor Yellow
Write-Host "  git rm --cached configs/config.yaml" -ForegroundColor Gray
