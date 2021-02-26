# bcs-watch部署
bcs-watch使用kubelet托管，所以需要提供k8s yaml文件。

## /etc/kubernetes/manifests/bcs-datawatch.yaml
```
apiVersion: v1
kind: Pod
metadata:
  name: bcs-datawatch
  namespace: kube-system
  labels:
    app: datawatch
spec:
  hostNetwork: true
  containers:
    - name: bcs-datawatch
      image: dockerhub.com:8443/public/k8s/bcs/datawatch:{TAG}
      imagePullPolicy: IfNotPresent
      volumeMounts:
        - name: watcherconfig
          mountPath: /config/
          readOnly: true
  volumes:
    - name: watcherconfig
      hostPath:
        path: /data/bcs/kubeops/bcs/datawatch/config/
```
- {TAG}: image tag

## /data/bcs/kubeops/bcs/datawatch/config/config.json
```
{
  "default": {
    "clusterIDSource": "config",
    "clusterID": "",
    "hostIP": "x.x.x.x"
  },
  "bcs": {
  "zk": ["zk-1:2181","zk-2:2181"],
    "tls": {
      "ca-file": "/cert/bcs-inner-ca.crt",
      "cert-file": "/cert/bcs-inner-client.crt",
      "key-file": "/cert/bcs-inner-client.key",
      "password": ""
    },
    "custom-storage-endpoints": null
  },
  "k8s": {
  "master": "http://127.0.0.1:8080",
    "tls": {
      "ca-file": "",
      "cert-file": "",
      "key-file": ""
    }
  }
}
```
- clusterID: 集群id，跨云部署时需要填写
- hostIP：物理机ip
- zk: bcs-services zk地址，用于服务发现
- custom-storage-endpoints: 自定义storage地址，跨云部署时需要填写，例如：["https://x.x.x.x:11005","https://x.x.x.x:11005"]
- master: kube-apiserver端口