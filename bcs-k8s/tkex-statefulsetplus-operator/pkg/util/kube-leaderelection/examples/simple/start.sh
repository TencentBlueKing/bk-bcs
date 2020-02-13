#!/bin/bash

echo "-----------starting kube-leaderelection--------------"

cd  `dirname $0`
./server --electionConfigPath="./config.json"