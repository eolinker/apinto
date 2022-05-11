#!/bin/bash
set -e

#启动网关
./apinto start

echo "APINTO start Success!" >> a.out
tail -f a.out