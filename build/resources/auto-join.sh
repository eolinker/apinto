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
MAX_ATTEMPTS=60          # 最大尝试次数（节点等待），约1分钟
RETRY_INTERVAL=5         # 重试间隔，单位秒
MAX_JOIN_RETRIES=12      # 加入集群失败的最大重试次数
MAX_POD_INDEX=${MAX_POD_INDEX:-10}  # 默认检查最多10个Pod，可通过环境变量配置

# 加入集群成功后写入的标记文件，位于持久化卷（PVC）中，随节点数据一同持久化。
# 作用：区分「全新节点（需要 join）」和「重启/滚动升级恢复的节点（已是 etcd 成员，靠本地数据恢复，禁止再 join）」。
# 该文件在 leave（缩容）时被删除，与 etcd 数据同生命周期。
JOINED_FLAG=${JOINED_FLAG:-/var/lib/apinto/.joined}

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

# 关键：若本节点已经是集群成员（存在标记文件），说明这是重启 / 滚动升级 / 节点恢复。
# 此时 etcd 会依据本地持久化数据以 existing 状态自动恢复，绝不能再次 join，否则会破坏 etcd 成员配置导致 crash-loop。
if [ -f "$JOINED_FLAG" ]; then
    echo_info "This is $HOSTNAME. Already a cluster member (found $JOINED_FLAG). Recovering from local data, skip join."
    exit 0
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
    target_addr=$1
    join_retries=0
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

# 探测某个 peer 是否就绪，就绪后尝试 join；成功则写标记并返回 0
join_via_peer() {
    peer_host=$1
    attempt=0
    while [ $attempt -lt $MAX_ATTEMPTS ]; do
        if curl --max-time 5 --silent --fail http://${peer_host}:9401; then
            echo_info "Found a running node: ${peer_host}."
            if try_join_cluster "${peer_host}:9401"; then
                touch "$JOINED_FLAG"
                echo_info "Marked $HOSTNAME as cluster member ($JOINED_FLAG)."
                return 0
            fi
            echo_info "Failed to join via ${peer_host}. Trying next node..."
            return 1
        else
            echo_info "${peer_host} is not ready yet. Retrying in $RETRY_INTERVAL seconds..."
        fi
        sleep $RETRY_INTERVAL
        attempt=$((attempt + 1))
    done
    echo_info "Timeout waiting for ${peer_host} after $MAX_ATTEMPTS attempts."
    return 1
}

# 全新节点：遍历除自身外的所有 peer（不写死 apinto-0，leader 会漂移），谁健康就 join 谁。
echo_info "This is $HOSTNAME (new node). Trying to join an existing cluster node..."
for i in $(seq 0 "$MAX_POD_INDEX"); do
    if [ "$i" = "$CURRENT_INDEX" ]; then
        continue
    fi
    OTHER_POD="${BASE_NAME}-${i}.${SERVICE}.${NAMESPACE}.svc.cluster.local"
    if join_via_peer "$OTHER_POD"; then
        exit 0
    fi
done

# 没有任何可加入的节点。
if [ "$CURRENT_INDEX" -eq 0 ]; then
    # 仅 0 号节点、且首次部署（无标记文件）时，允许作为集群的初始（bootstrap）节点。
    # etcd 主进程会以自身为单节点集群启动，后续其他节点再 join 进来。
    echo_info "No other nodes are joinable. This is the first node ($HOSTNAME), bootstrapping a new cluster."
    touch "$JOINED_FLAG"
    echo_info "Marked $HOSTNAME as cluster member ($JOINED_FLAG)."
    exit 0
else
    # 非 0 号节点绝不自立新集群，否则会造成脑裂。交给 K8s 重启后重试，等待可加入的节点出现。
    echo_error "This is $HOSTNAME. No joinable node found and this is not the first node. Refusing to bootstrap a new cluster (avoid split-brain). Exiting to retry."
    exit 1
fi
