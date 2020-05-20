# bcs daemonset配置说明

bcs daemonset实现与k8s deamonset同等概念

## json配置模板
```json
{
	"apiVersion": "v4",
	"kind": "daemonset",
	"killPolicy": {
		"gracePeriod": 10
	},
	"metadata": {
		"labels": {
			"test_label": "test_label",
		},
		"name": "ri-test-rc-001",
		"namespace": "nfsol"
	},
	"spec": {
		"template": {
			"metadata": {
				"labels": {
					"test_label": "test_label"
				}
			},
			"spec": {
				"containers": [{
					"hostname": "container-hostname",
					"command": "bash",
					"args": [
						"args1",
						"args2"
					],
					"parameters": [{
							"key": "rm",
							"value": "false"
						},
						{
							"key": "ulimit",
							"value": "nproc=8092"
						},
						{
							"key": "ulimit",
							"value": "nofile=65535"
						},
						{
							"key": "ip",
							"value": "127.0.0.1"
						}
					],
					"type": "MESOS",
					"env": [
                        {
                            "name": "test_env",
                            "value": "test_env"
                        },
                        {
                            "name": "namespace",
                            "value": "${bcs.namespace}"
                        },
                        {
                            "name": "http-port",
                            "value": "${bcs.ports.http_port}"
                        }
					],
					"image": "docker.hub.com/nfsol/log:92763",
					"imagePullUser": "userName",
					"imagePullPasswd": "passwd",
					"imagePullPolicy": "Always|IfNotPresent",
					"privileged": false,
					"ports": [{
							"containerPort": 8090,
							"hostPort": 8090,
							"name": "test-tcp",
							"protocol": "TCP"
						},
						{
							"containerPort": 8080,
							"hostPort": 8080,
							"name": "http-port",
							"protocol": "http"
						}
					],
					"healthChecks": [{
						"type": "HTTP|TCP|COMMAND|REMOTE_HTTP|REMOTE_TCP",
						"intervalSeconds": 30,
						"timeoutSeconds": 5,
						"consecutiveFailures": 3,
						"gracePeriodSeconds": 5,
						"http": {
							"port": 8080,
							"portName": "test-http",
							"scheme": "http|https",
							"path": "/check"
						},
						"tcp": {
							"port": 8090,
							"portName": "test-tcp"
						},
						"command": {
                            "value": "ls /"
                        }
					}],
					"resources": {
						"limits": {
							"cpu": "2",
							"memory": "8"
						},
						"requests": {
							"cpu": "2",
							"memory": "8"
						}
					},
					"volumes": [{
						"volume": {
							"hostPath": "/data/host/path",
							"mountPath": "/container/path",
							"readOnly": false
						},
						"name": "test-vol"
					}],
					"secrets": [{
						"secretName": "mySecret",
						"items": [{
								"type": "env",
								"dataKey": "abc",
								"keyOrPath": "SRECT_ENV"
							},
							{
								"type": "file",
								"dataKey": "abc",
								"keyOrPath": "/data/container/path/filename.conf",
								"subPath": "relativedir/",
								"readOnly": false,
								"user": "root"
							}
						]
					}],
					"configmaps": [{
						"name": "template-configmap",
						"items": [{
								"type": "env",
								"dataKey": "config-one",
								"keyOrPath": "SECRET_ENV"
							},
							{
								"type": "file",
								"dataKey": "config_two",
								"dataKeyAlias": "config-two",
								"KeyOrPath": "/data/contianer/path/filename.txt",
								"readOnly": false,
								"user": "root"
							}
						]
					}]
				}],
				"networkMode": "BRIDGE",
				"networkType": "cnm",
				"netLimit": {
					"egressLimit": 100
				}
			}
		}
	}
}
```