FROM centos:latest

RUN mkdir -p /data/bcs/cert /data/bcs/logs/bcs
RUN mkdir /data/bcs/bcs-clb-controller/logs
COPY bcs-clb-controller /data/bcs/bcs-clb-controller/
RUN chmod +x /data/bcs/bcs-clb-controller/bcs-clb-controller
WORKDIR /data/bcs/bcs-clb-controller
