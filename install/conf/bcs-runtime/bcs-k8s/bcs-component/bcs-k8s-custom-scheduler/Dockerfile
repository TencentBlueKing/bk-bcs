FROM tencentos/tencentos4-minimal

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-k8s-custom-scheduler 
ADD ./bcs-k8s-custom-scheduler /data/bcs/bcs-k8s-custom-scheduler/

ENTRYPOINT ["/data/bcs/bcs-k8s-custom-scheduler/bcs-k8s-custom-scheduler"]
