#!/bin/bash

#script use for starting necessary module for
# for test, demonstration

echo "begin to start modules $@"
# test modules
for module in $@
do
  echo "checking module $module binary"
  if [ ! -e /data/bcs/$module/$module ]; then 
    echo "lost module $module in /data/bcs"
    exit 1
  fi
  echo "checking module $module binary successfully"
done

#ready to start all specified module by 
# using container-start script
for module in $@
do
  cd /data/bcs/$module/
  echo "starting module $module ... "
  ./container-start.sh -f $module.json
done

echo "waiting for signal to exit..."
trap "exit 1" HUP INT PIPE QUIT TERM
