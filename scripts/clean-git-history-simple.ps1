# 从 Git 历史记录中完全删除不应该被跟踪的文件
# ⚠️ 警告：这是破坏性操作，会重写 Git 历史！
# 
# 使用方法: 
#   1. 确保当前工作目录干净: git status
#   2. 运行: .\scripts\clean-git-history-simple.ps1
#   3. 完成后强制推送: git push --force --all

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Git 历史记录清理工具" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

Write-Host "⚠️  重要警告：" -ForegroundColor Red
Write-Host "  1. 此操作将重写 Git 历史记录" -ForegroundColor Yellow
Write-Host "  2. 执行后需要强制推送到远程: git push --force --all" -ForegroundColor Yellow
Write-Host "  3. 所有团队成员需要重新克隆仓库" -ForegroundColor Yellow
Write-Host "  4. 建议先备份仓库" -ForegroundColor Yellow
Write-Host ""

# 检查工作目录是否干净
$status = git status --porcelain
if ($status) {
    Write-Host "❌ 错误：工作目录不干净，请先提交或暂存更改" -ForegroundColor Red
    Write-Host "当前状态：" -ForegroundColor Yellow
    git status --short
    exit 1
}

Write-Host "✅ 工作目录干净" -ForegroundColor Green
Write-Host ""

$confirm = Read-Host "确认要继续清理历史记录吗？(输入 yes 继续)"

if ($confirm -ne "yes") {
    Write-Host "操作已取消" -ForegroundColor Yellow
    exit
}

Write-Host "`n开始清理 Git 历史记录..." -ForegroundColor Cyan
Write-Host "这可能需要几分钟时间，请耐心等待...`n" -ForegroundColor Yellow

# 要删除的文件和目录（使用相对于仓库根目录的路径）
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

Write-Host "将删除以下文件和目录：" -ForegroundColor Cyan
foreach ($item in $itemsToRemove) {
    Write-Host "  - $item" -ForegroundColor Gray
}
Write-Host ""

# 构建 filter-branch 命令
# 注意：在 PowerShell 中需要转义特殊字符
$removeCommands = ""
foreach ($item in $itemsToRemove) {
    $removeCommands += "git rm -rf --cached --ignore-unmatch `"$item`" 2>nul; "
}

Write-Host "执行 git filter-branch..." -ForegroundColor Yellow
Write-Host "(如果看到 'WARNING: Ref 'refs/heads/xxx' is unchanged'，这是正常的)" -ForegroundColor Gray
Write-Host ""

# 执行清理
git filter-branch --force --index-filter $removeCommands --prune-empty --tag-name-filter cat -- --all

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n✅ Git 历史记录清理完成！" -ForegroundColor Green
    Write-Host ""
    
    Write-Host "清理备份引用（释放空间）：" -ForegroundColor Yellow
    git for-each-ref --format='delete %(refname)' refs/original | git update-ref --stdin
    git reflog expire --expire=now --all
    git gc --prune=now --aggressive
    
    Write-Host "`n下一步操作：" -ForegroundColor Yellow
    Write-Host "1. 验证文件已删除: git log --all --full-history -- '*.exe' | Select-Object -First 5" -ForegroundColor Gray
    Write-Host "2. 强制推送到远程: git push --force --all" -ForegroundColor Gray
    Write-Host "3. 强制推送标签: git push --force --tags" -ForegroundColor Gray
    Write-Host "4. 通知团队成员重新克隆仓库" -ForegroundColor Gray
} else {
    Write-Host "`n❌ 清理过程中出现错误" -ForegroundColor Red
    Write-Host "如果遇到问题，可以运行: git filter-branch --abort" -ForegroundColor Yellow
}
