#!/bin/bash
set -e

mkdir /etc/apinto

#将模板配置文件复制到/etc/apinto目录, 若已存在则不覆盖
cp -in ./apinto.yml.tpl /etc/apinto/apinto.yml
cp -in ./config.yml.tpl /etc/apinto/config.yml

#将程序链接至/usr/sbin
ln -snf ./apinto /usr/sbin/apinto