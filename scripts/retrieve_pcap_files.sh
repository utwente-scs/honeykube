#!/bin/bash

# Script to retrieve pcap files from the pod containers.

HOST_DIR=$1
ARCHIVE_DIR=/tmp/archive

# Get all the pods in default namespace and the conatainers in those pods
PODS_STR=$(kubectl get pods -o jsonpath='{range .items[*]}{"\n"}{.metadata.name}{":"}{range .spec.containers[*]}{.name}{","}{end}{" | "}{end}')
# Split into a list
PODS_LIST=( $(echo $PODS_STR | tr " | " " ") )

# echo ${PODS_LIST[*]}

for i in "${PODS_LIST[@]}"
do
    # Get the pod name and the list of containers
    POD_NAME=$(echo $i | cut -f1 -d:)
    CONTAINERS=$(echo $i | cut -f2 -d:)
    SERVICE_NAME=$(echo $POD_NAME | cut -f1 -d-)

    # echo $POD_NAME
    if [[ $CONTAINERS == *"tcpdump"* ]]; then
        mkdir -p $HOST_DIR/$SERVICE_NAME
        FILE_LIST=($(kubectl exec $POD_NAME -c tcpdump -it -- ls $ARCHIVE_DIR))
        len=${#FILE_LIST[@]}
        echo $len
        if [[ $len > 0 ]]; then
            kubectl cp $POD_NAME:$ARCHIVE_DIR -c tcpdump $HOST_DIR/$SERVICE_NAME
            kubectl exec $POD_NAME -c tcpdump -it -- sh /home/root/clear.sh
        fi
    fi
done