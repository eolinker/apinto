#!/usr/bin/env bash
#
# 部署脚本：打包 apinto，上传到远程 Linux 服务器并通过软链方式更新二进制。
#
# 前提：本机公钥已同步到远程服务器，可直接免密 ssh。
#
# 参数通过环境变量传入，脚本会自动加载项目根目录的 .env 文件（如存在）。
#
#   必填：
#     REMOTE          远程服务器地址，形如 user@host
#
#   可选（含默认值）：
#     VERSION         版本号（不填则自动获取：优先 git 最新 tag，无 tag 则回退到 commit 短哈希）
#     ARCH            架构，默认 amd64
#     REMOTE_DIR      远程部署根目录，默认 /opt/apinto
#     SSH_PORT        SSH 端口，默认 22
#     SKIP_BUILD      设为 1 时跳过打包（复用已有产物）
#     RESTART_CMD     部署完成后在远程执行的重启命令
#     ENV_FILE        env 文件路径，默认 ${项目根}/.env
#
# 示例：
#   ./deploy.sh                                             # 全部走 .env
#   REMOTE=root@10.0.0.10 ./deploy.sh
#   REMOTE=root@10.0.0.10 VERSION=v0.20.0 REMOTE_DIR=/data/apinto ./deploy.sh
#   REMOTE=root@10.0.0.10 RESTART_CMD='systemctl restart apinto' ./deploy.sh
#

set -e

. $(dirname $0)/common.sh

# ---------- 自动加载 .env ----------
# 优先使用 ENV_FILE 指定的文件；否则读取项目根目录下的 .env
: "${ENV_FILE:=${BasePath}/.env}"
if [[ -f "$ENV_FILE" ]]; then
    echo ">>> 加载环境变量文件: ${ENV_FILE}"
    set -a
    # shellcheck disable=SC1090
    . "$ENV_FILE"
    set +a
fi

# ---------- 环境变量默认值 ----------
: "${REMOTE:=}"
: "${VERSION:=}"
: "${ARCH:=amd64}"
: "${REMOTE_DIR:=/opt/apinto}"
: "${SSH_PORT:=22}"
: "${SKIP_BUILD:=0}"
: "${RESTART_CMD:=}"

if [[ -z "$REMOTE" ]]; then
    echo "错误: 必须通过环境变量 REMOTE 指定远程服务器 (例: REMOTE=user@host)"
    echo "     可通过命令行传入，或写在 ${ENV_FILE} 中"
    sed -n '2,26p' "$0"
    exit 1
fi

# ---------- ssh / scp 命令 ----------
SSH_OPTS="-p ${SSH_PORT} -o StrictHostKeyChecking=no"
SCP_OPTS="-P ${SSH_PORT} -o StrictHostKeyChecking=no"

# ---------- 打包 ----------
# VERSION 未指定时，genVersion 会自动使用 git 最新 tag，若无 tag 则回退到 commit 短哈希
VERSION=$(genVersion "$VERSION")
echo ">>> 版本号: ${VERSION}"
PKG_NAME="apinto_${VERSION}_linux_${ARCH}.tar.gz"
PKG_PATH="${BasePath}/out/${PKG_NAME}"
BUILD_DIR="${BasePath}/out/apinto-${VERSION}-${ARCH}"

if [[ "$SKIP_BUILD" != "1" ]]; then
    # package.sh 里判断中间目录不存在才会重新构建；这里先清掉旧目录和旧包，
    # 避免上一次构建失败留下的空目录被直接打包，导致 tar 里没有 apinto 二进制。
    echo ">>> 清理旧产物: ${BUILD_DIR} 和 ${PKG_PATH}"
    rm -rf "${BUILD_DIR}"
    rm -f "${PKG_PATH}"

    echo ">>> 开始打包: version=${VERSION}, arch=${ARCH}"
    ${CMD}/package.sh "${VERSION}" "${ARCH}"
    if [[ "$?" != "0" ]]; then
        echo "打包失败"
        exit 1
    fi
fi

if [[ ! -f "$PKG_PATH" ]]; then
    echo "错误: 打包产物不存在: $PKG_PATH"
    exit 1
fi

# 校验 tar 包内确实包含 apinto 二进制，避免打了个空目录上去
if ! tar -tzf "$PKG_PATH" | grep -q "^apinto/apinto$"; then
    echo "错误: 打包产物 ${PKG_PATH} 中未找到 apinto/apinto 二进制"
    echo "     请检查 go build 是否成功，或删除 out/ 目录后重试"
    exit 1
fi

echo ">>> 打包产物: ${PKG_PATH}"

# ---------- 上传 ----------
REMOTE_TMP="/tmp/${PKG_NAME}"
echo ">>> 上传到 ${REMOTE}:${REMOTE_TMP}"
scp ${SCP_OPTS} "${PKG_PATH}" "${REMOTE}:${REMOTE_TMP}"

# ---------- 远程部署 ----------
#
# 远程目录结构：
#   ${REMOTE_DIR}/
#     ├── apinto            -> releases/${VERSION}/apinto   (软链，指向当前版本二进制)
#     ├── releases/
#     │     ├── ${VERSION}/apinto
#     │     └── ...历史版本
#     ├── config/           (配置目录，仅首次从包中拷贝，后续不覆盖)
#     └── ...其它资源（同 config，仅首次拷贝）
#
# 每次部署只替换二进制 apinto，配置文件保持不变。
#
echo ">>> 在远程执行部署"
# 通过 bash -s -- 位置参数把变量安全地传给远端脚本，避免含空格时被拆词
ssh ${SSH_OPTS} "${REMOTE}" \
    "bash -s -- \"${REMOTE_DIR}\" \"${VERSION}\" \"${REMOTE_TMP}\" \"${RESTART_CMD}\"" <<'REMOTE_SCRIPT'
set -e

REMOTE_DIR="$1"
VERSION="$2"
PKG="$3"
RESTART_CMD="$4"

RELEASE_DIR="${REMOTE_DIR}/releases/${VERSION}"
CURRENT_LINK="${REMOTE_DIR}/apinto"

echo ">>> 准备目录 ${REMOTE_DIR}"
mkdir -p "${REMOTE_DIR}/releases"

# 解压到临时目录（包内已经是 apinto/ 顶层目录）
TMP_EXTRACT=$(mktemp -d)
echo ">>> 解压 ${PKG} 到 ${TMP_EXTRACT}"
tar -zxf "${PKG}" -C "${TMP_EXTRACT}"

SRC="${TMP_EXTRACT}/apinto"
if [[ ! -d "${SRC}" ]]; then
    echo "错误: 解压后未找到 apinto 目录"
    rm -rf "${TMP_EXTRACT}"
    exit 1
fi
if [[ ! -f "${SRC}/apinto" ]]; then
    echo "错误: 解压后未找到 apinto 二进制 (${SRC}/apinto)"
    rm -rf "${TMP_EXTRACT}"
    exit 1
fi

# 1. 存放本次版本的二进制
echo ">>> 部署二进制到 ${RELEASE_DIR}"
mkdir -p "${RELEASE_DIR}"
cp -f "${SRC}/apinto" "${RELEASE_DIR}/apinto"
chmod +x "${RELEASE_DIR}/apinto"

# 2. 首次部署时，把配置和其它 resources 拷贝到部署目录；已存在则保留不覆盖
for entry in "${SRC}"/*; do
    name=$(basename "${entry}")
    # apinto 二进制单独通过软链管理，跳过
    if [[ "${name}" == "apinto" ]]; then
        continue
    fi
    target="${REMOTE_DIR}/${name}"
    if [[ -e "${target}" ]]; then
        echo "    保留已存在: ${target}"
    else
        echo "    初次拷贝:   ${target}"
        cp -a "${entry}" "${target}"
    fi
done

# 3. 更新软链: ${REMOTE_DIR}/apinto -> releases/${VERSION}/apinto
echo ">>> 更新软链 ${CURRENT_LINK} -> ${RELEASE_DIR}/apinto"
ln -sfn "${RELEASE_DIR}/apinto" "${CURRENT_LINK}"

# 4. 清理临时文件
rm -rf "${TMP_EXTRACT}"
rm -f "${PKG}"

# 5. 可选重启
if [[ -n "${RESTART_CMD}" ]]; then
    echo ">>> 执行重启命令: ${RESTART_CMD}"
    bash -c "${RESTART_CMD}"
fi

echo ">>> 部署完成: 版本 ${VERSION}"
REMOTE_SCRIPT

echo ">>> 全部完成: ${REMOTE}:${REMOTE_DIR} 已更新到 ${VERSION}"

cd ${ORGPATH}
