#!/bin/sh
chmod +x /data/bcs/bcs-*

/data/bcs/bcs-services/bcs-metriccollector -f /data/bcs/config_file_docker.json --log-dir=/data/bcs/logs --logtostderr=true