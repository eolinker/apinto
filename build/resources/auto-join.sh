#!/bin/sh

ERR_LOG=/var/log/apinto/error.log
echo_info() {
    echo "[$(date "+%Y-%m-%d %H:%M:%S")] [INFO] $1" >> $ERR_LOG
}

echo_error() {
    echo "[$(date "+%Y-%m-%d %H:%M:%S")] [ERROR] $1" >> $ERR_LOG
}

# 检查环境变量
if [ -z "$SERVICE" ] || [ -z "$NAMESPACE" ]; then
    echo_error "Environment variables SERVICE and NAMESPACE must be set."
    exit 1
fi

# 解析当前 Pod 的序号
CURRENT_INDEX=${HOSTNAME##*-}
BASE_NAME=${HOSTNAME%-*}
MAX_ATTEMPTS=60  # 最大尝试次数（节点等待），约1分钟
RETRY_INTERVAL=5  # 重试间隔，单位秒
MAX_JOIN_RETRIES=12  # 加入集群失败的最大重试次数
MAX_POD_INDEX=${MAX_POD_INDEX:-10}  # 默认检查最多10个Pod，可通过环境变量配置

# 等待本地服务启动
attempt=0
until curl --max-time 5 --silent --fail http://127.0.0.1:9400 || [ $attempt -ge $MAX_ATTEMPTS ]; do
    echo_info "Waiting for localhost to be ready... Attempt $attempt"
    sleep 1
    attempt=$((attempt + 1))
done
if [ $attempt -ge $MAX_ATTEMPTS ]; then
    echo_error "Timeout waiting for localhost to be ready after $MAX_ATTEMPTS attempts."
    exit 1
fi

# 检查是否成功加入集群
check_cluster_join() {
    info_output=$(./apinto info 2>&1)
    peer_count=$(echo "$info_output" | grep -c -- "--Peer")
    if [ "$peer_count" -ge 2 ]; then
        echo_info "Successfully joined the cluster with $peer_count peers. Cluster info: $info_output"
        return 0
    else
        echo_info "Failed to join the cluster. Only $peer_count peer(s) found. Info: $info_output"
        return 1
    fi
}

# 尝试加入集群，带重试
try_join_cluster() {
    local target_addr=$1
    local join_retries=0
    while [ $join_retries -lt $MAX_JOIN_RETRIES ]; do
        echo_info "Attempting to join cluster via $target_addr (Retry $join_retries/$MAX_JOIN_RETRIES)..."
        join_output=$(./apinto join -addr "$target_addr" 2>&1)
        if [ $? -eq 0 ]; then
            if check_cluster_join; then
                return 0
            else
                echo_info "Join via $target_addr executed but cluster validation failed.Details: $join_output"
            fi
        else
            echo_info "Join via $target_addr failed. Details: $join_output"
        fi
        join_retries=$((join_retries + 1))
        if [ $join_retries -lt $MAX_JOIN_RETRIES ]; then
            echo_info "Retrying join in $RETRY_INTERVAL seconds..."
            sleep $RETRY_INTERVAL
        fi
    done
    echo_error "Failed to join cluster via $target_addr after $MAX_JOIN_RETRIES retries."
    return 1
}

if [ "$CURRENT_INDEX" -eq 0 ]; then
    # apinto-0: 检查其他 Pod 是否在运行
    echo_info "This is $HOSTNAME. Checking if other nodes are running..."
    for i in $(seq 1 "$MAX_POD_INDEX"); do
        OTHER_POD="${BASE_NAME}-${i}.${SERVICE}.${NAMESPACE}.svc.cluster.local"
        attempt=0
        while [ $attempt -lt $MAX_ATTEMPTS ]; do
            if curl --max-time 5 --silent --fail http://${OTHER_POD}:9401; then
                echo_info "Found a running node: ${OTHER_POD}."
                if try_join_cluster "${OTHER_POD}:9401"; then
                    exit 0
                fi
                echo_info "Failed to join via ${OTHER_POD}. Trying next node..."
                break
            else
                echo_info "${OTHER_POD} is not ready yet. Retrying in $RETRY_INTERVAL seconds..."
            fi
            sleep $RETRY_INTERVAL
            attempt=$((attempt + 1))
        done
        echo_info "Timeout waiting for ${OTHER_POD} after $MAX_ATTEMPTS attempts."
    done
    echo_info "No other nodes are available or joinable. Starting as the first node."
else
    # 非 apinto-0 的 Pod，加入 apinto-0
    LEADER_POD="${BASE_NAME}-0.${SERVICE}.${NAMESPACE}.svc.cluster.local"
    echo_info "This is $HOSTNAME. Waiting for $LEADER_POD to be ready..."
    attempt=0
    until curl --max-time 5 --silent --fail http://$LEADER_POD:9401 || [ $attempt -ge $MAX_ATTEMPTS ]; do
        echo_info "Waiting for $LEADER_POD to be ready... Attempt $attempt"
        sleep 1
        attempt=$((attempt + 1))
    done
    if [ $attempt -ge $MAX_ATTEMPTS ]; then
        echo_error "Timeout waiting for $LEADER_POD to be ready after $MAX_ATTEMPTS attempts."
        exit 1
    fi
    echo_info "$LEADER_POD is ready."
    if try_join_cluster "$LEADER_POD:9401"; then
        exit 0
    else
        echo_error "Failed to join cluster via $LEADER_POD after $MAX_JOIN_RETRIES retries."
        exit 1
    fi
fi