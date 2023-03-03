#!/bin/bash
set -e

CURRENT_PATH="$(pwd)"

install() {
	mkdir -p /etc/apinto

	#将模板配置文件复制到/etc/apinto目录, 若已存在则不覆盖
	cp -in ./apinto.yml.tpl /etc/apinto/apinto.yml
	cp -in ./config.yml.tpl /etc/apinto/config.yml

	#将程序链接至/usr/sbin
	ln -snf $CURRENT_PATH/apinto /usr/sbin/apinto
}

upgrade() {
	apinto stop
	install
	sleep 10s
	apinto start
}

case "$1" in
    install)
        install
        exit 0
    ;;
    upgrade)
        upgrade
        exit 0
    ;;
    **)
        echo "Usage: $0 {install|upgrade} " 1>&2
        exit 1
    ;;
esac