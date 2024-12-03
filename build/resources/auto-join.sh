#!/bin/sh

ERR_LOG=/var/log/apinto/error.log
echo_info() {
	echo "[$(date "+%Y-%m-%d %H:%M:%S")] [INFO] $1" >> $ERR_LOG
}

echo_error() {
	echo "[$(date "+%Y-%m-%d %H:%M:%S")] [ERROR] $1" >> $ERR_LOG
}

# 解析当前 Pod 的序号
CURRENT_INDEX=${HOSTNAME##*-}
BASE_NAME=${HOSTNAME%-*}

# 检查当前程序是否启动，若无，则等待
until curl --max-time 5 --silent --fail http://127.0.0.1:9400; do
	echo_info "Waiting for localhost to be ready..."
	sleep 1
done

# 如果当前是 apinto-0，需要特殊处理
if [ "$CURRENT_INDEX" -eq 0 ]; then
  echo_info "This is ${HOSTNAME}. Checking if other nodes exist..."

  # 尝试连接 apinto-1
  NEXT_POD="${BASE_NAME}-$((CURRENT_INDEX + 1)).${SERVICE}.${NAMESPACE}.svc.cluster.local"
  if curl --max-time 5 --silent --fail http://${NEXT_POD}:9401; then
    echo_info "Found a running node: ${NEXT_POD}. Joining the cluster..."
    ./apinto join -addr ${NEXT_POD}:9401 >> $ERR_LOG 2>&1
    if [ $? -ne 0 ]; then
			echo_error "Error: Failed to join the cluster."
		fi
  else
    echo_info "No other nodes are available. Starting as the first node."
  fi
else
  # 对于非 apinto-0 的 Pod，连接到前一个 Pod
  PREVIOUS_POD="${BASE_NAME}-$((CURRENT_INDEX - 1)).${SERVICE}.${NAMESPACE}.svc.cluster.local"
  echo_info "This is $HOSTNAME. Waiting for $PREVIOUS_POD to be ready..."

  until curl --max-time 5 --silent --fail http://$PREVIOUS_POD:9401; do
    echo_info "Waiting for $PREVIOUS_POD to be ready..."
    sleep 1
  done

  echo_info "$PREVIOUS_POD is ready. Joining the cluster..."
  ./apinto join -addr $PREVIOUS_POD:9401 >> $ERR_LOG 2>&1
  if [ $? -ne 0 ]; then
		echo_error "Error: Failed to join the cluster."
	fi
fi

if [ $? -ne 0 ]; then
	echo_error "Error: Failed to join the cluster."
	exit 1
fi

echo_info "Successfully joined the cluster."

