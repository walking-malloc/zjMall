# 初始化数据库并插入测试数据
# 使用方法: .\scripts\init-database.ps1

$mysqlUser = "root"
$mysqlPassword = "root123456"
$containerName = "zjmall-mysql"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "初始化数据库并插入测试数据" -ForegroundColor Cyan
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

# 定义需要执行的脚本顺序
# 第一阶段：数据库结构初始化（按依赖顺序）
# 注意：mysql-schema.sql 是 Nacos 配置表，需要先创建 nacos_config 数据库
$schemaScripts = @(
    @{Name="Nacos配置数据库结构"; File="mysql-schema.sql"; Database="nacos_config"},
    @{Name="用户服务数据库结构"; File="user-service.sql"},
    @{Name="商品服务数据库结构"; File="product-service.sql"},
    @{Name="库存服务数据库结构"; File="inventory-service.sql"},
    @{Name="订单服务数据库结构"; File="order-service.sql"},
    @{Name="订单Outbox表结构"; File="order-outbox.sql"},
    @{Name="支付服务数据库结构"; File="payment-service.sql"},
    @{Name="支付Outbox表结构"; File="payment-outbox.sql"},
    @{Name="购物车服务数据库结构"; File="cart-service.sql"},
    @{Name="促销服务数据库结构"; File="promotion-service.sql"}
)

# 第二阶段：测试数据（在结构创建后执行）
# 注意：如果数据已存在，会跳过插入（使用 INSERT IGNORE 或先删除再插入）
$dataScripts = @(
    @{Name="测试品牌数据"; File="insert-test-brands.sql"},
    @{Name="测试数据（商品、库存、支付渠道）"; File="insert-test-data.sql"; ClearFirst=$false}
)

$scriptPath = "deploy/mysql/init"

# 执行数据库结构脚本
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "第一阶段：创建数据库结构" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

foreach ($script in $schemaScripts) {
    $scriptFile = Join-Path $scriptPath $script.File
    
    if (-not (Test-Path $scriptFile)) {
        Write-Host "⚠️ 跳过: $($script.Name) - 文件不存在: $scriptFile" -ForegroundColor Yellow
        continue
    }
    
    Write-Host "执行: $($script.Name)..." -ForegroundColor Yellow
    
    # 如果指定了数据库，先创建数据库并选择
    if ($script.Database) {
        $createDbQuery = "CREATE DATABASE IF NOT EXISTS $($script.Database) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci; USE $($script.Database);"
        $dbResult = $createDbQuery | docker exec -i $containerName mysql -u"$mysqlUser" -p"$mysqlPassword" 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Host "  ⚠️ 创建数据库 $($script.Database) 失败，继续执行..." -ForegroundColor Yellow
        }
        # 执行脚本时指定数据库
        $result = Get-Content $scriptFile | docker exec -i $containerName mysql -u"$mysqlUser" -p"$mysqlPassword" $($script.Database) 2>&1
    } else {
        $result = Get-Content $scriptFile | docker exec -i $containerName mysql -u"$mysqlUser" -p"$mysqlPassword" 2>&1
    }
    
    # 过滤警告信息
    $errorOnly = $result | Where-Object { $_ -match "ERROR|error|Error" -and $_ -notmatch "Warning|warning" }
    
    if ($LASTEXITCODE -eq 0 -and [string]::IsNullOrWhiteSpace($errorOnly)) {
        Write-Host "  ✅ $($script.Name) 执行成功" -ForegroundColor Green
    } else {
        # 检查是否是警告（如数据库已存在）
        if ($result -match "already exists|Duplicate entry|already exist|Duplicate key") {
            Write-Host "  ⚠️ $($script.Name) 执行完成（部分对象已存在）" -ForegroundColor Yellow
        } else {
            Write-Host "  ❌ $($script.Name) 执行失败" -ForegroundColor Red
            if ($errorOnly) {
                Write-Host "  错误信息: $errorOnly" -ForegroundColor Red
            }
        }
    }
    Write-Host ""
}

# 执行测试数据脚本
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "第二阶段：插入测试数据" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

foreach ($script in $dataScripts) {
    $scriptFile = Join-Path $scriptPath $script.File
    
    if (-not (Test-Path $scriptFile)) {
        Write-Host "⚠️ 跳过: $($script.Name) - 文件不存在: $scriptFile" -ForegroundColor Yellow
        continue
    }
    
    Write-Host "执行: $($script.Name)..." -ForegroundColor Yellow
    
    # 如果需要先清理数据
    if ($script.ClearFirst) {
        Write-Host "  清理旧数据..." -ForegroundColor Gray
        # 这里可以添加清理逻辑，但通常不需要，因为测试数据脚本应该使用 INSERT IGNORE
    }
    
    $result = Get-Content $scriptFile | docker exec -i $containerName mysql -u"$mysqlUser" -p"$mysqlPassword" 2>&1
    
    # 过滤警告信息，只保留真正的错误
    $errorOnly = $result | Where-Object { 
        $_ -match "ERROR" -and 
        $_ -notmatch "Warning|warning" -and
        $_ -notmatch "Duplicate entry.*for key.*PRIMARY" -and  # 主键重复可以忽略（数据已存在）
        $_ -notmatch "Duplicate entry.*for key.*UNIQUE"         # 唯一键重复可以忽略（数据已存在）
    }
    
    if ($LASTEXITCODE -eq 0 -and [string]::IsNullOrWhiteSpace($errorOnly)) {
        Write-Host "  ✅ $($script.Name) 执行成功" -ForegroundColor Green
    } else {
        # 检查是否是重复键错误（数据已存在，可以忽略）
        if ($result -match "Duplicate entry.*for key") {
            Write-Host "  ⚠️ $($script.Name) 执行完成（部分数据已存在，已跳过重复项）" -ForegroundColor Yellow
        } else {
            Write-Host "  ❌ $($script.Name) 执行失败" -ForegroundColor Red
            if ($errorOnly) {
                Write-Host "  错误信息: $errorOnly" -ForegroundColor Red
            } elseif ($result -match "ERROR") {
                # 显示所有错误
                $allErrors = $result | Where-Object { $_ -match "ERROR" }
                Write-Host "  错误信息: $allErrors" -ForegroundColor Red
            }
        }
    }
    Write-Host ""
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "数据库初始化完成" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "提示: 运行以下命令验证数据:" -ForegroundColor Yellow
Write-Host "  .\scripts\verify-test-data.ps1" -ForegroundColor Cyan
