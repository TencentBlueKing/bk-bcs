# bcs command配置说明

bcs command支持对deployment，application对应的容器下发command命令，例如：./reload.sh。

## json配置模板
```json
{
    "apiVersion": "v4",
    "kind": "Command",
    "spec": {
        "commandTargetRef": {
          "kind": "Deployment | Application",
          "name": "deployment-name",
          "namespace": "defaultGroup"
        },
        "taskgroups":[],
        "command":["/bin/bash","-c","ps -ef |grep game"],
        "env":[],
        "user":"root",
        "workingDir":"",
        "privileged":false,
        "reserveTime": 60
    }
}
```

字段含义：(field comment)

- commandTargetRef: 执行command命令的应用
  - kind：类型，Deployment或Application
  - name：应用名称
  - namespace： 应用命名空间

- taskgroups：应用的taskgroup id列表，如果不填默认所有taskgroup
- command：命令
- env：环境变量
- user：执行command的用户，默认值是root
- workingDir：工作目录
- privileged：容器特权模式
- reserveTime：任务信息保存时间，单位是minutes，默认值：24x60x7
