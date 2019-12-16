FROM centos:latest

COPY ./bcs-k8s-watch /
RUN mkdir /cert
ADD ./bcs-ca.crt /cert/
ADD ./bcs-client.crt /cert/
ADD ./bcs-client.key /cert/

ENTRYPOINT ["/bcs-k8s-watch", "--config", "/config/config.json"]
