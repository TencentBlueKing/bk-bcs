# BscpConfig

> BscpConfig用于Bscp定义结构化配置。

## 使用 BscpConfig

常见的 BscpConfig 配置示例如下

```yaml
apiVersion: bk.tencent.com/v1alpha1
kind: BscpConfig
metadata:
  name: "bscpconfig-simple"
  namespace: default
  labels: # 灰度发布使用
    region: "ap-guangzhou2"
spec:
  provider:
    feedAddr: "feed.bscp.example.com:9510"
    biz: 100 # 业务
    token: "xxx" # 客户端秘钥
    app: "bcs-gateway-certs" # bscp app的名称

  configSyncer:
    - configmapName: "" # 生成固定名称configmap
      matchConfigs: ["*"] # 配置项匹配规则, 可使用linux wilecard语法


    - configmapName: "" # 生成固定名称configmap
      matchConfigs: ["*credentials*", "xxx"] # 配置项匹配规则, 可使用linux wilecard语法

    - secretName: bcs-gateway-certs #生成固定名称secret
      type: kubernetes.io/tls # secret指定类型, 默认是Opaque
      data:
        - key: tls.key # secret data.key 名称
          refConfig: uat_tls_key # secret data.value, refConfig是精确配置项名称
        - key: tls.crt
          refConfig: uat_tls_ca
```
