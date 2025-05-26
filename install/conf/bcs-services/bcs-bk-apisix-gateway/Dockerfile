FROM bk-apisix:v0.0.3

RUN chmod 777 /usr/local/apisix/logs

#install apisix plugin
COPY ./plugins/ /usr/local/apisix/extra_plugins/blueking/apisix/plugins/
