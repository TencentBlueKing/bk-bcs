#!/bin/bash
set -e

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
# Skip when the package is already present: installing it would also upgrade the C
# toolchain to the newest version in the repo, and swapping gcc/glibc files mid-build
# breaks the cgo compilations that other `make -j` targets run at the same time.
install_package() {
    local package_manager=$(determine_package_manager)

    if [ "$package_manager" == "apt-get" ]; then
        if command -v gcc >/dev/null 2>&1; then
            echo "gcc already installed"
            return 0
        fi
        echo "Using apt-get to install gcc"
        apt-get update
        apt-get install -y gcc
    elif [ "$package_manager" == "yum" ]; then
        if rpm -q glibc-static >/dev/null 2>&1; then
            echo "glibc-static already installed"
            return 0
        fi
        echo "Using yum to install glibc-static"
        yum install -y glibc-static
    else
        echo "Could not determine package manager. Please install manually."
    fi
}

if [ -d $TONGSUO_PATH ]; then
  echo "tongsuo already exists"
  exit 0
fi

if ! [ -e 8.3.2.tar.gz ]; then
  wget --no-check-certificate https://github.com/Tongsuo-Project/Tongsuo/archive/refs/tags/8.3.2.tar.gz
fi

if ! [ -d Tongsuo-8.3.2 ]; then
  tar zxvf 8.3.2.tar.gz > /dev/null
fi

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