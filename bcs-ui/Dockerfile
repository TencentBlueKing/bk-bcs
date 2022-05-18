FROM python:3.6.8-stretch
WORKDIR /data

RUN apt-get update && apt-get install less

ENV HELM_LINK=http://bkopen-1252002024.file.myqcloud.com/bcs/kubectl-helm.tar.gz
ENV HELM_BASE_DIR=/usr/local/helm
ENV HELM_BIN_DIR=/usr/local/helm/bin
RUN mkdir -p $HELM_BIN_DIR && curl -sL $HELM_LINK | tar xzf - -C $HELM_BIN_DIR && chmod a+x $HELM_BIN_DIR/*

ADD . .
RUN pip install -r requirements.txt

# 因不存在 DB 服务，镜像构建时禁用 Django-Prometheus Migration
ARG PROMETHEUS_EXPORT_MIGRATIONS=False
# 镜像构建时需通过 build-arg 指定 BKPAAS_APP_ID, BKPAAS_APP_SECRET！
ARG BKPAAS_APP_ID
ARG BKPAAS_APP_SECRET
RUN python manage.py collectstatic --settings=backend.settings.helm.prod
