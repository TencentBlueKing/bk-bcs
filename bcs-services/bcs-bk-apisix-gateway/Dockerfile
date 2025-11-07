FROM hub.bktencent.com/blueking/bk-apisix-oss:v3.13.0.2

RUN chmod 777 /usr/local/apisix/logs

#install apisix plugin
COPY ./plugins/ /usr/local/apisix/extra_plugins/blueking/apisix/plugins/