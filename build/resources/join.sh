#!/bin/bash
set -e
#This script is used to join the K8S cluster
sleep 5s

#Gets the IP addresses of all pods under the target service
response=$(curl -s GET "https://kubernetes.default.svc:443/api/v1/namespaces/${SVC_NAMESPACE}/endpoints/${SVC_NAME}" -k -H "Authorization: ${SVC_TOKEN}")

#Determines whether the request was successful
if [[ ${response} =~ 'Failure' ]]
then
    echo ${response}
    exit 1
fi

set +e
#Check whether there is a POD ID in the result
ips=$( echo ${response} | jq -r '.subsets[].addresses[].ip' )

#If the ips is null
if [[ "$( echo ${ips} )" == '' ]]
then
    echo "There are no pods ip in Service "
    exit 1
fi

set -e

#Traverses all the Node's IP
for ip in ${ips}
do
  if [ ${ip} != ${POD_IP} ]
  then
  #join the cluster
    ./apinto join --ip ${POD_IP} --addr=${ip}:${APINTO_ADMIN_PORT}
    break
  fi
done