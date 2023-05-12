#!/bin/bash

# Copy a directory from the node to the local system.
# Used to get system log files from the nodes

CLUSTER=$1
SRC_PATH=$2
DEST_PATH=$3

# Get all nodes in the cluster
NODES_STR=$(kubectl get nodes --kubeconfig ${CLUSTER} -o jsonpath='{range .items[*]}{"\n"}{.metadata.name}{":"}{range .spec.containers[*]}{.name}{","}{end}{" | "}{end}')
# Split into a list
NODES_LIST=($(echo $NODES_STR | tr ": | " " "))

echo ${NODES_LIST[*]}

for node in "${NODES_LIST[@]}"
do
    mkdir -p $DEST_PATH/$node
    gcloud compute scp --recurse padfoot@$node:$SRC_PATH  $DEST_PATH/$node --zone europe-west4-a
done