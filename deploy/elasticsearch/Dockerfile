FROM elasticsearch:8.17.2

# 安装 IK 分词器插件
# 注意：需要先停止 ES 才能安装插件
RUN elasticsearch-plugin install --batch https://github.com/infinilabs/analysis-ik/releases/download/v8.17.2/elasticsearch-analysis-ik-8.17.2.zip || \
    elasticsearch-plugin install --batch https://github.com/medcl/elasticsearch-analysis-ik/releases/download/v8.17.2/elasticsearch-analysis-ik-8.17.2.zip || \
    echo "IK plugin installation failed, will use standard analyzer"
