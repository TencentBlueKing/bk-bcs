FROM centos:latest

RUN mkdir /gw && mkdir /gw/logs
COPY bcs-gw-controller /gw/gw-controller
RUN chmod +x /gw/gw-controller
WORKDIR /gw