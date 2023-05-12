#!/bin/bash

# Main directory that would contain the directories for the services.
MAIN_DIR=/tmp

# List of services
SERVICES=("adservice"  "cartservice"  "checkoutservice"  "currencyservice"  "emailservice"  "frontend"  "paymentservice"  "productcatalogservice"  "productdb"  "recommendationservice"  "redis"  "shippingservice"  "userdb"  "userdbservice")

# Directories for userdb and productdb
mkdir -p $MAIN_DIR/data/users
mkdir -p $MAIN_DIR/data/products

# Traverse through the list of services to create the directories to store the pcap files. 
for i in "${SERVICES[@]}"
do
    mkdir -p $MAIN_DIR/files/$i
done