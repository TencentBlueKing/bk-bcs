# bcs-bscp 本地开发10分钟启动教程(helm) / 文档

使用helm chart部署bcs-bscp应用

#### 部署环境

```
操作系统环境：CentOS Linux 7 (Core)
go               1.17
Docker           20.10.17
k8s(minikube)    1.23.5
git              2.38.1      
helm             3.9.4
node             16.19.1
npm              8.19.3
```



#### 构建**bcs-bscp**镜像

- **拉取[bk-bcs](https://github.com/Tencent/bk-bcs)项目，拉取完成后找到bcs-bscp项目目录**

  ```
  git clone https://github.com/Tencent/bk-bcs.git
  cd bk-bcs/bcs-services/bcs-bscp/
  ```

- **构建bcs-bscp项目镜像**

  注意：编译前端和UI模块 要求 1.14 版本的 NodeJS（node版本不要太高，经验证1.16也可以）

  要求 1.17 版本的 golang

  编译 pb

  ```bash
  # 下载正确的 protoc 二进制版本到 .bin 目录
  make init
  
  # 把 .bin/protoc 加到路径中
  export PATH=`pwd`/.bin:$PATH
  
  # 创建 bscp.io 软连接, 已经 gitignore 了这个文件，不会提交到 git 库
  cd .. && ln -sf bcs-bscp bscp.io && cd bscp.io
  
  # 前面的步骤一次性， OK后编译
  make pb
  ```

  编译二进制

  ```bash
  make build_bscp
  ```

  编译前端和UI模块

  ```bash
  make build_frontend
  make build_ui
  ```

  

  **生成镜像**

  ```
  make docker
  ```

#### helm chart部署bcs-bscp项目

- **返回到bk-bcs项目根目录，找到部署bcs-cluster-resources项目的chart包**

  ```
  cd install/helm/bcs-bscp
  ```

- **bcs-bcs-bscp chart配置参数都在Value.yaml中，一般只需要修改bcs-bscp配置参数即可运行**

  *cluster-resources配置参数*

  | **Name**            | Description                                                  | Value    |
  | ------------------- | ------------------------------------------------------------ | -------- |
  | image.repository    | 镜像名称                                                     | bcs-bscp |
  | image.pullPolicy    | 镜像拉取策略，由于本地存在cluster-resources项目镜像所以默认Never | Never    |
  | image.tag           | 镜像版本                                                     | latest   |
  | credentials.enabled | 证书挂载目录，默认关闭                                       | false    |

  

  *mariadb配置参数*

  | **Name**                              | Description                                                  | Value     |
  | ------------------------------------- | ------------------------------------------------------------ | --------- |
  | mariadb.auth.rootPassword             | redis架构。允许值：`standalone`或`replication`               | root      |
  | mariadb.primary.service.type=NodePort | mariadb服务暴露的方式（建议选NodePort，有个sql初始化脚本需要执行） | NodePort  |
  | mariadb.initdbScriptsConfigMap        | ConfigMap with the initdb scripts                            | initdb_cm |
  | mariadb.auth.database                 | Name for a custom database to create                         | bscp      |

  *redis配置参数*

  | **Name**                  | Description                                    | Value      |
  | ------------------------- | ---------------------------------------------- | ---------- |
  | redis.architecture        | redis架构。允许值：`standalone`或`replication` | standalone |
  | redis.auth.enabled        | 启用密码验证                                   | false      |
  | redis.auth.password       | root账号密码                                   |            |
  | redis.master.service.type | 服务暴露的方式                                 | ClusterIP  |

  *etcd配置参数*

  | **Name**                 | Description                               | Value  |
  | ------------------------ | ----------------------------------------- | ------ |
  | etcd.replicaCount        | 要部署的etcd副本的数量                    | 1      |
  | etcd.auth.rbac.create    | 启用RBAC认证                              | false  |
  | etcd.auth.token.type     | 认证令牌类型。允许的值。'simple'或'jwt'。 | simple |
  | etcd.persistence.enabled | 是否开启持久化                            | false  |

- **安装bcs-bscp chart包**

  ```
  // 切换目录bcs-bscp chart目录
  cd /data/workspace/src/bk-bcs/install/helm/bcs-bscp/
  // ls命令确认下当前目录存在chart.yaml
  // 安装chart包, 执行helm install（部署到默认的命名空间或者加上 -n 名称，指定命名空间）
  
  helm install -n bcs-bscp bcs-bscp   --debug ./
  
  // 首次部署该chart会提示需要依赖，执行一下
  helm dependency build 
  
  // 再执行安装
  helm install -n bcs-bscp bcs-bscp   --debug ./
  
  // 查看部署的状态
  helm status bcs-cluster-resources -n default
  ```



