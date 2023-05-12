#!/bin/bash

# Start a K8s dashboard locally
USERDATA=$(kubectl describe serviceaccount admin-user -n kubernetes-dashboard)
TOKENNAME=$(echo "$USERDATA" | grep "Tokens" | awk '{print $2}')
TOKENDATA=$(kubectl describe secret $TOKENNAME -n kubernetes-dashboard)
TOKENVALUE=$(echo "$TOKENDATA" | grep token: | awk '{print $2}')
echo "Token for Kubenetes Dashboard: $TOKENVALUE"


kubectl proxy &

export POD_NAME=$(kubectl get pods -n kubernetes-dashboard -l "app.kubernetes.io/name=kubernetes-dashboard,app.kubernetes.io/instance=dashboard" -o jsonpath="{.items[0].metadata.name}")

echo "Access Dashboard at:  http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:dashboard-kubernetes-dashboard:https/proxy/#/login"