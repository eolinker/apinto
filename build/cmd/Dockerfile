# 名称：apinto镜像，携带了部署k8s集群所需要的脚本
# 创建时间：2022-3-30
FROM alpine:latest

RUN sed -i 's|https://dl-cdn.alpinelinux.org/alpine|https://mirrors.aliyun.com/alpine|g' /etc/apk/repositories  \
    && apk update \
    && apk add --no-cache bash tzdata

RUN apk add --no-cache netcat-openbsd curl

ENV TZ=Asia/Shanghai

RUN cp /usr/share/zoneinfo/${TZ} /etc/localtime && echo ${TZ} > /etc/timezone

#设置环境变量
ARG AppName=apinto

#定义数据卷
VOLUME /var/lib/${AppName}

#解压网关程序压缩包
COPY ./apinto.linux.x64.tar.gz /
RUN tar -zxvf apinto.linux.x64.tar.gz && ls /${AppName} / && rm -rf ../apinto.linux.x64.tar.gz

#复制程序默认配置文件以及修改脚本权限
RUN mkdir /etc/${AppName}
RUN cp /${AppName}/${AppName}.yml.tpl /etc/${AppName}/${AppName}.yml && cp /${AppName}/config.yml.tpl /etc/${AppName}/config.yml
RUN chmod 777 /${AppName}/*

WORKDIR /${AppName}

#容器启动命令
CMD bash "auto-start.sh"
