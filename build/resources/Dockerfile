# 名称：apinto镜像，携带了部署k8s集群所需要的脚本
# 创建时间：2022-3-30
FROM centos:7.9.2009
MAINTAINER eolink

#声明端口
EXPOSE 9400 8099

#定义数据卷
VOLUME /var/lib/apinto

#设置环境变量
ENV APP=apinto

#解压网关程序压缩包
COPY ./apinto.linux.x64.tar.gz /
RUN tar -zxvf apinto.linux.x64.tar.gz && rm -rf apinto.linux.x64.tar.gz

#修改脚本权限以及复制程序默认配置文件
RUN chmod 777 /apinto/start.sh && chmod 777 /apinto/join.sh && chmod 777 /apinto/leave.sh && cp /apinto/config.yml.tmp /apinto/config.yml

WORKDIR /apinto

#容器启动命令
CMD sh start.sh
