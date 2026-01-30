# 创建 ES 索引脚本
$indexBody = @'
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 1,
    "analysis": {
      "analyzer": {
        "ik_max_word": {
          "type": "standard"
        },
        "ik_smart": {
          "type": "standard"
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "id": {
        "type": "keyword"
      },
      "title": {
        "type": "text",
        "analyzer": "standard",
        "search_analyzer": "standard",
        "fields": {
          "keyword": {
            "type": "keyword"
          }
        }
      },
      "subtitle": {
        "type": "text",
        "analyzer": "standard",
        "search_analyzer": "standard"
      },
      "description": {
        "type": "text",
        "analyzer": "standard",
        "search_analyzer": "standard"
      },
      "category_id": {
        "type": "keyword"
      },
      "category_name": {
        "type": "text",
        "analyzer": "standard"
      },
      "brand_id": {
        "type": "keyword"
      },
      "brand_name": {
        "type": "text",
        "analyzer": "standard"
      },
      "tags": {
        "type": "keyword"
      },
      "skus": {
        "type": "nested",
        "properties": {
          "sku_name": {
            "type": "keyword"
          },
          "price": {
            "type": "float"
          }
        }
      },
      "attribute_values": {
        "type": "keyword"
      },
      "attributes": {
        "type": "nested",
        "properties": {
          "attribute_id": {
            "type": "keyword"
          },
          "attribute_name": {
            "type": "text",
            "analyzer": "standard"
          },
          "value": {
            "type": "text",
            "analyzer": "standard"
          }
        }
      },
      "status": {
        "type": "byte"
      },
      "on_shelf_time": {
        "type": "date"
      },
      "created_at": {
        "type": "date"
      },
      "updated_at": {
        "type": "date"
      }
    }
  }
}
'@

try {
    $response = Invoke-RestMethod -Uri "http://127.0.0.1:9200/products" -Method PUT -Body $indexBody -ContentType "application/json"
    Write-Host "✅ 索引创建成功！" -ForegroundColor Green
    Write-Host ($response | ConvertTo-Json -Depth 10)
} catch {
    if ($_.Exception.Response.StatusCode -eq 400) {
        $errorContent = $_.ErrorDetails.Message | ConvertFrom-Json
        if ($errorContent.error.reason -like "*already exists*") {
            Write-Host "✅ 索引已存在" -ForegroundColor Yellow
        } else {
            Write-Host "❌ 创建索引失败: $($errorContent.error.reason)" -ForegroundColor Red
            Write-Host "尝试使用 standard 分词器..." -ForegroundColor Yellow
            
            # 如果 IK 分词器不存在，使用 standard 分词器重试
            $indexBodyStandard = $indexBody -replace '"type": "ik_max_word"', '"type": "standard"' -replace '"type": "ik_smart"', '"type": "standard"'
            $indexBodyStandard = $indexBodyStandard -replace '"analyzer": "ik_max_word"', '"analyzer": "standard"' -replace '"analyzer": "ik_smart"', '"analyzer": "standard"'
            $indexBodyStandard = $indexBodyStandard -replace '"search_analyzer": "ik_smart"', '"search_analyzer": "standard"'
            
            try {
                $response = Invoke-RestMethod -Uri "http://127.0.0.1:9200/products" -Method PUT -Body $indexBodyStandard -ContentType "application/json"
                Write-Host "✅ 使用 standard 分词器创建索引成功！" -ForegroundColor Green
            } catch {
                Write-Host "❌ 创建索引失败: $($_.Exception.Message)" -ForegroundColor Red
            }
        }
    } else {
        Write-Host "❌ 错误: $($_.Exception.Message)" -ForegroundColor Red
    }
}



