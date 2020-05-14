FROM centos:latest

RUN mkdir /bcs && mkdir /bcs/logs
COPY bcs-networkpolicy /bcs/bcs-networkpolicy
RUN chmod +x /bcs/bcs-networkpolicy
WORKDIR /bcs
