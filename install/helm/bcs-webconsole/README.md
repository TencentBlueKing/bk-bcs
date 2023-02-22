# 本地启动helm部署bcs-webconsole 教程文档

使用helm chart部署cluster-webconsole应用

#### 部署环境

```
操作系统环境：CentOS Linux 7 (Core)
golang: go1.17
docker环境：20.10.23
Kubernetes：v1.23.8
helm: v3.11.0
```

#### 构建**bcs-webconsole**镜像

- **拉取[bk-bcs](https://github.com/Tencent/bk-bcs)项目，拉取完成后找到bcs-webconsole项目目录**

  ```
  git clone https://github.com/Tencent/bk-bcs.git
  cd bk-bcs/bcs-services/bcs-webconsole/
  ```

- 构建bcs-webconsol项目镜像

  ```
  // Build the binary
  make build
  
  // Build a docker image
  make docker
  ```

#### 使用helm chart部署bcs-webconsole项目

- **返回到bk-bcs项目根目录，找到部署bcs-webconsole项目的chart包**

  ```
  cd install/helm/bcs-webconsole
  ```

- **bcs-webconsole chart配置参数都在Value.yaml**

  *bcs-webconsole配置*

  | **Name**                   | Description                                                  | Value          |
  | -------------------------- | ------------------------------------------------------------ | -------------- |
  | mage.repository            | 镜像名称                                                     | bcs-webconsole |
  | image.pullPolicy           | 镜像拉取策略，由于本地存在bcs-webconsole项目镜像所以默认Never | Never          |
  | image.tag                  | 镜像版本                                                     | latest         |
  | svcConf.base_conf.app_code |                                                              | false          |
  | svcConf.bcs_conf.host      |                                                              |                |
  | envs                       | 环境变量，DEMO: "test"                                       |                |

  *redis配置参数*

  | **Name**           | Description                                    | Value      |
  | ------------------ | ---------------------------------------------- | ---------- |
  | redis.architecture | redis架构。允许值：`standalone`或`replication` | standalone |
  | redis.auth.enabled | 启用密码验证                                   | false      |

  *etcd配置参数*

  | **Name**              | Description                               | Value  |
  | --------------------- | ----------------------------------------- | ------ |
  | etcd.replicaCount     | 要部署的etcd副本的数量                    | 3      |
  | etcd.auth.rbac.create | 启用RBAC认证                              | false  |
  | etcd.auth.token.type  | 认证令牌类型。允许的值。'simple'或'jwt'。 | simple |

- **安装bcs-webconsole chart包**

  ```
  // 找到bcs-webconsole chart目录
  // 安装chart包, 执行helm install（部署到默认的命名空间或者加上 -n 名称，指定命名空间）
  helm install bcs-webconsole ./bcs-webconsole -n default
  
  // 首次部署该chart会提示需要依赖，执行一下
  helm dependency build ./bcs-webconsole
  
  // 再执行安装
  helm install bcs-webconsole ./bcs-webconsole -n default
  
  // 查看部署的状态
  helm status bcs-webconsole -n default
  
  // 卸载该应用
  helm uninstall bcs-webconsole -n default
  ```