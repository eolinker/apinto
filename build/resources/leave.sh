#!/bin/bash

ERR_LOG=/var/log/apinto/error.log
echo_info() {
    echo "[$(date "+%Y-%m-%d %H:%M:%S")] [INFO] $1" >> $ERR_LOG
}

echo_error() {
    echo "[$(date "+%Y-%m-%d %H:%M:%S")] [ERROR] $1" >> $ERR_LOG
}

# 该脚本由 preStop 触发，会在「滚动升级、节点重启/驱逐、OOMKill、缩容」等所有停止场景下执行。
# 关键：只有「真正缩容（本节点被永久移除）」时才允许 leave（把自己从 etcd 集群移除）。
# 其余场景（滚动升级、重启）必须保留 etcd 成员身份，靠 PVC 中的持久化数据恢复，绝不能 leave，
# 否则成员被移除后本地数据作废，重启将 crash-loop。
#
# 缩容判定：StatefulSet 缩容总是从高序号往低序号删。若本节点序号 >= 期望副本数(spec.replicas)，说明本节点被缩掉。
# 为兼容 HPA / 自动扩缩容（副本数动态变化，静态环境变量无法反映实时值），
# 期望副本数「实时查询 K8s API」获取；查询失败时才降级使用静态环境变量 REPLICAS。

# 加入集群的标记文件，与 auto-join.sh 保持一致。
JOINED_FLAG=${JOINED_FLAG:-/var/lib/apinto/.joined}

# 解析当前 Pod 的序号
CURRENT_INDEX=${HOSTNAME##*-}

# StatefulSet 名称，默认取 SERVICE（与部署一致）。
STS_NAME=${STS_NAME:-$SERVICE}

# 通过 Pod 内置 ServiceAccount 实时查询 StatefulSet 的 spec.replicas。
# 成功则把副本数打印到 stdout 并返回 0；失败返回非 0。
get_desired_replicas() {
    sa_dir=/var/run/secrets/kubernetes.io/serviceaccount
    token_file=$sa_dir/token
    ca_file=$sa_dir/ca.crt
    ns=$NAMESPACE
    [ -z "$ns" ] && [ -f "$sa_dir/namespace" ] && ns=$(cat "$sa_dir/namespace")

    if [ ! -f "$token_file" ] || [ -z "$ns" ] || [ -z "$STS_NAME" ]; then
        return 1
    fi

    api="https://kubernetes.default.svc/apis/apps/v1/namespaces/${ns}/statefulsets/${STS_NAME}/scale"
    resp=$(curl --max-time 5 --silent --fail \
        --cacert "$ca_file" \
        -H "Authorization: Bearer $(cat "$token_file")" \
        "$api" 2>/dev/null)
    if [ $? -ne 0 ] || [ -z "$resp" ]; then
        return 1
    fi

    # scale 子资源的 spec 中只有 replicas 一个字段；截取 "status" 之前的部分，避免误取 status.replicas。
    spec_part=${resp%%\"status\"*}
    replicas=$(echo "$spec_part" | grep -o '"replicas"[[:space:]]*:[[:space:]]*[0-9]*' | grep -o '[0-9]*' | head -1)
    if [ -z "$replicas" ]; then
        return 1
    fi
    echo "$replicas"
    return 0
}

# 停止本地进程并退出（保留 etcd 成员身份）
stop_keep_membership() {
    ./apinto stop
    echo_info "Apinto stopped (membership preserved)."
    exit 0
}

# 获取期望副本数：优先实时 API，失败降级静态环境变量 REPLICAS。
DESIRED_REPLICAS=$(get_desired_replicas)
if [ -n "$DESIRED_REPLICAS" ]; then
    echo_info "Desired replicas from K8s API: $DESIRED_REPLICAS."
else
    DESIRED_REPLICAS=$REPLICAS
    if [ -n "$DESIRED_REPLICAS" ]; then
        echo_info "K8s API query failed, fallback to static REPLICAS=$DESIRED_REPLICAS."
    fi
fi

# 无法获得期望副本数时，为避免误删成员，保守按「重启」处理，不 leave。
if [ -z "$DESIRED_REPLICAS" ]; then
    echo_info "Cannot determine desired replicas (API failed and REPLICAS unset). Skip leave to protect cluster membership."
    stop_keep_membership
fi

if [ "$CURRENT_INDEX" -lt "$DESIRED_REPLICAS" ]; then
    # 序号在期望副本范围内 -> 这是滚动升级 / 重启，不能 leave。
    echo_info "This is $HOSTNAME (index $CURRENT_INDEX < replicas $DESIRED_REPLICAS). Rolling update or restart, keep membership, skip leave."
    stop_keep_membership
fi

# 序号 >= 期望副本数 -> 本节点被缩容移除，执行 leave 并清理标记文件（连同数据一起作废）。
echo_info "This is $HOSTNAME (index $CURRENT_INDEX >= replicas $DESIRED_REPLICAS). Scaling down, leaving the cluster..."
leaveOutput=$(./apinto leave)
if [[ $? -ne 0 ]]; then
    echo_error "Failed to leave the cluster: $leaveOutput"
    exit 1
else
    echo_info "Successfully left the cluster."
    # 清理标记文件，使得该序号的 Pod 若日后扩容重建时，会被当作全新节点重新 join。
    rm -f "$JOINED_FLAG"
    echo_info "Removed $JOINED_FLAG."
    ./apinto stop
    echo_info "Apinto stopped successfully."
fi
