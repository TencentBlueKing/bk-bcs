# bcs-ui 本地开发10分钟启动教程(helm) / 文档

使用helm chart部署bcs-ui应用

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



#### 构建**bcs-ui镜像

- **拉取[bk-bcs](https://github.com/Tencent/bk-bcs)项目，拉取完成后找到bcs-ui项目目录**

  ```
  git clone https://github.com/Tencent/bk-bcs.git
  cd bk-bcs/bcs-ui/
  ```

- **构建bcs-ui项目镜像**

  注意：编译前端和UI模块 要求 1.14 版本的 NodeJS（node版本不要太高，经验证1.16也可以）

  要求 1.17 版本的 golang

  **生成镜像**

  ```bash
  make build_frontend
  make build_ui
  make docker
  ```



#### helm chart部署bcs-ui项目

- **返回到bk-bcs项目根目录，找到部署bcs-ui项目的chart包**

  ```
  cd install/helm/bcs-ui
  ```

  


- **安装bcs-ui chart包**

  ```
  // 切换目录bcs-ui chart目录
  cd /data/workspace/src/bk-bcs/install/helm/bcs-ui/
  // ls命令确认下当前目录存在chart.yaml
  // 安装chart包, 执行helm install（部署到默认的命名空间或者加上 -n 名称，指定命名空间）
  
  helm install -n bcs-ui bcs-ui   --debug ./
  
  // 查看部署的状态
  helm status bcs-ui -n bcs-ui
  ```



