#!/usr/bin/env bash

echo $0

. $(dirname $0)/common.sh

#echo ${BasePath}
#echo ${CMD}
#echo ${Hour}

VERSION=$(genVersion $1)
ARCH=$2
if [[ "$ARCH" == "" ]]
then
		ARCH="amd64"
fi

OUTPATH="${BasePath}/out/apinto-${VERSION}-${ARCH}"
buildApp apinto $VERSION ${ARCH}

cp -a ${BasePath}/build/resources/*  ${OUTPATH}/

cd ${ORGPATH}
