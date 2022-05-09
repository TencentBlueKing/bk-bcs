# Helm部署Jaeger

本文主要介绍如何使用Helm部署以ES作为后端的Jaeger

*注：以部署**jaeger-0.52.0.tgz**版本的helm包为例*

## 环境需求

- Helm 3.2.0+（如果小于该版本，可能会导致安装失败）
- 需要可用的Persistent Volumes，默认情况下需要5个8Gi的PV(如果ES集群已存在，则不需要)

## 下载Helm包

### 下载Jaeger Helm包

添加Jaeer Tracing Helm repo：

```
helm repo add jaegertracing https://jaegertracing.github.io/helm-charts
```

下载Jaeger Helm包：

```
helm pull --version 0.52.0 jaegertracing/jaeger
```

### 下载依赖包

如果想要安装Cassandra，Elasticsearch，Kafka，则需要添加相应的repo源：

```
# 添加Cassandra repo
helm repo add incubator https://charts.helm.sh/incubator
# 添加ES、Kafka repo
helm repo add bitnami https://charts.bitnami.com/bitnami
# 下载依赖包
helm pull [--version <version>] bitnami/cassandra
helm pull [--version <version>] bitnami/elasticsearch
helm pull [--version <version>] bitnami/kafka
```

**替换依赖包**

由于下载的Jaeger helm包有默认的charts依赖包，我们需要把它们替换为上面下载好的bitnami repo中的charts包。其实就是解压Jaeger、cassandra、elasticsearch、kafka helm包，并用解压出来的cassandra、elasticsearch、kafka替换jaeger的charts目录下原有的charts。

```
# tar xf jaeger-0.52.0.tgz
# rm -rf jaeger/charts/*
# tar xf common-1.10.3.tgz
# tar xf elasticsearch-17.6.1.tgz
# tar xf kafka-14.9.0.tgz
# tar xf cassandra-9.1.0.tgz
# mv common elasticsearch kafka cassandra jaeger/charts
```

## 配置ES

首先进入到上面准备好的jaeger helm包所在的目录。

- 如果环境中没有ES集群，则使用以下命令部署jaeger和ES：

```
helm install jaeger jaeger-0.52.0 \
  --set provisionDataStore.cassandra=false \
  --set provisionDataStore.elasticsearch=true \
  --set storage.type=elasticsearch
```

- 如果不是将jaeger部署到default namespace，则需要使用以下命令：

```
helm install -n <namespace> jaeger jaeger-0.52.0 \
  --set provisionDataStore.cassandra=false \
  --set provisionDataStore.elasticsearch=true \
  --set storage.type=elasticsearch \
  --set storage.elasticsearch.host=<namespace>-elasticsearch-master
```

- 如果想以现有的ES集群做后端,则执行：

```
helm install jaeger jaeger-0.52.0\
  --set provisionDataStore.cassandra=false \
  --set storage.type=elasticsearch \
  --set storage.elasticsearch.host=<HOST> \
  --set storage.elasticsearch.port=<PORT> \
  --set storage.elasticsearch.user=<USER> \
  --set storage.elasticsearch.password=<password>
```

- 如果现有ES集群配置了TLS，可以使用以下yaml文件:

```
storage:
  type: elasticsearch
  elasticsearch:
    host: <HOST>
    port: <PORT>
    scheme: https
    user: <USER>
    password: <PASSWORD>
provisionDataStore:
  cassandra: false
  elasticsearch: false
query:
  cmdlineParams:
    es.tls.ca: "/tls/es.pem"
  extraConfigmapMounts:
    - name: jaeger-tls
      mountPath: /tls
      subPath: ""
      configMap: jaeger-tls
      readOnly: true
collector:
  cmdlineParams:
    es.tls.ca: "/tls/es.pem"
  extraConfigmapMounts:
    - name: jaeger-tls
      mountPath: /tls
      subPath: ""
      configMap: jaeger-tls
      readOnly: true
spark:
  enabled: true
  cmdlineParams:
    java.opts: "-Djavax.net.ssl.trustStore=/tls/trust.store -Djavax.net.ssl.trustStorePassword=changeit"
  extraConfigmapMounts:
    - name: jaeger-tls
      mountPath: /tls
      subPath: ""
      configMap: jaeger-tls
      readOnly: true
```

更新 jaeger-tls configmap:

```
# keytool -import -trustcacerts -keystore trust.store -storepass changeit -alias es-root -file es.pem
# kubectl create configmap jaeger-tls --from-file=trust.store --from-file=es.pem
# helm install jaeger jaeger-0.52.0 --values jaeger-values.yaml
```

## 配置Ingester（可选）

如果需要Kafka，则安装时还需要设置下面的参数。

- 使用新的Kafka集群

```
helm install jaeger jaeger-0.52.0 \
  --set provisionDataStore.kafka=true \
  --set ingester.enabled=true
```

- 使用已存在的Kafka集群

```
helm install jaeger jaeger-0.52.0 \
  --set ingester.enabled=true \
  --set storage.kafka.brokers={<BROKER1:PORT>,<BROKER2:PORT>} \
  --set storage.kafka.topic=<TOPIC>
```