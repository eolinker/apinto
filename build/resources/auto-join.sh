#!/bin/sh

# 解析当前 Pod 的序号
CURRENT_INDEX=${HOSTNAME##*-}
BASE_NAME=${HOSTNAME%-*}
# 如果当前是 apinto-0，需要特殊处理
if [ "$CURRENT_INDEX" -eq 0 ]; then
  echo "This is ${HOSTNAME}. Checking if other nodes exist..."

  # 尝试连接 apinto-1
  NEXT_POD="${HOSTNAME}.${SERVICE}.${NAMESPACE}.svc.cluster.local"
  if nc -zv ${NEXT_POD} 9400; then
    echo "Found a running node: ${NEXT_POD}. Joining the cluster..."
    ./apinto join ${NEXT_POD}
  else
    echo "No other nodes are available. Starting as the first node."
  fi
else
  # 对于非 apinto-0 的 Pod，连接到前一个 Pod
  PREVIOUS_POD="${BASE_NAME}-$((CURRENT_INDEX - 1)).${SERVICE}.${NAMESPACE}.svc.cluster.local"
  echo "This is $HOSTNAME. Waiting for $PREVIOUS_POD to be ready..."

  until nc -zv $PREVIOUS_POD 9400; do
    echo "Waiting for $PREVIOUS_POD to be ready..."
    sleep 5
  done

  echo "$PREVIOUS_POD is ready. Joining the cluster..."
  ./apinto join $PREVIOUS_POD
fi

if [ $? -ne 0 ]; then
	echo "Error: Failed to join the cluster."
	exit 1
fi

echo "Successfully joined the cluster."

