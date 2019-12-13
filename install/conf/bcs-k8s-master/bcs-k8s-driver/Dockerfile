FROM centos:latest

USER root
ADD start.sh .
ADD ./bcs-ca.crt .
ADD ./bcs-server.crt .
ADD ./bcs-server.key .
ADD bcs-k8s-driver .

RUN chmod +x start.sh && chmod +x bcs-k8s-driver

ENTRYPOINT ["./start.sh"]