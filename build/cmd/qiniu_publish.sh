#!/bin/sh

set -e

. $(dirname $0)/common.sh


Version=$(genVersion)

ImageName="docker.eolinker.com/apinto/apinto-gateway"
APP="apinto-gateway"

ARCH=$1
if [[ $ARCH == "" ]];then
  ARCH="amd64"
fi

Tar="${APP}.${Version}.${ARCH}.tar.gz"

docker tag ${ImageName}:${Version}-${ARCH} ${ImageName}:${Version}

echo "docker save -o ${Tar} ${ImageName}:${Version}"
docker save -o ${Tar} ${ImageName}:${Version}

echo "qshell rput eolinker-main \"apinto/images/${Tar}\" ${Tar}"
qshell rput eolinker-main "apinto/images/${Tar}" ${Tar}

rm -f ${Tar}
docker rmi -f ${ImageName}:${Version}