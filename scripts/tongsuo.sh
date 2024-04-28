#!/bin/bash
set -e

if [ -d $TONGSUO_PATH ]; then
  echo "tongsuo already exists"
  exit 0
fi

#wget --no-check-certificate https://github.com/Tongsuo-Project/Tongsuo/archive/refs/tags/8.3.2.tar.gz
#tar zxvf 8.3.2.tar.gz > /dev/null
cd Tongsuo-8.3.2/

if [ "$IS_STATIC" == true ]; then
  yum -y install glibc-static
  ./config --prefix=$TONGSUO_PATH -static -fPIC
else
  ./config --prefix=$TONGSUO_PATH -fPIC
fi

# quiet output
make -j >/dev/null 2>&1
make install >/dev/null 2>&1
