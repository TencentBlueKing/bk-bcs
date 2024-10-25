#!/bin/bash
set -e

if [ -d $TONGSUO_PATH ]; then
  echo "tongsuo already exists"
  exit 0
fi

# determine package manager
determine_package_manager() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        if [[ "$ID" == "debian" || "$ID_LIKE" == *"debian"* ]]; then
            echo "apt-get"
        elif [[ "$ID" == "rhel" || "$ID" == "centos" || "$ID_LIKE" == *"rhel"* ]]; then
            echo "yum"
        else
            echo "unknown"
        fi
    else
        echo "unknown"
    fi
}

# install packages for different OS
install_package() {
    local package_manager=$(determine_package_manager)

    if [ "$package_manager" == "apt-get" ]; then
        echo "Using apt-get to install gcc"
        apt-get update
        apt-get install -y gcc
    elif [ "$package_manager" == "yum" ]; then
        echo "Using yum to install glibc-static"
        yum install -y glibc-static
    else
        echo "Could not determine package manager. Please install manually."
    fi
}

wget --no-check-certificate https://github.com/Tongsuo-Project/Tongsuo/archive/refs/tags/8.3.2.tar.gz
tar zxvf 8.3.2.tar.gz > /dev/null
cd Tongsuo-8.3.2/

if [ "$IS_STATIC" == true ]; then
  install_package
  ./config --prefix=$TONGSUO_PATH -static -fPIC
else
  ./config --prefix=$TONGSUO_PATH -fPIC
fi

# quiet output
make -j >/dev/null 2>&1
make install >/dev/null 2>&1
