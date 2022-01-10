#!/bin/bash 

cd /data/bcs/bcs-hook-operator
chmod +x bcs-hook-operator
#start operator
exec ./bcs-hook-operator --v=5

