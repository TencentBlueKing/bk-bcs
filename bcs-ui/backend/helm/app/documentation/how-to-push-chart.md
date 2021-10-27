## 推送业务 Helm Chart到仓库
开始之前，本文假设您已将业务部署方案改成 Helm Chart 格式。
为了方便您快速上手，本文将以[蓝鲸小游戏 Chart (rumpetroll)]({{ rumpetroll_demo_url }})为例, 说明如何推送 Chart 到仓库。

注意事项：文档内容根据项目生成，其中的账号信息均为项目的真实账号，请妥善保管。


### 1 安装 Helm
  - 安装 Helm
    + 方式一：包管理工具

     ```
     # package manager for Mac
     brew install helm

     # package manager for Windows
     choco install kubernetes-helm

     # cross-platform systems package manager
     gofish install helm
     ```

    + 方式二：手动下载二进制
        + [下载地址](https://github.com/helm/helm/releases/tag/v3.5.4), 下载与您操作系统对应的版本

  - 安装 Helm Chart 推送工具

    + 方式一：命令安装
    ```
    helm plugin install https://github.com/chartmuseum/helm-push
    ```

    + 方式二：手动下载二进制
        + [下载地址](https://github.com/chartmuseum/helm-push/releases)

### 2 添加 Helm Chart 仓库
  + 注意：仓库的账号密码为项目私有，请妥善保管

    ```
    helm repo add {{ project_code }} {{ repo_url }} --username={{ username }} --password={{ password }}
    ```

### 3 推送 Helm Chart
- 准备 Chart

下面将以蓝鲸小游戏的部署 Chart 为例，说明推送Chart。如果项目已经有 Chart，可以直接使用项目的 Chart。

```
wget {{ rumpetroll_demo_url }}
tar -xf rumpetroll.tgz
```

- 推送 Chart

如果 `push` 插件版本大于等于0.10.0，必须使用如下命令

```
helm cm-push rumpetroll/ {{ project_code }}
```

其它版本，使用如下命令

```
helm push rumpetroll/ {{ project_code }}
```

- 成功推送 Chart 后，可看到类似如下输出:

```
Pushing rumpetroll-0.1.22.tgz to {{ project_code }}...
Done.
```

### 4 同步项目 Chart 到产品展示页面
支持三种方式：

- 方式一：执行如下命令，同步 Chart 到产品中

```
curl -X POST {{ base_url }}bcs/k8s/configuration_noauth/{{ project_id }}/helm/repositories/sync/?format=json
```

注意：为保障你的正常使用，请不要使用 crontab 执行上述命令。

- 方式二: 页面中手动刷新, 在 "容器服务 -> 配置 -> Helm模板集" 页面中，点击 "同步仓库" 按钮触发
- 方式三: 每十分钟自动同步一次


## 温馨提示：故障排查
- CASE 1: 如果添加仓库成功了，推送 Chart 失败，错误码 411, 返回一段 Html 页面, 请先关闭代理再试试

```
Error: 411: could not properly parse response JSON:
<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN" "http://www.w3.org/TR/html4/loose.dtd">
<HTML><HEAD>
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=gb2312">
...
```

- CASE 2: Chart 版本已经存在, 如果出现如下 409 错误信息，请修改 Chart 版本号后重试。为保障您的数据安全，禁止对同一个版本重复推送。

```
Pushing rumpetroll-0.1.22.tgz to {{ project_code }}...
Error: 409: rumpetroll-0.1.22.tgz already exists
Error: plugin "push" exited with error
```
