#!/bin/bash

set -e  # 在命令失败时退出

# 定义日志路径
APINTO_LOG_DIR="/var/log/apinto"
APINTO_LOG_FILE="$APINTO_LOG_DIR/error.log"

if [[ "${HOSTNAME}" != "" && ${SERVICE} != "" && ${NAMESPACE} != "" ]];then
	# 替换配置文件中的 {IP}
	CONFIG_FILE="/etc/apinto/config.yml"
	IP=${HOSTNAME}.${SERVICE}.${NAMESPACE}.svc.cluster.local
	cp -f config.yml.k8s.tpl ${CONFIG_FILE}
	sed -i "s/{IP}/${IP}/g" "$CONFIG_FILE"
  echo "Replaced {IP} with ${IP} in $CONFIG_FILE."
fi

./apinto stop

sleep 5s

# 启动 Apinto
echo "Starting Apinto..."
./apinto start

# 等待 Apinto 启动完成
echo "Waiting for Apinto to start..."
MAX_RETRIES=10  # 最大重试次数
SLEEP_INTERVAL=5  # 每次重试间隔秒数

for ((i=1; i<=MAX_RETRIES; i++)); do
    if curl --max-time 5 --silent --fail http://127.0.0.1:9400; then  # 替换为 Apinto 的监听端口
        echo "Apinto started successfully."
        break
    else
        echo "Attempt $i: Apinto is not ready yet, retrying in $SLEEP_INTERVAL seconds..."
        sleep $SLEEP_INTERVAL
    fi

    if [ $i -eq $MAX_RETRIES ]; then
        echo "Error: Apinto failed to start after $MAX_RETRIES attempts."
        exit 1
    fi
done

# 动态跟踪日志文件并输出到 Docker 容器日志
echo "Redirecting Apinto logs to Docker output..."
tail -F "$APINTO_LOG_FILE"
