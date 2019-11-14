FROM centos:latest

USER root
ADD start.sh .
ADD ./cert .
ADD bcs-k8s-driver .

RUN chmod +x start.sh && chmod +x bcs-k8s-driver

ENTRYPOINT ["./start.sh"]