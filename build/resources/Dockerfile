# 名称：apinto自定义插件环境
# 用途：在该环境中编译goku-ce自定义插件，防止出现版本不兼容的问题
# 创建时间：2021-11-25
FROM centos:7

# 安装wget
RUN yum install -y wget

# 获取golang安装包
RUN wget https://dl.google.com/go/go1.17.3.linux-amd64.tar.gz
RUN tar -xvf go1.17.3.linux-amd64.tar.gz

RUN mkdir -p /golang/src/github.com/eolinker/apinto

# 配置环境变量
ENV GOPATH /golang
ENV GOROOT /go
ENV GO111MODULE on
ENV GOPROXY https://goproxy.io
ENV PATH $GOPATH/bin:$GOROOT/bin:$PATH

# 安装gcc编译程序
RUN yum install -y gcc

# 安装git
RUN yum install -y git
