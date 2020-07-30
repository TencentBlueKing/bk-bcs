FROM centos:7

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-cloud-netservice/swagger

ADD bcs-cloud-netservice /data/bcs/bcs-cloud-netservice/
ADD container-start.sh /data/bcs/bcs-cloud-netservice/
ADD ./swagger-ui /data/bcs/bcs-cloud-netservice/swagger/
ADD cloudnetservice.swagger.json /data/bcs/bcs-cloud-netservice/swagger/

RUN chmod +x /data/bcs/bcs-cloud-netservice/container-start.sh

WORKDIR /data/bcs/bcs-cloud-netservice/
CMD [ "/data/bcs/bcs-cloud-netservice/container-start.sh" ]