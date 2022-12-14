#!/usr/bin/env bash

set -e

IS_ARM=$1

IS_LATEST=$2

ORG_PATH=$(pwd)

COMMIT_ID=$(git rev-parse --short HEAD)
VERSION=$(git tag --contain ${COMMIT_ID} | awk 'END {print}')

if [ "${VERSION}" == "" ];then
	echo "[ERROR] Without tag commit."
	exit 1
fi

#cd ../../ && GOVERSION=$(go version) EoscVersion=$(sed -n 's/.*eosc v/v/p' go.mod) goreleaser release --skip-publish --rm-dist
#
#cd $(pwd)
#
#sleep 20

cp ../../dist/apinto_${VERSION}_linux_amd64.tar.gz apinto.linux.x64.tar.gz

PLATFORM=""
if [ $IS_ARM == "true" ];then
	PLATFORM="--platform=linux/amd64"
fi

docker build ${PLATFORM} -t "docker.eolinker.com/docker/apinto:${VERSION}" ./

docker push "docker.eolinker.com/docker/apinto:${VERSION}"

if [ $IS_LATEST == "true" ];then
	docker tag "docker.eolinker.com/docker/apinto:${VERSION}" "docker.eolinker.com/docker/apinto:latest"
	docker push "docker.eolinker.com/docker/apinto:latest"
fi

rm apinto.linux.x64.tar.gz