FROM centos:latest

RUN mkdir /stgw && mkdir /stgw/logs
COPY bcs-stgw-controller /stgw/bcs-stgw-controller
RUN chmod +x /stgw/bcs-stgw-controller
WORKDIR /stgw