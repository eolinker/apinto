#!/usr/bin/env bash

echo $0
. $(dirname $0)/common.sh

VERSION=`git describe --abbrev=0 --tags`
Username="eolinker"
if [[ "$1" != "" ]]
then
		Username=$1
fi

PackageName=apinto_${VERSION}_linux_amd64.tar.gz
cp dist/${PackageName} ./build/resources/apinto.linux.x64.tar.gz

docker build -t ${Username}/apinto-gateway:${VERSION} -f ./build/resources/Dockerfile ./build/resources

rm -rf ./build/resources/apinto.linux.x64.tar.gz

cd ${ORGPATH}
