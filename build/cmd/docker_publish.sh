#!/usr/bin/env bash


echo $0
. $(dirname $0)/common.sh


Version=$(genVersion)
echo ${Version}
ImageName="docker.eolinker.com/apinto/apinto-gateway"
echo "docker manifest rm  \"${ImageName}:${Version}\""
docker manifest rm "${ImageName}:${Version}"

set -e
./build/cmd/docker_build.sh ${Version} "docker.eolinker.com/apinto" amd64

./build/cmd/docker_build.sh ${Version} "docker.eolinker.com/apinto" arm64



echo "docker push \"${ImageName}:${Version}-amd64\""
docker push "${ImageName}:${Version}-amd64"
echo "docker push \"${ImageName}:${Version}-arm64\""
docker push "${ImageName}:${Version}-arm64"

echo "Create manifest ${ImageName}:${Version}"
docker manifest create "${ImageName}:${Version}" "${ImageName}:${Version}-amd64" "${ImageName}:${Version}-arm64"

echo "Annotate manifest ${ImageName}:${Version} ${ImageName}:${Version}-amd64 --os linux --arch amd64"
docker manifest annotate "${ImageName}:${Version}" "${ImageName}:${Version}-amd64" --os linux --arch amd64

echo "Annotate manifest ${ImageName}:${Version} ${ImageName}:${Version}-arm64 --os linux --arch arm64"
docker manifest annotate "${ImageName}:${Version}" "${ImageName}:${Version}-arm64" --os linux --arch arm64

echo "Push manifest ${ImageName}:${Version}"
docker manifest push "${ImageName}:${Version}"


PUBLISH=$1
if [[ "${PUBLISH}" == "upload" ]];then
  ./build/cmd/qiniu_publish.sh amd64
  ./build/cmd/qiniu_publish.sh arm64
fi

