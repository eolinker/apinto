# 名称：apinto镜像，携带了部署k8s集群所需要的脚本
# 创建时间：2022-3-30
FROM centos:7.9.2009
MAINTAINER eolink

#声明端口
EXPOSE 9400 8099

#设置环境变量
ARG AppName=apinto

#定义数据卷
VOLUME /var/lib/${AppName}

RUN ln -snf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo "Asia/Shanghai" > /etc/timezone

#解压网关程序压缩包
COPY ./apinto.linux.x64.tar.gz /
RUN tar -zxvf apinto.linux.x64.tar.gz && rm -rf apinto.linux.x64.tar.gz

#复制程序默认配置文件以及修改脚本权限
RUN mkdir /etc/${AppName}
RUN cp /${AppName}/${AppName}.yml.tpl /etc/${AppName}/${AppName}.yml && cp /${AppName}/config.yml.tpl /etc/${AppName}/config.yml
RUN chmod 777 /${AppName}/start.sh && chmod 777 /${AppName}/join.sh && chmod 777 /${AppName}/leave.sh
ADD check_remove.sh /apinto/check_remove.sh
RUN chmod 777 /apinto/check_remove.sh

ADD jq /usr/local/bin/jq

WORKDIR /${AppName}

#容器启动命令
CMD sh start.sh
