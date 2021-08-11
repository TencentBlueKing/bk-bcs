#!/bin/bash 

cd /data/bcs/bcs-hook-daemonset
chmod +x bcs-hook-daemonset
#start daemonset
./bcs-hook-daemonset --v=5 --leader-elect=false