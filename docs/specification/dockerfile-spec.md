# 容器打包规则

## 目录说明

bcs容器化模块相关目录结构需要遵循以下规则

* 模块目录规则: /data/bcs/$module, 例如: /data/bcs/bcs-storage
* 配置文件
  * 默认放置在模块目录下，命名为 $module.json, 例如 bcs-storage.json
  * 需要放置配置模板文件，$module.json.template
* logs目录: /data/bcs/logs/bcs
* cert目录: /data/bcs/cert

## 默认环境变量

为了保持与进程部署、配置渲染方式兼容，默认地，BCS模块启动需要使用配置文件。

配置文件渲染过程中，需要使用之前定义的环境变量。全环境变量参考可以查阅[参数列表](https://github.com/Tencent/bk-bcs/blob/master/scripts/base.env)

公共启动参数说明：

```shell
export BCS_HOME="/data/bcs"
# bcs common
export log_dir="${BCS_HOME}/logs/bcs"
export pid_dir="/var/run/bcs"
export caFile="${BCS_HOME}/cert/bcs/test-ca.crt"
export serverCertFile="${BCS_HOME}/cert/bcs/test-server.crt"
export serverKeyFile="${BCS_HOME}/cert/bcs/test-server.key"
export clientCertFile="${BCS_HOME}/cert/bcs/test-client.crt"
export clientKeyFile="${BCS_HOME}/cert/bcs/test-client.key"

#! localIp for module startup.
# it's better to get local IP address within start.sh
export localIp=127.0.0.1
```

每一个模块启动需要填充每个模块所需要的环境变量，用于渲染首模块的配置模板。

例如bcs-storage的配置模块

```shell
{
  "address": "${localIp}",
  "port": ${bcsStoragePort},
  "metric_port": ${bcsStorageMetricPort},
  "log_dir": "${log_dir}",
  "pid_dir": "${pid_dir}",
  "bcs_zookeeper": "${bcsZkHost}",
  "database_config_file": "${storageDbConfig}",
  "event_max_day": ${eventMaxDay},
  "event_max_cap": ${eventMaxCap},
  "alarm_max_day": ${alarmMaxDay},
  "alarm_max_cap": ${alarmMaxCap},
  "ca_file": "${caFile}",
  "server_cert_file": "${serverCertFile}",
  "server_key_file": "${serverKeyFile}"
}
```

在启动前需要完成对变量替换，在容器环境中启动需要在deployment/statefulset中提前注入相关环境变量，
在启动过程中使用`envsubst`命令实现替换。使用样例如下：

```shell
#!/bin/bash
module="bcs-egress-controller"

cd /data/bcs/${module}
chmod +x ${module}

#check configuration render
if [ $BCS_CONFIG_TYPE == "render" ] then
  cat ${module}.json.template | envsubst | tee ${module}.json
fi

#ready to start
/data/bcs/${module}/${module} $@
```

建议各模块不需要单独设置日志目录，直接将日志打印至容器输出流即可。
