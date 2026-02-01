#!/bin/bash
# IK 分词器安装脚本

ES_VERSION="8.17.2"
CONTAINER_NAME="zjmall-elasticsearch"

echo "正在停止 Elasticsearch 容器..."
docker-compose stop elasticsearch

echo "正在下载 IK 分词器插件..."
# 尝试多个可能的版本和 URL
URLS=(
    "https://github.com/infinilabs/analysis-ik/releases/download/v8.17.2/elasticsearch-analysis-ik-8.17.2.zip"
    "https://github.com/medcl/elasticsearch-analysis-ik/releases/download/v8.17.2/elasticsearch-analysis-ik-8.17.2.zip"
    "https://github.com/infinilabs/analysis-ik/releases/download/v8.17.0/elasticsearch-analysis-ik-8.17.0.zip"
    "https://github.com/medcl/elasticsearch-analysis-ik/releases/download/v8.17.0/elasticsearch-analysis-ik-8.17.0.zip"
)

for URL in "${URLS[@]}"; do
    echo "尝试下载: $URL"
    if docker run --rm --entrypoint="" elasticsearch:8.17.2 bash -c "curl -L -f $URL -o /tmp/ik.zip 2>/dev/null && echo 'SUCCESS'"; then
        echo "✅ 下载成功！"
        docker run --rm --entrypoint="" -v zjmall_elasticsearch_data:/usr/share/elasticsearch/data elasticsearch:8.17.2 bash -c "cd /usr/share/elasticsearch/data && mkdir -p plugins/ik && cd plugins/ik && unzip -q /tmp/ik.zip && rm -f /tmp/ik.zip"
        break
    else
        echo "❌ 下载失败，尝试下一个 URL..."
    fi
done

echo "正在启动 Elasticsearch 容器..."
docker-compose start elasticsearch

echo "等待 ES 启动..."
sleep 10

echo "验证插件安装..."
docker exec $CONTAINER_NAME elasticsearch-plugin list




