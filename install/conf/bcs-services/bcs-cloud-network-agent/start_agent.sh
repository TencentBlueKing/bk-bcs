#!/bin/bash

# copy cni files
cp /bcs/bcs-eni /bcs/cni/bin/
cp /bcs/bcs-eni.conf /bcs/cni/conf/

# start agent
/bcs/bcs-cloud-network-agent $@