#!/usr/bin/env bash


echo $0
. $(dirname $0)/common.sh


Version=$(genVersion $1)
echo ${Version}

Username="eolinker"
if [[ "$1" != "" ]]
then
		Username=$2
fi
ImageName="${Username}/apinto-gateway"

./build/cmd/docker_build.sh ${Version} ${Username} amd64

./build/cmd/docker_build.sh ${Version} ${Username} arm64

publish_hub() {
	Version=$1
	echo "docker manifest rm  \"${ImageName}:${Version}\""
  docker manifest rm "${ImageName}:${Version}"
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
}

publish_hub ${Version}

PUBLISH=$3
if [[ "${PUBLISH}" == "upload" ]];then
  ./build/cmd/qiniu_publish.sh amd64
  ./build/cmd/qiniu_publish.sh arm64
fi

if [[ $4 == "latest" ]];then
	echo "Publish latest version"
	docker tag "${ImageName}:${Version}-amd64" "${ImageName}:latest-amd64"
	docker tag "${ImageName}:${Version}-arm64" "${ImageName}:latest-arm64"
	publish_hub "latest"
fi
