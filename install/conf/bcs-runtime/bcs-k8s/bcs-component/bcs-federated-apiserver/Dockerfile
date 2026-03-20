FROM tencentos/tencentos4-minimal
RUN yum install -y  ca-certificates
ADD apiserver /usr/bin
CMD ["/usr/bin/apiserver"]
