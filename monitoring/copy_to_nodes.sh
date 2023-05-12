#!/bin/bash

# Script to copy files to all the nodes from the local system
# Used to copy the script to start monitoring the system's kernel for executed system calls.

CLUSTER=$1
SRC_PATH=$2
DEST_PATH=$3

# Get all nodes in the cluster
NODES_STR=$(kubectl get nodes --kubeconfig ${CLUSTER} -o jsonpath='{range .items[*]}{"\n"}{.metadata.name}{":"}{range .spec.containers[*]}{.name}{","}{end}{" | "}{end}')
# Split into a list
NODES_LIST=($(echo $NODES_STR | tr ": | " " "))

echo ${NODES_LIST[*]}

# Copy files to all the nodes
for node in "${NODES_LIST[@]}"
do
    gcloud compute scp $SRC_PATH username@$node:$DEST_PATH --zone europe-west4-a
done