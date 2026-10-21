# 云凭证管理

蓝鲸容器管理平台（BCS）支持新建或导入多家主流公有云厂商（腾讯云 TKE、谷歌云 GKE、亚马逊云 EKS、微软云 AKS、华为云 CCE）的容器集群。用户可在【项目管理】→【云凭证】中集中托管各云厂商的访问凭证，供后续创建或导入公有云集群时复用。

---

## 一、新建云凭证

### 1. 前提条件

- 已登录蓝鲸平台，并拥有当前 BCS 项目的**项目查看**及**云账号创建**权限。
- 已在对应公有云控制台准备好具备 Kubernetes 集群管理权限的 API 密钥或服务账号文件。

### 2. 配置字段说明

进入【项目管理】→【云凭证】，选择对应云服务商标签页，点击【新建凭证】：

| 云厂商 | 凭证名称 | 凭证内容字段 | 描述 |
| :--- | :--- | :--- | :--- |
| **腾讯云（Tencent Cloud）** | 自定义易识别名称 | 输入腾讯云 API 密钥的 `SecretID` 和 `SecretKey` | 填写凭证用途、归属账号说明 |
| **谷歌云（Google Cloud）** | 自定义易识别名称 | 上传 Google Cloud 服务账号生成的 JSON 格式密钥文件（`gke-sa.json`） | 填写凭证用途、关联 GCP 项目说明 |
| **亚马逊云 / 微软云 / 华为云** | 自定义易识别名称 | 按页面提示填入对应云厂商的 AccessKey/SecretKey 或订阅服务主体凭据 | 填写凭证用途说明 |

---

## 二、管理云凭证

- **权限要求**：需具备对应项目的**云账号管理**权限。
- **安全脱敏与展示**：
  - **腾讯云凭证列表**：展示系统 ID、名称、描述、`SecretID`（完整展示）、`SecretKey`（脱敏展示，仅显示前后 2 个字符）、关联集群列表及创建时间。
  - **谷歌云凭证列表**：展示 Google 项目 ID、凭证名称、描述（可编辑）、密钥 ID、OAuth2 客户端 ID、关联集群、创建时间及创建人。
- **操作限制**：
  - 为保障凭据安全，已添加的云凭证**不支持查看明文密钥内容，也不支持直接修改密钥**。
  - 若密钥轮换或失效，请新建云凭证；仅当云凭证未关联任何存量集群时，才允许执行【删除】操作。

---

## 三、各云厂商服务账号与权限申请指引

### 1. 腾讯云（TKE）API 密钥申请与授权指引

1. **登录腾讯云控制台**：使用腾讯云账号登录，进入 [访问管理 CAM - API 密钥管理](https://console.cloud.tencent.com/cam/capi)。
2. **创建 API 密钥**：点击【新建密钥】，生成并妥善保存 `SecretID` 和 `SecretKey`。
3. **配置 CAM 权限策略**：
   - 进入 [访问管理 CAM - 用户管理](https://console.cloud.tencent.com/cam)，找到对应子账号，点击【添加策略】。
   - **仅导入已有 TKE 集群**：建议绑定 `QcloudKubernetesAdmin` 或 `QcloudTKEFullAccess` 策略。
   - **通过 BCS 新建 TKE 集群或节点扩容**：除 `QcloudTKEFullAccess` 外，还需绑定 `QCloudFinanceFullAccess`（用于创建 CVM/TKE 资源时的支付扣费校验）以及 CVM/VPC 相关查询权限。
   - **开启节点池自动扩缩容（Cluster Autoscaler）**：云凭证至少需包含以下自定义策略权限：

```json
{
  "version": "2.0",
  "statement": [
    {
      "effect": "allow",
      "action": ["tke:*"],
      "resource": ["*"]
    },
    {
      "effect": "allow",
      "action": [
        "cvm:DescribeSecurityGroups",
        "cvm:DescribeImages",
        "cvm:DescribeInstances"
      ],
      "resource": ["*"]
    },
    {
      "effect": "allow",
      "action": ["vpc:Describe*"],
      "resource": ["*"]
    },
    {
      "effect": "allow",
      "action": ["as:*"],
      "resource": ["*"]
    }
  ]
}
```

### 2. 谷歌云（GKE）服务账号申请指引

1. **登录 Google Cloud 控制台**：访问 [Google Cloud Console - IAM & Admin](https://console.cloud.google.com/iam-admin)。
2. **创建服务账号**：点击【创建服务账号（Create Service Account）】，输入服务账号名称与描述。
3. **授予角色权限**：在权限分配步骤中，为该服务账号授予 `Kubernetes Engine Admin` 角色，确保其具备 GKE 集群与节点池管理权限。
4. **生成并下载 JSON 密钥**：进入新创建的服务账号详情页 →【密钥（Keys）】→【添加密钥】→【创建新密钥】，选择 **JSON** 格式下载到本地，随后在 BCS 云凭证页面上传该 JSON 文件。
