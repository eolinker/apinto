#!/bin/bash
apinto_info=$(./apinto info)

nodes_count=$(echo $apinto_info | grep -o 'Node' | wc -l)
cluster_node=$(( $nodes_count + 1 ))

for ((i=0;i<$cluster_node;i++))
do
        count_addr=$(( $i*8+6 ))
        count_node=$(( $i*8+4 ))
        info_addr=$(echo $apinto_info | awk '{print $'"$count_addr"'}')
        info_node=$(echo $apinto_info | awk '{print $'"$count_node"'}')
        {
                curl --max-time 5 --silent --fail $info_addr
        } || {
                ./apinto remove $info_node
        }
        echo ""

done
