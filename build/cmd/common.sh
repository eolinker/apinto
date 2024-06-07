#!/usr/bin/env bash

ORGPATH=$(pwd) #原始目录

cd  $(dirname $0) # 当前位置跳到脚本位置
CMD=$(pwd) # 脚本所在位置
cd ../..
BasePath=$(pwd) ## 项目根目录


# 生成版本号
function genVersion(){

    if [[ "$1" = "" ]]
    then
       v=$(git rev-parse --short HEAD)
       time=$(date "+%Y%m%d%H")
       echo "$time-$v"
       exit 0
    fi
    echo "$1"
}

# 构建app
function buildApp(){
    APP=$1
    VERSION=$2
    OUTPATH="${BasePath}/out/${APP}-${VERSION}"
    echo "rm -rf ${OUTPATH}"
    rm -rf ${OUTPATH}
    echo "mkdir -p ${OUTPATH}"
    mkdir -p ${OUTPATH}
    BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    EOSC_VERSION=$(sed -n 's/.*eosc v/v/p' ${BasePath}/go.mod)
    flags="-X 'github.com/eolinker/apinto/utils/version.Version=${VERSION}'
           -X 'github.com/eolinker/apinto/utils/version.gitCommit=$(git rev-parse HEAD)'
           -X 'github.com/eolinker/apinto/utils/version.buildTime=${BUILD_TIME}'
           -X 'github.com/eolinker/apinto/utils/version.buildUser=gitlab'
           -X 'github.com/eolinker/apinto/utils/version.goVersion=$(go version)'
           -X 'github.com/eolinker/apinto/utils/version.eoscVersion=${EOSC_VERSION}'"
    echo -e "build $APP:go build -ldflags "-w -s $flags" -o ${OUTPATH}/$APP ${BasePath}/app/$APP"
    ARCH=$3
    if [[ "$ARCH" = "" ]]
    then
    		ARCH="amd64"
    fi
    CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -ldflags "-w -s $flags" -o ${OUTPATH}/$APP ${BasePath}/app/$APP
#    echo "build $APP:${buildCMD}"

#    echo `${buildCMD}`

    if [[ "$?" != "0" ]]
    then
        rm -rf $OUTPATH
        exit 1
    fi
    echo "$VERSION" > ${OUTPATH}/version

}
#打包app
function packageApp(){
    APP=$1
    VERSION=$2
    ARCH=$3
    if [[ "$ARCH" = "" ]]
    then
    		ARCH="amd64"
    fi
    cp -r "${BasePath}/out/${APP}-${VERSION}" "${BasePath}/out/${APP}"
		cd "${BasePath}/out"
    tar -zcf "${BasePath}/out/${APP}_${VERSION}_linux_${ARCH}.tar.gz" ${APP}
    rm -rf "${BasePath}/out/${APP}"
    cd "${BasePath}"
}

function buildPlugin() {
    pluginName=$1
    OUTPATH="${BasePath}/out/plugins"
    CODEPATH="${BasePath}/app/plugins/$pluginName"
    mkdir -p ${OUTPATH}
    rm -f "${OUTPATH}/$pluginName.so"

    buildCMD="go build  --buildmode=plugin -o ${OUTPATH}/$pluginName.so "

    echo "build plugin $pluginName:$buildCMD ${CODEPATH}"

    orgPath="$(pwd)"
    cd ${CODEPATH}
    $buildCMD

    if [[ "$?" != "0" ]]
    then
        cd $orgPath
        exit 1
    fi

    cd $orgPath
}
