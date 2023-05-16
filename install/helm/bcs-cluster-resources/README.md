# bcs-cluster-resources 本地开发10分钟启动教程(helm) / 文档

使用helm chart部署cluster-resources应用

#### 部署环境

```
操作系统环境：CentOS Linux 7 (Core)
golang: go1.17
docker环境：20.10.23
Kubernetes：v1.23.8
helm: v3.11.0
```



#### 构建**cluster-resources**镜像

- **拉取[bk-bcs](https://github.com/Tencent/bk-bcs)项目，拉取完成后找到cluster-resources项目目录**

  ```
  git clone https://github.com/Tencent/bk-bcs.git
  cd bk-bcs/bcs-services/cluster-resources/
  ```

  ![image-20230220112257699](images\image-20230220112257699.png)

- **构建cluster-resources项目镜像**

  **注意：**由于本地开发环境，所以构建出来的镜像也需要设置开发环境，更改Dockerfile文件，把原来的make build改成make dev

  ```
  // 原来的
  RUN make build VERSION=$VERSION GITCOMMIT=$GITCOMMIT
  
  // 改成
  RUN make dev VERSION=$VERSION GITCOMMIT=$GITCOMMIT
  ```

- **生成镜像**

  ```
  make docker
  ```

#### helm chart部署cluster-resources项目

- **返回到bk-bcs项目根目录，找到部署bcs-cluster-resources项目的chart包**

  ```
  cd install/helm/bcs-cluster-resources
  ```

  ![image-20230220141411726](images\image-20230220141411726.png)

- **bcs-cluster-resources chart配置参数都在Value.yaml中，一般只需要修改cluster-resources配置参数即可运行**

  *cluster-resources配置参数*

  | **Name**         | Description                                                  | Value                 |
  | ---------------- | ------------------------------------------------------------ | --------------------- |
  | mage.repository  | 镜像名称                                                     | bcs-cluster-resources |
  | image.pullPolicy | 镜像拉取策略，由于本地存在cluster-resources项目镜像所以默认Never | Never                 |
  | image.tag        | 镜像版本                                                     | latest                |
  | certs.enabled    | 证书挂载目录，默认关闭                                       | false                 |

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

- **安装bcs-cluster-resources chart包**

  ```
  // 找到bcs-cluster-resources chart目录
  // 安装chart包, 执行helm install（部署到默认的命名空间或者加上 -n 名称，指定命名空间）
  helm install bcs-cluster-resources ./bcs-cluster-resources
  
  // 首次部署该chart会提示需要依赖，执行一下
  helm dependency build ./bcs-cluster-resources
  
  // 再执行安装
  helm install bcs-cluster-resources ./bcs-cluster-resources
  
  // 查看部署的状态
  helm status bcs-cluster-resources -n default
  ```

![image-20230220143500057](images\image-20230220143500057.png)

![image-20230220143545078](images\image-20230220143545078.png)
