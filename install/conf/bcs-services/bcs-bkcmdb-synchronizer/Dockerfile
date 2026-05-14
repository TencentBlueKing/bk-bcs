FROM tencentos/tencentos4-minimal

# Update package lists and install gettext
RUN yum install -y gettext sqlite

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-bkcmdb-synchronizer/db

ADD bcs-bkcmdb-synchronizer /data/bcs/bcs-bkcmdb-synchronizer/
ADD bcs-bkcmdb-synchronizer.json.template /data/bcs/bcs-bkcmdb-synchronizer/
ADD container-start.sh /data/bcs/bcs-bkcmdb-synchronizer/

RUN chmod +x /data/bcs/bcs-bkcmdb-synchronizer/bcs-bkcmdb-synchronizer
RUN chmod +x /data/bcs/bcs-bkcmdb-synchronizer/container-start.sh

WORKDIR /data/bcs/bcs-bkcmdb-synchronizer/
