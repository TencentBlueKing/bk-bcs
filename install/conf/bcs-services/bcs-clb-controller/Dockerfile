FROM centos:latest

RUN mkdir /clb && mkdir /clb/logs
COPY bcs-clb-controller /clb/clb-controller
RUN chmod +x /clb/clb-controller
WORKDIR /clb
