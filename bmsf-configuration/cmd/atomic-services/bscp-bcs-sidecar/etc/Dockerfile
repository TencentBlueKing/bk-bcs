FROM ubuntu:latest

#init project directory
RUN mkdir -p /bk-bscp/etc
RUN mkdir -p /bk-bscp/log
RUN mkdir -p /bk-bscp/tools

COPY bk-bscp-bcs-sidecar /bk-bscp/
COPY etc/sidecar.yaml /bk-bscp/etc/
COPY etc/init_task.sh /bk-bscp/

#BSCP gateway name
ENV BSCP_BCSSIDECAR_GW_HOSTNAME=gw.bkbscp.bk.com
ENV BSCP_BCSSIDECAR_GW_PORT=8080

#business required information
ENV BSCP_BCSSIDECAR_APPINFO_IP_ETH=eth0

#some session setting
ENV BSCP_BCSSIDECAR_SESSION_TIMEOUT=30s

#fix DNS BUG
#RUN [ ! -e /etc/nsswitch.conf ] && echo 'hosts: files dns' > /etc/nsswitch.conf

#this seems dumb, but the libc from the build stage is not the same as the alpine libc
#create a symlink to where it expects it since they are compatable. https://stackoverflow.com/a/35613430/3105368
#RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

#time zone setting
ENV TimeZone=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TimeZone /etc/localtime && echo $TimeZone > /etc/timezone

#add run permission
RUN chmod 755 /bk-bscp/init_task.sh
RUN chmod 755 /bk-bscp/bk-bscp-bcs-sidecar

WORKDIR /bk-bscp/
CMD ["/bk-bscp/init_task.sh"]
