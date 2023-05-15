# HoneyKube: Designing and Deploying a microservices-based Web Honeypot


## Directory Structure
```
.
├── db_scripts : stores sql scripts to create databases
├── docs : stores architecture diagram of the cluster
├── kubernetes-manifests : stores all the yaml files that define different components of the cluster
├── monitoring : stores scripts and codes required for the monitoring framework for the honeypot
├── release : stores the release version of the yaml files for the microservices. This version has the links to the pre-compiled images of the microservices to make deployment faster.
├── scripts : stores the different scripts used in deploying, installation and other automated actions.
└── src : stores the code for all the individual microsevices.
```


### **db_script** directory

1. **create_productcatalog.sql** : Script to create the database to store the product catalogs and insert in all the products in the database.
2. **create_user.sql** : Script to create the users for the databases.
3. **create_userdb.sql** : Script to create the database to store user information entered by the users when registering for the platform.

### **scripts** directory

1. **create_volume_directories.sh** : Creates the directories on the system to match to the persistent volumes for the databases and the pcap files.
2. **delete_cluster.sh** : Delete the deployed cluster and all of its components.
3. **deploy_gke.sh** : Script to deploy the cluster on the GKE platform.
4. **deploy_local.sh** : Script to deploy the cluster locally using *kind*.
5. **install_reqs.sh** : Script to install all the required dependencies in the system.
6. **retrieve_pcap_files.sh** : Script to retrieve pcap files from the pod containers to the host system.
7. **ssh_node.sh** : Script to SSH into the nodes in the GKE cluster by naming the cluster and the node number.
8. **start_k8_dashboard.sh** : Script to start a k8s dasboard on a locally deployed cluster.
9. **tcpdump_script.sh** : Script that is run on the sidecar container for all microservices to record the incoming and outgoing network traffic from the pods.


## How to Run the Honeypot Locally

1. Run the script to install all the dependencies.
```
./install_reqs.sh
```

2. Run the script to deploy the honeypot locally. Run it from the home directory of the repository.
```
sudo ./scripts/deploy_local.sh
```

##  How to Run the Honeypot on GKE
1. Run the script to deploy on GKE and give the path to the home directory of the repository.
```
sudo ./scripts/deploy_gke.sh {HOME_DIR}
```


Send an email to c.gupta@utwente.nl to get access to the data collected by our experiment.