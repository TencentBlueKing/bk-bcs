# Use cent os as base image
FROM centos:7
# for command envsubset
RUN yum install -y gettext
RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-monitor-controller

# empty.tar.gz是一个包含空文件夹的压缩包， 删除某个CR时，通过上传empty到蓝鲸监控来删除监控上对应的数据
COPY empty.tar.gz /data/bcs
ADD bcs-monitor-controller /data/bcs/bcs-monitor-controller/
ADD container-start.sh /data/bcs/bcs-monitor-controller/
RUN chmod +x /data/bcs/bcs-monitor-controller/container-start.sh

WORKDIR /data/bcs/bcs-monitor-controller/
CMD ["/data/bcs/bcs-monitor-controller/container-start.sh"]