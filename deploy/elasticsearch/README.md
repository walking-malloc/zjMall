# Elasticsearch IK 分词器安装指南

## 方法 1：在运行中的容器安装（快速）

```bash
# 进入容器
docker exec -it zjmall-elasticsearch bash

# 安装 IK 分词器（版本需匹配 ES 8.17.2）
elasticsearch-plugin install https://github.com/medcl/elasticsearch-analysis-ik/releases/download/v8.17.2/elasticsearch-analysis-ik-8.17.2.zip

# 退出容器
exit

# 重启容器
docker restart zjmall-elasticsearch
```

## 方法 2：使用自定义镜像（推荐生产环境）

1. 构建自定义镜像：
```bash
docker build -t elasticsearch-ik:8.17.2 ./deploy/elasticsearch
```

2. 修改 `docker-compose.yml`：
```yaml
elasticsearch:
  image: elasticsearch-ik:8.17.2  # 使用自定义镜像
  # ... 其他配置
```

3. 重启服务：
```bash
docker-compose up -d elasticsearch
```

## 验证安装

```bash
# 检查插件列表
curl http://127.0.0.1:9200/_cat/plugins

# 应该看到：analysis-ik
```

## 测试分词效果

```bash
# 测试 IK 分词器
curl -X POST "http://127.0.0.1:9200/_analyze" -H 'Content-Type: application/json' -d'
{
  "analyzer": "ik_max_word",
  "text": "中华人民共和国"
}'
```

## 注意事项

1. **版本匹配**：IK 分词器版本必须与 ES 版本完全匹配
2. **重启必需**：安装插件后必须重启 ES 才能生效
3. **数据持久化**：插件会保存在容器中，如果删除容器需要重新安装
4. **生产环境**：建议使用方法 2（自定义镜像），避免每次重建容器都要重新安装

