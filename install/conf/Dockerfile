FROM centos:latest

LABEL Author="DeveloperJim"
#all bcs modules in on container for testing

RUN mkdir -p /data/bcs/logs /data/bcs/cert

# bcs-service installation
COPY ./bcs-services/* /data/bcs/
#gateway installation

# bcs-k8s installation
COPY ./bcs-k8s-master/* /data/bcs/
#COPY ./bcs-k8s-node/* /data/bcs/

#config type: render, sidecar, off
# > render: container start script render configuration file with ENV
# > sidecar: supported by BSCP
# > off: do nothing
ENV BCS_CONFIG_TYPE=off
ENV BCS_HOME="/data/bcs"
# bcs common
ENV log_dir="${BCS_HOME}/logs/bcs"
ENV pid_dir="/var/run/bcs"
ENV caFile="${BCS_HOME}/cert/bcs/commutity-ca.crt"
ENV serverCertFile="${BCS_HOME}/cert/bcs/commutity-server.crt"
ENV serverKeyFile="${BCS_HOME}/cert/bcs/commutity-server.key"
ENV clientCertFile="${BCS_HOME}/cert/bcs/commutity-client.crt"
ENV clientKeyFile="${BCS_HOME}/cert/bcs/commutity-client.key"
#!import for module discovery
ENV localIp=127.0.0.1

WORKDIR /data/bcs
