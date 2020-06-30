FROM centos:latest

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-service-prometheus-service

ADD bcs-service-prometheus-service /data/bcs/bcs-service-prometheus-service/
ADD container-start.sh /data/bcs/bcs-service-prometheus-service/
RUN chmod +x /data/bcs/bcs-service-prometheus-service/container-start.sh

WORKDIR /data/bcs/bcs-service-prometheus-service/
ENTRYPOINT ["/data/bcs/bcs-service-prometheus-service/container-start.sh"]

