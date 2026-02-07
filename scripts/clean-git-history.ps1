# 从 Git 历史记录中完全删除不应该被跟踪的文件
# ⚠️ 警告：这是破坏性操作，会重写 Git 历史！
# 使用方法: .\scripts\clean-git-history.ps1

Write-Host "⚠️  警告：此操作将重写 Git 历史记录！" -ForegroundColor Red
Write-Host "⚠️  执行后需要强制推送到远程仓库：git push --force" -ForegroundColor Red
Write-Host "⚠️  所有团队成员需要重新克隆仓库！" -ForegroundColor Red
Write-Host ""
$confirm = Read-Host "确认要继续吗？(yes/no)"

if ($confirm -ne "yes") {
    Write-Host "操作已取消" -ForegroundColor Yellow
    exit
}

Write-Host "`n开始清理 Git 历史记录..." -ForegroundColor Cyan
Write-Host "这可能需要几分钟时间，请耐心等待...`n" -ForegroundColor Yellow

# 要删除的文件和目录列表
$itemsToRemove = @(
    "cart-service.exe",
    "inventory.exe",
    "order.exe",
    "payment.exe",
    "product-service.exe",
    "user-service.exe",
    "cache",
    "gen",
    "log",
    "logs",
    "frontend/dist",
    "frontend/node_modules"
)

Write-Host "正在删除以下文件和目录：" -ForegroundColor Cyan
foreach ($item in $itemsToRemove) {
    Write-Host "  - $item" -ForegroundColor Gray
}

# 构建 git filter-branch 命令
# 使用 --index-filter 来删除文件，这样更快
$filterScript = ""
foreach ($item in $itemsToRemove) {
    $filterScript += "git rm -rf --cached --ignore-unmatch `"$item`" 2>/dev/null; "
}

Write-Host "`n执行 git filter-branch（这可能需要几分钟）..." -ForegroundColor Yellow

# 执行清理
$env:GIT_AUTHOR_NAME = (git config user.name)
$env:GIT_AUTHOR_EMAIL = (git config user.email)
$env:GIT_COMMITTER_NAME = (git config user.name)
$env:GIT_COMMITTER_EMAIL = (git config user.email)

git filter-branch --force --index-filter $filterScript --prune-empty --tag-name-filter cat -- --all

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n✅ Git 历史记录清理完成！" -ForegroundColor Green
    Write-Host ""
    Write-Host "清理备份引用（可选，释放空间）：" -ForegroundColor Yellow
    Write-Host "  git for-each-ref --format='delete %(refname)' refs/original | git update-ref --stdin" -ForegroundColor Gray
    Write-Host "  git reflog expire --expire=now --all" -ForegroundColor Gray
    Write-Host "  git gc --prune=now --aggressive" -ForegroundColor Gray
    Write-Host ""
    Write-Host "下一步操作：" -ForegroundColor Yellow
    Write-Host "1. 检查更改: git log --oneline" -ForegroundColor Gray
    Write-Host "2. 验证文件已删除: git log --all --full-history -- '*.exe'" -ForegroundColor Gray
    Write-Host "3. 强制推送到远程: git push --force --all" -ForegroundColor Gray
    Write-Host "4. 强制推送标签: git push --force --tags" -ForegroundColor Gray
    Write-Host "5. 通知团队成员重新克隆仓库" -ForegroundColor Gray
} else {
    Write-Host "`n❌ 清理过程中出现错误，请检查输出信息" -ForegroundColor Red
    Write-Host "如果遇到问题，可以运行: git filter-branch --abort" -ForegroundColor Yellow
}
