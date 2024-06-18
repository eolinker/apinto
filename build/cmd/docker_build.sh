#!/usr/bin/env bash

echo $0
. $(dirname $0)/common.sh

#VERSION=`git describe --abbrev=0 --tags`
VERSION=$(genVersion $1)

Username="eolinker"
if [[ "$2" != "" ]]
then
		Username=$2
fi

ARCH=$3
PLATFORM=""
if [[ "$ARCH" == "" ]]
then
		ARCH="amd64"
fi
if [[ "$ARCH" == "amd64" ]]
then
	PLATFORM="--platform=linux/amd64"
fi
./build/cmd/package.sh ${VERSION} ${ARCH}
PackageName=apinto_${VERSION}_linux_${ARCH}.tar.gz
cp out/${PackageName} ./build/resources/apinto.linux.x64.tar.gz

docker build $PLATFORM -t ${Username}/apinto-gateway:${VERSION}-${ARCH} -f ./build/cmd/Dockerfile ./build/resources

rm -rf ./build/resources/apinto.linux.x64.tar.gz

cd ${ORGPATH}
