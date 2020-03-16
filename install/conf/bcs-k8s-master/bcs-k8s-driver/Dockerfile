FROM centos:latest

USER root
ADD ./bcs-ca.crt .
ADD ./bcs-server.crt .
ADD ./bcs-server.key .
ADD bcs-k8s-driver .

RUN chmod +x bcs-k8s-driver

ENTRYPOINT ["./bcs-k8s-driver"]