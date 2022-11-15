#!/bin/sh

# target app.
if [ $# != 3 ] ; then
    echo "Usage: $0 APPBIN APPARGS BINPATH"
    exit 1
fi

# variables.
APPBIN=$1
APPARGS=$2
BINPATH=$3

# escape character ‘/’ in app args.
APPARGS=$(echo "$APPARGS" | sed 's/\//\\\//g')

# tools.
TOOLS=$(dirname $(readlink -f "$0"))

# generate app daemon-control tool.
cp -rf ${TOOLS}/daemon-control.sh ${BINPATH}/$1.sh

sed -i "s/.*APPBIN=.*/APPBIN=\"${APPBIN}\"/" ${BINPATH}/$1.sh
sed -i "s/.*APPARGS=.*/APPARGS=\"${APPARGS}\"/" ${BINPATH}/$1.sh
sed -i "s/.*BINPATH=.*/BINPATH=\".\"/" ${BINPATH}/$1.sh
