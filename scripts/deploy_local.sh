#!/bin/bash

# Run the script with "sudo" permissions
if [[ "$EUID" != 0 ]]; then
    echo "Run the script with \"sudo\"."
    exit
fi

# Give home directory to be used in argument
FALCO_DIR=../charts/falco

# start kind cluster with 4 nodes
kind create cluster --config kind-config.yaml


# Clone the falco charts repository and switch to the "honeykube" branch.
# In this branch, we specified the rules for the honeykube honeypot.
if [[ ! -d $FALCO_DIR ]]
then
    git clone https://github.com/ChakshuGupta/charts.git
fi

cd $FALCO_DIR
git switch honeykube

# Create the directories for the volumes
./scripts/create_volume_directories.sh

# Create the secrets, volumes, cofigmaps and databases in the cluster
kubectl apply -f ./kubernetes-manifests/secrets/
kubectl apply -f ./kubernetes-manifests/volumes/
kubectl apply -f ./kubernetes-manifests/configmaps/
kubectl apply -f ./kubernetes-manifests/databases/

# Create the service account "test"
kubectl create sa test

# Deploy the services on the cluster using skaffold
skaffold run