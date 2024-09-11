FROM ubuntu:22.04

# Update package lists and install gettext
RUN apt-get update && apt-get install -y gettext sqlite3 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-bkcmdb-synchronizer/db

ADD bin/bcs-bkcmdb-synchronizer /data/bcs/bcs-bkcmdb-synchronizer/
ADD config/bcs-bkcmdb-synchronizer.json.template /data/bcs/bcs-bkcmdb-synchronizer/
ADD config/container-start.sh /data/bcs/bcs-bkcmdb-synchronizer/

RUN chmod +x /data/bcs/bcs-bkcmdb-synchronizer/bcs-bkcmdb-synchronizer
RUN chmod +x /data/bcs/bcs-bkcmdb-synchronizer/container-start.sh

WORKDIR /data/bcs/bcs-bkcmdb-synchronizer/