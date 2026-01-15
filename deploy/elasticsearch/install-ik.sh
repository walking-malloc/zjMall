#!/bin/bash
# Elasticsearch IK 分词器安装脚本

# 进入 ES 容器
docker exec -it zjmall-elasticsearch bash

# 在容器内执行以下命令安装 IK 分词器
# 注意：版本需要与 ES 版本匹配（8.17.2）
elasticsearch-plugin install https://github.com/medcl/elasticsearch-analysis-ik/releases/download/v8.17.2/elasticsearch-analysis-ik-8.17.2.zip

# 安装完成后重启容器
# 退出容器后执行：
# docker restart zjmall-elasticsearch


