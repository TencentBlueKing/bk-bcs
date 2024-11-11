### 描述

根据查询条件过滤Cluster列表，如果某项条件值为空，则忽略此条件。如果所有条件都为空，则默认查询所有集群

### 调用示例

```bash
curl -X GET \
-H 'X-Bkapi-Authorization: {"bk_app_code": "bk_apigw_test", "bk_app_secret": "***"}' \
'http://bcs-api.bkdomain/bcsapi/v4/clustermanager/v1/cluster?projectID={projectID}'
```

### 请求参数

| 名称                  | 位置    | 类型             | 必选 | 说明                                                                                  |
|---------------------|-------|----------------|----|-------------------------------------------------------------------------------------|
| clusterName         | query | string         | 否  | clusterName. 集群名称                                                                   |
| provider            | query | string         | 否  | provider. 集群提供者，云模板ID，影响具体云设施管理流程                                                   |
| region              | query | string         | 否  | region. 集群所在地域                                                                      |
| vpcID               | query | string         | 否  | vpcID. 集群私有网络id，云上集群填写                                                              |
| projectID           | query | string         | 否  | projectID. 项目ID                                                                     |
| businessID          | query | string         | 否  | businessID. CMDB业务ID                                                                |
| environment         | query | string         | 否  | environment. 集群环境, 例如[prod, debug, test]                                            |
| engineType          | query | string         | 否  | engineType. 引擎类型，[k8s, mesos]，默认k8s                                                 |
| isExclusive         | query | boolean        | 否  | isExclusive. 是否为业务独占集群                                                              |
| clusterType         | query | string         | 否  | clusterType. 集群类型, 例如[federation, single], federation表示为联邦集群，single表示独立集群，默认为single |
| federationClusterID | query | string         | 否  | federationClusterID. 如果该集群为联邦集群的一部分，用该ID记录联邦Host信息                                  |
| status              | query | string         | 否  | status. 集群状态，可能状态CREATING，RUNNING，DELETING，FALURE，INITIALIZATION，DELETED            |
| offset              | query | integer(int64) | 否  | offset. 查询偏移量                                                                       |
| limit               | query | integer(int64) | 否  | limit. 查询限制数量                                                                       |
| operator            | query | string         | 否  | operator. 操作者(传参时获取该用户对应集群的权限信息,不传时仅获取集群列表信息)                                       |
| systemID            | query | string         | 否  | systemID. cloudID过滤集群                                                               |
| extraClusterID      | query | string         | 否  | extraClusterID. 导入集群的额外集群ID标识信息,默认时空值                                               |
| isCommonCluster     | query | boolean        | 否  | isCommonCluster. 是否为公共集群,默认false                                                    |
| clusterID           | query | string         | 否  | clusterID. 集群ID                                                                     |
| all                 | query | boolean        | 否  | all. true时获取所有的集群信息,包括被软删除的集群; false时获取非DELETED状态的集群信息                              |

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
  "clusterExtraInfo": {
    "property1": {
      "canDeleted": true,
      "providerType": "string",
      "autoScale": true
    },
    "property2": {
      "canDeleted": true,
      "providerType": "string",
      "autoScale": true
    }
  },
  "web_annotations": {
    "perms": {
      "property1": {},
      "property2": {}
    }
  }
}
```