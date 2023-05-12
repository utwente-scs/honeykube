#!/bin/bash

# Run this script on the sidecar container in the pods.
SERVICE=$1
TIMEOUT=$2
ARCHIVE_DIR=/tmp/archive/

# Make directory
mkdir -p $ARCHIVE_DIR

# Start the infinite loop to record the network traffic into pcap files.
while :
do

    CURRENTDATE=`date +"%Y-%m-%d_%H-%M-%S"`
    FILENAME=/tmp/${SERVICE}-${CURRENTDATE}.pcap
    
    timeout ${TIMEOUT} tcpdump -s 0 -n -w $FILENAME

    mv $FILENAME $ARCHIVE_DIR

done