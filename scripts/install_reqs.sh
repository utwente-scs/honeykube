#!/bin/bash

### Install Docker

# Remove docker packages if they exist in the system
sudo apt remove docker docker-engine docker.io containerd runc

sudo apt update
sudo apt install -y \
   apt-transport-https \
   ca-certificates \
   curl \
   gnupg

curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

### Replace "$(lsb_release -cs)" with "bionic" for Linux Mint
echo \
  "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null


### Update the system and install docker and containerd
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io

### Install Kind
 sudo apt install -y curl
 curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.10.0/kind-linux-amd64
 chmod +x kind
 sudo mv kind /usr/bin


### Install Kubernetes

sudo apt update && sudo apt install -y apt-transport-https gnupg2 curl
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -

echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee -a /etc/apt/sources.list.d/kubernetes.list

sudo apt update
sudo apt install -y kubectl


### Install Helm

curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3
chmod 700 get_helm.sh
./get_helm.sh


### Install Sysdig
curl -s https://s3.amazonaws.com/download.draios.com/stable/install-sysdig | sudo bash


### Install Skaffold
curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64 && sudo install skaffold /usr/local/bin/
