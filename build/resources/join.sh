#!/bin/bash
set -e

sleep 5s

#获取目标服务下所有pod的ip
response=$(curl -s GET "https://kubernetes.default.svc:443/api/v1/namespaces/${SVC_NAMESPACE}/endpoints/${SVC_NAME}" -k -H "Authorization: ${SVC_TOKEN}")

#判断请求是否成功
if [[ ${response} =~ 'Failure' ]]
then
    echo ${response}
    exit 1
fi

set +e
#判断返回结果内是否有pod id
ips=$( echo ${response} | jq -r '.subsets[].addresses[].ip' )

#若ips为空
if [[ "$( echo ${ips} )" == '' ]]
then
    echo "There are no pods ip in Service "
    exit 1
fi

set -e

#遍历ip节点
for ip in ${ips}
do
  if [ ${ip} != ${POD_IP} ]
  then
  #加入集群
    ./apinto join --ip ${POD_IP} --addr=${ip}:${APINTO_ADMIN_PORT}
    break
  fi
done