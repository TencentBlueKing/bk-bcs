### 描述

查询公共集群及公共集群所属权限

### 调用示例

```bash
curl -X GET \
-H 'X-Bkapi-Authorization: {"bk_app_code": "bk_apigw_test", "bk_app_secret": "***"}' \
'http://bcs-api.bkdomain/bcsapi/clustermanager/v1/sharedclusters?showVCluster=true'
```

### 请求参数

| 名称           | 位置    | 类型      | 必选 | 说明                                          |
|--------------|-------|---------|----|---------------------------------------------|
| showVCluster | query | boolean | 否  | showVCluster. 展示vcluster集群的host共享集群(默认全部展示) |

### 响应示例

```json
{
  "code": 0,
  "message": "string",
  "result": true,
  "data": [
    {
      "clusterID": "string",
      "clusterName": "string",
      "federationClusterID": "string",
      "provider": "string",
      "region": "string",
      "vpcID": "string",
      "projectID": "string",
      "businessID": "string",
      "environment": "string",
      "engineType": "string",
      "isExclusive": true,
      "clusterType": "string",
      "labels": {
        "property1": "string",
        "property2": "string"
      },
      "creator": "string",
      "createTime": "string",
      "updateTime": "string",
      "bcsAddons": {
        "property1": {
          "system": "string",
          "link": "string",
          "params": {
            "property1": "string",
            "property2": "string"
          },
          "allowSkipWhenFailed": true
        },
        "property2": {
          "system": "string",
          "link": "string",
          "params": {
            "property1": "string",
            "property2": "string"
          },
          "allowSkipWhenFailed": true
        }
      },
      "extraAddons": {
        "property1": {
          "system": "string",
          "link": "string",
          "params": {
            "property1": "string",
            "property2": "string"
          },
          "allowSkipWhenFailed": true
        },
        "property2": {
          "system": "string",
          "link": "string",
          "params": {
            "property1": "string",
            "property2": "string"
          },
          "allowSkipWhenFailed": true
        }
      },
      "systemID": "string",
      "manageType": "string",
      "master": {
        "property1": {
          "nodeID": "string",
          "innerIP": "string",
          "instanceType": "string",
          "CPU": 0,
          "mem": 0,
          "GPU": 0,
          "status": "string",
          "zoneID": "string",
          "nodeGroupID": "string",
          "clusterID": "string",
          "VPC": "string",
          "region": "string",
          "passwd": "string",
          "zone": 0,
          "deviceID": "string",
          "nodeTemplateID": "string",
          "nodeType": "string",
          "nodeName": "string",
          "innerIPv6": "string",
          "zoneName": "string",
          "taskID": "string",
          "failedReason": "string",
          "chargeType": "string"
        },
        "property2": {
          "nodeID": "string",
          "innerIP": "string",
          "instanceType": "string",
          "CPU": 0,
          "mem": 0,
          "GPU": 0,
          "status": "string",
          "zoneID": "string",
          "nodeGroupID": "string",
          "clusterID": "string",
          "VPC": "string",
          "region": "string",
          "passwd": "string",
          "zone": 0,
          "deviceID": "string",
          "nodeTemplateID": "string",
          "nodeType": "string",
          "nodeName": "string",
          "innerIPv6": "string",
          "zoneName": "string",
          "taskID": "string",
          "failedReason": "string",
          "chargeType": "string"
        }
      },
      "networkSettings": {
        "clusterIPv4CIDR": "string",
        "serviceIPv4CIDR": "string",
        "maxNodePodNum": 0,
        "maxServiceNum": 0,
        "enableVPCCni": true,
        "eniSubnetIDs": [
          "string"
        ],
        "subnetSource": {
          "new": [
            null
          ],
          "existed": {}
        },
        "isStaticIpMode": true,
        "claimExpiredSeconds": 0,
        "multiClusterCIDR": [
          "string"
        ],
        "cidrStep": 0,
        "clusterIpType": "string",
        "clusterIPv6CIDR": "string",
        "serviceIPv6CIDR": "string",
        "status": "string",
        "networkMode": "string"
      },
      "clusterBasicSettings": {
        "OS": "string",
        "version": "string",
        "clusterTags": {
          "property1": "string",
          "property2": "string"
        },
        "versionName": "string",
        "subnetID": "string",
        "clusterLevel": "string",
        "isAutoUpgradeClusterLevel": true,
        "area": {
          "bkCloudID": 0,
          "bkCloudName": "string"
        },
        "module": {
          "masterModuleID": "string",
          "masterModuleName": "string",
          "workerModuleID": "string",
          "workerModuleName": "string"
        },
        "upgradePolicy": {
          "supportType": "string"
        }
      },
      "clusterAdvanceSettings": {
        "IPVS": true,
        "containerRuntime": "string",
        "runtimeVersion": "string",
        "extraArgs": {
          "property1": "string",
          "property2": "string"
        },
        "networkType": "string",
        "deletionProtection": true,
        "auditEnabled": true,
        "enableHa": true,
        "clusterConnectSetting": {
          "isExtranet": true,
          "subnetId": "string",
          "domain": "string",
          "securityGroup": "string",
          "internet": {}
        }
      },
      "nodeSettings": {
        "dockerGraphPath": "string",
        "mountTarget": "string",
        "unSchedulable": 0,
        "labels": {
          "property1": "string",
          "property2": "string"
        },
        "extraArgs": {
          "property1": "string",
          "property2": "string"
        },
        "taints": [
          {
            "key": null,
            "value": null,
            "effect": null
          }
        ],
        "masterLogin": {
          "initLoginUsername": "string",
          "initLoginPassword": "string",
          "keyPair": {}
        },
        "workerLogin": {
          "initLoginUsername": "string",
          "initLoginPassword": "string",
          "keyPair": {}
        },
        "masterSecurityGroups": [
          "string"
        ],
        "workerSecurityGroups": [
          "string"
        ]
      },
      "status": "string",
      "updater": "string",
      "networkType": "string",
      "autoGenerateMasterNodes": true,
      "template": [
        {
          "region": "string",
          "zone": "string",
          "vpcID": "string",
          "subnetID": "string",
          "applyNum": 0,
          "CPU": 0,
          "Mem": 0,
          "GPU": 0,
          "instanceType": "string",
          "instanceChargeType": "string",
          "systemDisk": {
            "diskType": null,
            "diskSize": null
          },
          "dataDisks": [
            {}
          ],
          "imageInfo": {
            "imageID": null,
            "imageName": null,
            "imageType": null
          },
          "initLoginPassword": "string",
          "securityGroupIDs": [
            "string"
          ],
          "isSecurityService": true,
          "isMonitorService": true,
          "cloudDataDisks": [
            {}
          ],
          "initLoginUsername": "string",
          "keyPair": {
            "keyID": null,
            "keySecret": null,
            "keyPublic": null
          },
          "dockerGraphPath": "string",
          "nodeRole": "string",
          "charge": {
            "period": null,
            "renewFlag": null
          },
          "internetAccess": {
            "internetChargeType": null,
            "internetMaxBandwidth": null,
            "publicIPAssigned": null,
            "bandwidthPackageId": null,
            "publicIP": null,
            "publicAccessCidrs": null
          }
        }
      ],
      "extraInfo": {
        "property1": "string",
        "property2": "string"
      },
      "moduleID": "string",
      "extraClusterID": "string",
      "isCommonCluster": true,
      "description": "string",
      "clusterCategory": "string",
      "is_shared": true,
      "kubeConfig": "string",
      "importCategory": "string",
      "cloudAccountID": "string",
      "message": "string",
      "isMixed": true,
      "clusterIamRole": "string",
      "sharedRanges": {
        "bizs": [
          "string"
        ],
        "projectIdOrCodes": [
          "string"
        ]
      }
    }
  ],
  "web_annotations": {
    "perms": {
      "property1": {},
      "property2": {}
    }
  }
}
```