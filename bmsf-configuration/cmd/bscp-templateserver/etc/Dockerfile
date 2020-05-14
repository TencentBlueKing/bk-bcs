FROM centos:latest

#init project directory
RUN mkdir -p /bk-bscp/etc
RUN mkdir -p /bk-bscp/log

COPY bk-bscp-templateserver /bk-bscp/
COPY etc/server.yaml /bk-bscp/etc/

#fix DNS BUG
#RUN [ ! -e /etc/nsswitch.conf ] && echo 'hosts: files dns' > /etc/nsswitch.conf

#this seems dumb, but the libc from the build stage is not the same as the alpine libc
#create a symlink to where it expects it since they are compatable. https://stackoverflow.com/a/35613430/3105368
#RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

#time zone setting
ENV TimeZone=Asia/Shanghai   
RUN ln -snf /usr/share/zoneinfo/$TimeZone /etc/localtime && echo $TimeZone > /etc/timezone

WORKDIR /bk-bscp/
CMD ["/bk-bscp/bk-bscp-templateserver", "run"]