#!/bin/bash

# SSH into the nodes in the GKE cluster by naming the cluster and the node number.

NODE_NUM=$1-1
CLUSTER=$2

# Get all nodes in the cluster
NODES_STR=$(kubectl get nodes --kubeconfig ${CLUSTER} -o jsonpath='{range .items[*]}{"\n"}{.metadata.name}{":"}{range .spec.containers[*]}{.name}{","}{end}{" | "}{end}')
# Split into a list
NODES_LIST=($(echo $NODES_STR | tr ": | " " "))

echo ${NODES_LIST[*]}

# SSH into the GKE cluster node.
gcloud compute ssh ${NODES_LIST[$NODE_NUM]} --zone europe-west4-a