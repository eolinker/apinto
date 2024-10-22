#!/bin/bash

# 判断进程是否正在运行
isProcessRunning() {
    pid=`ps ax | grep apinto | grep -v "grep" | wc -l`
    if [[ $pid != "3" ]] ; then
        echo "false"
        return "$?"
    else
        echo "true"
        return "$?"
    fi
}

set -e

if [[ $APINTO_DEBUG == "true" ]]; then
                #Launch the gateway
    ./apinto debug master
    echo "[$(date "+%Y-%m-%d %H:%M:%S")] Gateway Stop"
    exit 0
fi

is_process_running=$(isProcessRunning)
if [[ "$is_process_running" = "true" ]] ; then
        echo "[$(date "+%Y-%m-%d %H:%M:%S")] process still running, waiting..."
        sleep 5s
fi

./apinto start
sleep 5s
is_process_running=$(isProcessRunning)
if [[ "$is_process_running" = "true" ]] ; then
        echo "[$(date "+%Y-%m-%d %H:%M:%S")] APINTO start Success!" >> a.out
  tail -f a.out
  echo "[$(date "+%Y-%m-%d %H:%M:%S")] Gateway Stop"
  exit 0
else
        echo "[$(date "+%Y-%m-%d %H:%M:%S")] APINTO start Failed!" >> a.out
fi
