#!/bin/bash

ERR_LOG=/var/log/apinto/error.log
echo_info() {
    echo "[$(date "+%Y-%m-%d %H:%M:%S")] [INFO] $1" >> $ERR_LOG
}

echo_error() {
    echo "[$(date "+%Y-%m-%d %H:%M:%S")] [ERROR] $1" >> $ERR_LOG
}
#This script is used to leave the K8S cluster
leaveOutput=$(./apinto leave)
if [[ $? -ne 0 ]]; then
	echo_error "Failed to leave the cluster: $leaveOutput"
	exit 1
else
	echo_info "Successfully left the cluster."
	./apinto stop
	echo_info "Apinto stopped successfully."
fi

