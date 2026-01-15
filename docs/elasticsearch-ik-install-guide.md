# Elasticsearch IK 分词器安装指南

## 当前状态

由于 Elasticsearch 8.17.2 版本的 IK 分词器插件在 GitHub 上暂时没有对应的版本，当前系统使用 **standard 分词器**（ES 内置），可以正常使用，但中文分词效果不如 IK 分词器。

## 后续安装 IK 分词器的方法

### 方法 1：等待官方发布对应版本（推荐）

当 GitHub 上发布了 ES 8.17.2 对应的 IK 分词器版本后，可以按以下步骤安装：

```bash
# 1. 停止 ES 容器
docker-compose stop elasticsearch

# 2. 进入容器安装插件
docker exec -it zjmall-elasticsearch bash
elasticsearch-plugin install https://github.com/infinilabs/analysis-ik/releases/download/v8.17.2/elasticsearch-analysis-ik-8.17.2.zip
exit

# 3. 重启容器
docker-compose restart elasticsearch

# 4. 验证安装
docker exec zjmall-elasticsearch elasticsearch-plugin list
```

### 方法 2：使用兼容版本（如果可用）

如果 ES 8.17.2 与某个较低版本的 IK 分词器兼容，可以尝试安装：

```bash
# 尝试安装 8.17.0 或 8.16.x 版本的 IK 分词器
docker exec -it zjmall-elasticsearch bash
elasticsearch-plugin install https://github.com/infinilabs/analysis-ik/releases/download/v8.17.0/elasticsearch-analysis-ik-8.17.0.zip
```

**注意**：需要确保版本兼容，否则可能导致 ES 启动失败。

### 方法 3：使用自定义 Dockerfile（生产环境推荐）

创建一个包含 IK 分词器的自定义镜像：

1. 创建 `deploy/elasticsearch/Dockerfile`：
```dockerfile
FROM elasticsearch:8.17.2

# 安装 IK 分词器（当有对应版本时）
RUN elasticsearch-plugin install --batch \
    https://github.com/infinilabs/analysis-ik/releases/download/v8.17.2/elasticsearch-analysis-ik-8.17.2.zip || \
    echo "IK plugin installation failed"
```

2. 修改 `docker-compose.yml`：
```yaml
elasticsearch:
  build:
    context: ./deploy/elasticsearch
    dockerfile: Dockerfile
  # ... 其他配置
```

3. 重新构建并启动：
```bash
docker-compose build elasticsearch
docker-compose up -d elasticsearch
```

## 安装 IK 分词器后的操作

安装 IK 分词器后，需要：

1. **删除旧索引**（如果已创建）：
```bash
curl -X DELETE "http://127.0.0.1:9200/products"
```

2. **重启 product-service**，让服务使用 IK 分词器重新创建索引

3. **重新同步商品数据**到 ES（如果有已上架的商品）

## 检查 IK 分词器是否可用

```bash
# 查看已安装的插件
docker exec zjmall-elasticsearch elasticsearch-plugin list

# 应该看到：analysis-ik

# 测试 IK 分词效果
curl -X POST "http://127.0.0.1:9200/_analyze" -H 'Content-Type: application/json' -d'
{
  "analyzer": "ik_max_word",
  "text": "中华人民共和国"
}'
```

## 当前使用的分词器

- **analyzer**: `standard`（ES 内置标准分词器）
- **search_analyzer**: `standard`
- **特点**: 可以正常使用，但中文分词效果不如 IK 分词器

## 参考链接

- IK 分词器 GitHub: https://github.com/infinilabs/analysis-ik
- IK 分词器 Releases: https://github.com/infinilabs/analysis-ik/releases
- Elasticsearch 插件文档: https://www.elastic.co/guide/en/elasticsearch/plugins/current/index.html


