FROM centos:latest

RUN mkdir -p /bcs/certs && mkdir /bcs/logs && mkdir /bcs/cni && mkdir /bcs/etc
COPY ./bcs-cloud-network-agent /bcs/bcs-cloud-network-agent
COPY ./bcs-eni /bcs/bcs-eni
COPY ./bcs-eni.conf /bcs/bcs-eni.conf
COPY ./start_agent.sh /bcs/start_agent.sh
RUN chmod +x /bcs/bcs-cloud-network-agent && chmod +x /bcs/bcs-eni && chmod +x /bcs/start_agent.sh
WORKDIR /bcs
