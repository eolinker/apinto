#!/usr/bin/env bash

. $(dirname $0)/common.sh

cd ${BasePath}/


VERSION=$(genVersion $1)
ARCH=$2
if [[ "$ARCH" == "" ]]
then
	ARCH="amd64"
fi
folder="${BasePath}/out/apinto-${VERSION}-${ARCH}"

if [[ ! -d "$folder" ]]
then
  mkdir -p "$folder"
  ${CMD}/build.sh $1 ${ARCH}
  if [[ "$?" != "0" ]]
  then
    exit 1
  fi
fi
packageApp apinto $VERSION ${ARCH}

cd ${ORGPATH}
