FROM centos:7

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert

ADD bcs-cloud-netcontroller /data/bcs/bcs-cloud-netcontroller/
ADD container-start.sh /data/bcs/bcs-cloud-netcontroller/

RUN chmod +x /data/bcs/bcs-cloud-netcontroller/container-start.sh

WORKDIR /data/bcs/bcs-cloud-netcontroller/
CMD [ "/data/bcs/bcs-cloud-netcontroller/container-start.sh" ]