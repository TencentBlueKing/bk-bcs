# bcs deployment 说明
## 1. bcs-deployment简介
bcs-deployment是基于bcs-application抽象出的顶层概念、主要满足应用的滚动升级，回滚，暂停，扩、缩容等需求。

## 2. json配置模板

```json
{
    "apiVersion": "v4",
    "kind": "deployment",
    "metadata": {
        "labels": {
            "label_deployment": "label_deployment"
        },
        "name": "deployment-test-001",
        "namespace": "defaultGroup"
    },
    "restartPolicy": {
        "policy": "Always",
        "interval": 5,
        "backoff": 10
    },
     "killPolicy":{
        "gracePeriod": 10
    },
    "constraint": {
        "IntersectionItem": [
            {
                "UnionData": [
                    {
                        "name": "hostname",
                        "operate": "CLUSTER",
                        "type": 4,
                        "set": {
                            "item": [
                                "mesos-slave-1",
                                "mesos-slave-2"
                            ]
                        }
                    }
                ]
            }
        ]
    },
    "spec": {
        "instance": 2,
        "selector": {
            "podname": "app-test-001"
        },
        "strategy": {
            "type": "RollingUpdate",
            "rollingupdate": {
                "maxUnavilable": 1,
                "maxSurge": 1,
                "upgradeDuration": 60,
                "rollingOrder": "CreateFirst",
                "rollingManually": false
            }
        },
        "template": {
            "metadata": {
                "labels": {
                    "label_deployment": "label_deployment"
                },
                "name": "deployment-test-001",
                "namespace": "defaultGroup"
            },
            "spec": {
                "containers": [
                    {
                        "command": "python",
                        "args": [
                            "-m",
                            "SimpleHTTPServer",
                            "8888"
                        ],
                        "parameters": [],
                        "type": "MESOS",
                        "env": [
                            {
                                "name": "DNS_HOSTS",
                                "value": "test_env"
                            }
                        ],
                        "image": "docker.hub.com/xxx/xxx:v1",
                        "imagePullUser": "xxxx",
                        "imagePullPasswd": "xxxxx",
                        "imagePullPolicy": "Always",
                        "privileged": false,
                        "ports": [
                            {
                                "containerPort": 8899,
                                "name": "test-port",
                                "protocol": "HTTP"
                            }
                        ],
                        "healthChecks": [
                            {
                                "protocol": "tcp",
                                "path": "/http/path/only",
                                "delaySeconds": 10,
                                "gracePeriodSeconds": 12,
                                "intervalSeconds": 10,
                                "timeoutSeconds": 10,
                                "consecutiveFailures": 10
                            }
                        ],
                        "resources": {
                            "limits": {
                                "cpu": "0.5",
                                "memory": "8"
                            }
                        },
                        "volumes": [],
                        "secrets": [],
                        "configmaps": []
                    }
                ],
                "networkMode": "BRIDGE",
                "networktype": "cnm"
            }
        }
    }
}

```
## 基础信息简介
由于bcs-deployment是基于bcs-application构建，其中的大部分信息与bcs-application一致，包含以下内容：
- restartPolicy
- killPolicy
- constraint
- spec.template中所有字段信息

关于这部分字段与结构的详细信息请见[这里](./bcs-application.md)。

`下面介绍一下关于bcs-deployment本身特性的相关策略。`
```json
"spec": {
    "instance": 2,
    "selector": {
        "podname": "app-test-001"
    },
    "strategy": {
        "type": "RollingUpdate",
        "rollingupdate": {
            "maxUnavilable": 1,
            "maxSurge": 1,
            "upgradeDuration": 60,
            "rollingOrder": "CreateFirst",
            "rollingManually":false
        }
    }
}
```
## 实例数
相关参数为spec.instance(第2行)，用于配置要创建的taskgroup的数量。该taskgroup是由deployment创建的一个application来管理。

## application选择器
相关参数为spec.selector（第3行），用于配置deployment所需要管理的bcs-appliction,默认这些bcs-application是由bcs-deployment自动创建的。

## deployment升级策略
相关配置项为spec.strategy（第6-15行），用于配置deployment执行rolling操作时所需要的策略：
- type: 定义deployment进行rolling时要选择的策略，目前只支持RollingUpdate：
  - `RollingUpdate`:
    RollingUpdate 即为滚动升级，该策略允许我们对滚动操作的过程中每次新创建的容器数量，删除的容器数量，创建间隔等策略进行控制。当原有的taskgroup全部删除，新的taskgroup（个数通过instances参数定义）全部创建，则update结束。

- RollingUpdate
可以对Rolling的操作进行详细的配置，包含以下参数：
  - `maxUnavilable`:
  决定了每个rolling周期内可以`删除`的taskgroup数量。如果原有的taskgroup已经全部删除，则后续每一次rolling中不会再删除taskgroup。
  - `maxSurge`:
  决定了每个rolling周期内可以`创建`的taskgroup数量。如果新的taskgroup已经全部创建，则后续每一次rolling中不会再创建taskgroup。
  - `upgradeDuration`:
  配置每次rolling操作之间的`最小`间隔时间。
  - `rollingOrder`:
  配置在进行每次rolling操作期间的每个周期内，创建和删除应用的先后顺序。该配置支持两种模式`CreateFirst`, `DeleteFirst`。 **CreateFirst**策略会先创建新的应用，然后删除老的应用。而**DeleteFirst**策略会先删除老的应用，再创建新的应用。
  - `rollingManually`:
  配置每次滚动是否需要手动触发，默认为false，即一次滚动完成之后在时间间隔结束之后自动进行下一次滚动，如果配置为true，则在每次滚动后自动pause，需输入resume命令才会在时间间隔结束之后进行下一次滚动

## Note
创建deployment时，如果deployment（通过selector）关联的application已经存在，则会delete掉现有的application，并根据spec.template创建新的application。
如果不想更新application，仅仅只是做deployment与application的关联，则填写json时，spec.template不填。注意：不是spec.template:{}，而是该字段不填写。

## rolling update的策略示例
我们假设rolling前deployment的instances为oldInstances, rolling的deployment instances为newInstances。可以预见在rolling过程中会有以下三种场景：
- oldInstances < newInstances
对application进行一次rolling update，并且达到扩容的效果
- oldInstances = newInstances
对application进行一次rolling update，容器数量保持不变
- oldInstances > newInstances
对application进行一次rolling update，并且达到缩容的效果

## 支持的deployment操作
- create
创建一个deployment：如果已有绑定的application，则delete当前application并创建新的application；如果不存在绑定的application，则创建application。也可以创建一个空的deployment绑定到当前已有的application。
- udpate
对deployment进行rolling update： 通过instances的变化，可以同时到达扩容和缩容的效果。
- rollback
停止当前的update操作，将新创建的application和taskgroup删除，将原有的application的taskgroup恢复到原有的instances个数
- pause
暂停rolling update： rolling update过程中可以通过该命令暂停update。
- resume
继续rolling update： 可以将暂停的update继续。
- delete
删除deployment，以及相应的application