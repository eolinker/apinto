#!/bin/bash
set -xe

#This script is used to join the K8S cluster
sleep 10s

#Gets the IP addresses of all pods under the target service
response=$(curl -s "https://kubernetes.default.svc:443/api/v1/namespaces/${SVC_NAMESPACE}/endpoints/${SVC_NAME}" -k -H "Authorization: Bearer ${SVC_TOKEN}")

#Determines whether the request was successful
if [[ ${response} =~ 'Failure' ]]
then
    echo ${response}
    exit 1
fi

set +e
#Check whether there is a POD ID in the result
{
        hostnames=$( echo ${response} | jq -r '.subsets[].addresses[].hostname' )
} || {
        exit 0
}

podHostName=$(hostname)
#If the ips is null
if [[ "$( echo ${hostnames} )" == '' ]]
then
    echo "There are no pods ip in Service "
    exit 1
fi

set -e

#Traverses all the Node's IP
for hn in ${hostnames}
do
  if [ ${hn} != ${podHostName} ]
      then
      #join the cluster
      return_info=$(./apinto join --addr=${hn}.${SVC_NAME}:${APINTO_ADMIN_PORT})
      if [[ $return_info = '' ]]
      then
        break
      fi
  fi
done
