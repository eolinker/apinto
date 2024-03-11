#!/bin/bash

# 判断进程是否正在运行
isProcessRunning() {
    pid=`ps ax | grep "$PROG_PATH/$PROG" | grep -v "grep" | awk '{print $1}'`
    if [[ $pid != "" ]] ; then
        echo "true"
        return "$?"
    else
        echo "false"
        return "$?"
    fi
}

set -e

if [[ $APINTO_DEBUG == "true" ]]; then
		#Launch the gateway
    ./apinto start
    echo "[$(date "+%Y-%m-%d %H:%M:%S")] Gateway Stop"
    exit 0
fi

is_process_running=$(isProcessRunning)
if [[ "$is_process_running" = "true" ]] ; then
	echo "[$(date "+%Y-%m-%d %H:%M:%S")] process still running, waiting..."
	sleep 5s
fi

./apinto start
is_process_running=$(isProcessRunning)
if [[ "$is_process_running" = "true" ]] ; then
	echo "[$(date "+%Y-%m-%d %H:%M:%S")] APINTO start Success!" >> a.out
  tail -f a.out
  echo "[$(date "+%Y-%m-%d %H:%M:%S")] Gateway Stop"
  exit 0
else
	echo "[$(date "+%Y-%m-%d %H:%M:%S")] APINTO start Failed!" >> a.out
fi


