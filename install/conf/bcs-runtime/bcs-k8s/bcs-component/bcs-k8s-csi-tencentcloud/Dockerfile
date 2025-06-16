FROM tencentos/tencentos4-minimal

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-k8s-csi-tencentcloud/

COPY  bcs-k8s-csi-tencentcloud /data/bcs/bcs-k8s-csi-tencentcloud/

WORKDIR /data/bcs/bcs-k8s-csi-tencentcloud/
CMD ["/data/bcs/bcs-k8s-csi-tencentcloud/bcs-k8s-csi-tencentcloud"]
