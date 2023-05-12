#!/bin/bash

# Delete the services using skaffold
skaffold delete

# Delete the kubernetes components from the cluster
kubectl delete all --all
kubectl delete pvc --all
kubectl delete pv --all
kubectl delete secrets --all
kubectl delete configmaps --all

# Delete the cluster
kind delete cluster