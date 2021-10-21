FROM centos:7

# for command envsubset
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-cloud-netservice/swagger

COPY bcs-cloud-netservice bcs-cloud-netservice.json.template /data/bcs/bcs-cloud-netservice/
COPY container-start.sh /data/bcs/bcs-cloud-netservice/
COPY ./swagger-ui /data/bcs/bcs-cloud-netservice/swagger/
COPY cloudnetservice.swagger.json /data/bcs/bcs-cloud-netservice/swagger/

RUN chmod +x /data/bcs/bcs-cloud-netservice/container-start.sh

WORKDIR /data/bcs/bcs-cloud-netservice/
CMD [ "/data/bcs/bcs-cloud-netservice/container-start.sh" ]