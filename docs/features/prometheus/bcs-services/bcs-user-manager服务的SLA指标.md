# bcs-user-manager SLA服务指标

指标详情可参考prom指标定义文件：

bk-bcs/bcs-services/bcs-user-manager/app/metrics/metrics.go

## 指标聚合

### 权限和token管理成功率

```
# 权限和token管理成功率
sum(bkbcs_usermanager_api_request_latency_time_count{handler=~"GrantPermission|GetPermission|RevokePermission|VerifyPermission|VerifyPermissionV2|GetRegisterToken|CreateToken| GetToken|DeleteToken|UpdateToken|CreateTempToken|CreateClientToken|RefreshPlainToken|RefreshSaasToken",status="success"})/sum(bkbcs_usermanager_api_request_latency_time_count{handler=~"GrantPermission|GetPermission|RevokePermission|VerifyPermission|VerifyPermissionV2|GetRegisterToken|CreateToken|GetToken|DeleteToken|UpdateToken|CreateTempToken|CreateClientToken|RefreshPlainToken|RefreshSaasToken"})

# 权限管理成功率
sum(bkbcs_usermanager_api_request_latency_time_count{handler=~"GrantPermission|GetPermission|RevokePermission|VerifyPermission|VerifyPermissionV2",status="success"})/sum(bkbcs_usermanager_api_request_latency_time_count{handler=~"GrantPermission|GetPermission|RevokePermission|VerifyPermission|VerifyPermissionV2"})

# token管理成功率
sum(bkbcs_usermanager_api_request_latency_time_count{handler=~"GetRegisterToken|CreateToken|GetToken|DeleteToken|UpdateToken|CreateTempToken|CreateClientToken|RefreshPlainToken|RefreshSaasToken",status="success"})/sum(bkbcs_usermanager_api_request_latency_time_count{handler=~"GetRegisterToken|CreateToken|GetToken|DeleteToken|UpdateToken|CreateTempToken|CreateClientToken|RefreshPlainToken|RefreshSaasToken"})
```

### 权限和token管理延迟<256ms

```
# 权限和token管理延迟<256ms
sum(bkbcs_usermanager_api_request_latency_time_bucket{handler=~"GrantPermission|GetPermission|RevokePermission|VerifyPermission|VerifyPermissionV2|GetRegisterToken|CreateToken| GetToken|DeleteToken|UpdateToken|CreateTempToken|CreateClientToken|RefreshPlainToken|RefreshSaasToken",le="0.256",status="success"})/sum(bkbcs_usermanager_api_request_latency_time_count{handler=~"GrantPermission|GetPermission|RevokePermission|VerifyPermission|VerifyPermissionV2|GetRegisterToken|CreateToken|GetToken|DeleteToken|UpdateToken|CreateTempToken|CreateClientToken|RefreshPlainToken|RefreshSaasToken",status="success"})

# 权限管理延迟<256ms
sum(bkbcs_usermanager_api_request_latency_time_bucket{handler=~"GrantPermission|GetPermission|RevokePermission|VerifyPermission|VerifyPermissionV2",le="0.256",status="success"})/sum(bkbcs_usermanager_api_request_latency_time_count{handler=~"GrantPermission|GetPermission|RevokePermission|VerifyPermission|VerifyPermissionV2",status="success"})

# token管理延迟<256ms
sum(bkbcs_usermanager_api_request_latency_time_bucket{handler=~"GetRegisterToken|CreateToken|GetToken|DeleteToken|UpdateToken|CreateTempToken|CreateClientToken|RefreshPlainToken|RefreshSaasToken",le="0.256",status="success"})/sum(bkbcs_usermanager_api_request_latency_time_count{handler=~"GetRegisterToken|CreateToken|GetToken|DeleteToken|UpdateToken|CreateTempToken|CreateClientToken|RefreshPlainToken|RefreshSaasToken",status="success"})
```

