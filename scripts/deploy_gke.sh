#!/bin/bash

HOME=$1

FALCO_DIR=$HOME/charts/falco
HONEYKUBE_DIR=$HOME/honeykube

cd $HOME
pwd

if [[ ! -d $FALCO_DIR ]]
then
    git clone https://github.com/ChakshuGupta/charts.git
fi

cd $HONEYKUBE_DIR
pwd

./scripts/create_volume_directories.sh

# Create the service account "test"
kubectl create sa test

# Create the secrets, volumes, cofigmaps and databases in the cluster
kubectl apply -f ./kubernetes-manifests/secrets/
kubectl apply -f ./kubernetes-manifests/volumes/
kubectl apply -f ./kubernetes-manifests/configmaps/
kubectl apply -f ./kubernetes-manifests/databases/

# skaffold run --default-repo='eu.gcr.io/research-gcp-credits'
kubectl apply -f ./release/

cd $FALCO_DIR
git checkout honeykube

helm dependency update
helm install falco . --namespace falco --create-namespace