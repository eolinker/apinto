#!/usr/bin/env bash

echo $0

. $(dirname $0)/common.sh

#echo ${BasePath}
#echo ${CMD}
#echo ${Hour}

VERSION=$(genVersion $1)
ARCH=$2
OUTPATH="${BasePath}/out/apinto-${VERSION}"
buildApp apinto $VERSION ${ARCH}

cp -a ${BasePath}/build/resources/*  ${OUTPATH}/

cd ${ORGPATH}
