# BCS PROJECT

BCS 项目管理，提供项目的CRUD服务

## 构建镜像
```
docker build . -t docker.io/bcs-poroject:v0.1.0
```
## 快速部署
允许使用chart方式或者二进制方式部署服务


### TODO chart部署

### 二进制启动

```
# 在项目的根目录执行
make build
./bcs-project-service -c project.yaml
```
