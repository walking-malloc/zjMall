# 验证测试数据脚本
# 使用方法: .\scripts\verify-test-data.ps1

$mysqlUser = "root"
$mysqlPassword = "root123456"
$containerName = "zjmall-mysql"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "验证测试数据" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 检查 MySQL 容器是否运行
Write-Host "检查 MySQL 容器状态..." -ForegroundColor Yellow
$containerStatus = docker ps --filter "name=$containerName" --format "{{.Status}}" 2>&1
if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($containerStatus)) {
    Write-Host "❌ MySQL 容器未运行，请先启动容器" -ForegroundColor Red
    Write-Host "提示: docker-compose up -d mysql" -ForegroundColor Yellow
    exit 1
}
Write-Host "✅ MySQL 容器运行中: $containerStatus" -ForegroundColor Green
Write-Host ""

# 检查 MySQL 连接（使用 docker exec）
Write-Host "检查 MySQL 连接..." -ForegroundColor Yellow
$testQuery = "SELECT 1;"
$testOutput = docker exec $containerName mysql -u"$mysqlUser" -p"$mysqlPassword" -e "$testQuery" 2>&1

if ($testOutput -match "1" -and -not ($testOutput -match "ERROR|error|Error")) {
    Write-Host "✅ MySQL 连接成功" -ForegroundColor Green
} else {
    Write-Host "❌ MySQL 连接失败" -ForegroundColor Red
    Write-Host "错误信息: $testOutput" -ForegroundColor Red
    Write-Host ""
    Write-Host "提示: 使用以下命令手动验证:" -ForegroundColor Yellow
    Write-Host "  docker exec -it $containerName mysql -u$mysqlUser -p$mysqlPassword" -ForegroundColor Cyan
    exit 1
}
Write-Host ""

# 验证商品数据
Write-Host "验证商品数据..." -ForegroundColor Yellow
$productQuery = @"
USE product_db;
SELECT 
    (SELECT COUNT(*) FROM categories) as category_count,
    (SELECT COUNT(*) FROM products WHERE status = 3) as product_count,
    (SELECT COUNT(*) FROM skus WHERE status = 1) as sku_count,
    (SELECT COUNT(*) FROM attributes) as attribute_count,
    (SELECT COUNT(*) FROM attribute_values) as attribute_value_count,
    (SELECT COUNT(*) FROM sku_attributes) as sku_attribute_count;
"@

$productResult = $productQuery | docker exec -i $containerName mysql -u"$mysqlUser" -p"$mysqlPassword" -N 2>&1 | Where-Object { $_ -notmatch "Warning|warning|mysql:" }

if ($LASTEXITCODE -eq 0 -and -not ($productResult -match "ERROR|error|Error")) {
    $values = ($productResult | Where-Object { $_ -match "^\d+" }) -split '\s+'
    Write-Host "  ✅ 类目数量: $($values[0])" -ForegroundColor Green
    Write-Host "  ✅ 已上架商品数: $($values[1])" -ForegroundColor Green
    Write-Host "  ✅ SKU数量: $($values[2])" -ForegroundColor Green
    Write-Host "  ✅ 属性数量: $($values[3])" -ForegroundColor Green
    Write-Host "  ✅ 属性值数量: $($values[4])" -ForegroundColor Green
    Write-Host "  ✅ SKU属性关联数: $($values[5])" -ForegroundColor Green
} else {
    Write-Host "  ❌ 查询商品数据失败" -ForegroundColor Red
    if ($productResult) {
        Write-Host "  错误信息: $productResult" -ForegroundColor Red
    }
}

Write-Host ""

# 验证库存数据
Write-Host "验证库存数据..." -ForegroundColor Yellow
$inventoryQuery = @"
USE inventory_db;
SELECT 
    COUNT(*) as stock_count,
    COALESCE(SUM(available_stock), 0) as total_stock
FROM inventory_stocks;
"@

$inventoryResult = $inventoryQuery | docker exec -i $containerName mysql -u"$mysqlUser" -p"$mysqlPassword" -N 2>&1 | Where-Object { $_ -notmatch "Warning|warning|mysql:" }

if ($LASTEXITCODE -eq 0 -and -not ($inventoryResult -match "ERROR|error|Error")) {
    $values = ($inventoryResult | Where-Object { $_ -match "^\d+" }) -split '\s+'
    Write-Host "  ✅ 库存记录数: $($values[0])" -ForegroundColor Green
    Write-Host "  ✅ 总库存数: $($values[1])" -ForegroundColor Green
} else {
    Write-Host "  ❌ 查询库存数据失败" -ForegroundColor Red
    if ($inventoryResult) {
        Write-Host "  错误信息: $inventoryResult" -ForegroundColor Red
    }
}

Write-Host ""

# 验证支付渠道数据
Write-Host "验证支付渠道数据..." -ForegroundColor Yellow
$paymentQuery = @"
USE payment_db;
SELECT 
    COUNT(*) as channel_count,
    SUM(CASE WHEN is_enabled = 1 THEN 1 ELSE 0 END) as enabled_count
FROM payment_channels;
"@

$paymentResult = $paymentQuery | docker exec -i $containerName mysql -u"$mysqlUser" -p"$mysqlPassword" -N 2>&1 | Where-Object { $_ -notmatch "Warning|warning|mysql:" }

if ($LASTEXITCODE -eq 0 -and -not ($paymentResult -match "ERROR|error|Error")) {
    $values = ($paymentResult | Where-Object { $_ -match "^\d+" }) -split '\s+'
    Write-Host "  ✅ 支付渠道总数: $($values[0])" -ForegroundColor Green
    Write-Host "  ✅ 启用渠道数: $($values[1])" -ForegroundColor Green
} else {
    Write-Host "  ❌ 查询支付渠道数据失败" -ForegroundColor Red
    if ($paymentResult) {
        Write-Host "  错误信息: $paymentResult" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "验证完成" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "提示: 如果数据为空，请先执行测试数据脚本:" -ForegroundColor Yellow
Write-Host "  docker exec -i $containerName mysql -u$mysqlUser -p$mysqlPassword < deploy/mysql/init/insert-test-data.sql" -ForegroundColor Cyan
