#!/bin/bash

# Script to copy files from the nodes to the local system.
# Used to copy the system trace files from the nodes.

CLUSTER=$1
SRC_PATH=$2
DEST_PATH=$3

# Get all nodes in the cluster
NODES_STR=$(kubectl get nodes --kubeconfig ${CLUSTER} -o jsonpath='{range .items[*]}{"\n"}{.metadata.name}{":"}{range .spec.containers[*]}{.name}{","}{end}{" | "}{end}')
# Split into a list
NODES_LIST=($(echo $NODES_STR | tr ": | " " "))

echo ${NODES_LIST[*]}

# Copy files from all the nodes
for node in "${NODES_LIST[@]}"
do
    gcloud compute scp username@$node:$SRC_PATH  $DEST_PATH --zone europe-west4-a
done