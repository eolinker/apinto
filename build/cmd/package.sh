#!/usr/bin/env bash

. $(dirname $0)/common.sh

cd ${BasePath}/


VERSION=$(genVersion $1)
folder="${BasePath}/out/apinto-${VERSION}"
if [[ ! -d "$folder" ]]
then
#  mkdir -p "$folder"
  ${CMD}/build.sh $1
  if [[ "$?" != "0" ]]
  then
    exit 1
  fi
fi
packageApp apinto $VERSION

cd ${ORGPATH}
