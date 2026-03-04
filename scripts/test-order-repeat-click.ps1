# 重复下单测试：每次按回车发送一次相同的创建订单请求，用于验证防重复
# 用法：.\scripts\test-order-repeat-click.ps1

param(
    [string]$BaseUrl = $env:BASE_URL,
    [string]$Jwt     = $env:JWT
)

if (-not $BaseUrl) { $BaseUrl = "http://localhost:8086" }
if (-not $Jwt) {
    Write-Host "请先设置 JWT，例如：" -ForegroundColor Yellow
    Write-Host '  $env:JWT = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."; $env:BASE_URL = "http://localhost:8086"' -ForegroundColor Gray
    exit 1
}

# 请求体（与您提供的一致）
$body = @'
{
  "items": [
    {
      "cart_item_id": "01KHE5VJPK1DQPDT2BG9SKQCX2",
      "product_id": "01PRD00000000000000000002",
      "sku_id": "01SKU00000000000000000003",
      "quantity": 1
    }
  ],
  "address_id": "01KHDVEYQAFSJMK89VMQYGEP8X",
  "token": "f2b3856c-94f9-4936-ae9d-cd27f20a95b3"
}
'@

$headers = @{
    "Content-Type"  = "application/json"
    "Authorization" = if ($Jwt.StartsWith("Bearer ")) { $Jwt } else { "Bearer $Jwt" }
}

$count = 0
Write-Host "已使用固定 token，每次按回车将发送一次创建订单请求（同一 token）。按 Q 回车退出。`n" -ForegroundColor Cyan

while ($true) {
    $key = Read-Host "按 Enter 发送第 $($count + 1) 次请求（Q 退出）"
    if ($key -eq "Q" -or $key -eq "q") { break }

    $count++
    Write-Host "[$count] 发送中..." -NoNewline
    try {
        $resp = Invoke-RestMethod -Uri "$BaseUrl/api/v1/orders" -Method Post -Headers $headers -Body $body
        $color = if ($resp.code -eq 0) { "Green" } else { "Yellow" }
        Write-Host " code=$($resp.code) | message=$($resp.message) | order_no=$($resp.order_no)" -ForegroundColor $color
    } catch {
        Write-Host " 请求异常: $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host "`n共发送 $count 次。预期：第 1 次成功，后续为「Token已失效或已使用」或返回同一 order_no。" -ForegroundColor Cyan
