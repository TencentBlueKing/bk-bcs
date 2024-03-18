# bcs-ops X bk-sops

bcs-ops 借助 bk-sops 的编排能力，实现原生集群的流程化创建、维护等功能。

## 镜像

### 镜像编译

```bash
✦ ➜  cd bk-bcs/bcs-ops
✦ ➜  make build_image
# check
✦ ➜  docker run --rm --entrypoint /bin/ls bcs-ops-upload:test
bcs-ops.tar.gz # 脚本包
bcs_bk_sops_common.dat # bcs-ops x 标准运维 公共流程模板
upload.py # 导入脚本
```

### 镜像使用说明

注意[导入事项](#导入事项)
env 文件配置见[envfile 示例](#envfile%20示例)

```bash
# 上传脚本包至 bkrepo, env 文件见下文
✦ ➜  docker run --env-file env bcs-ops-upload:test upload bkrepo
DEBUG:root:bkrepo_url: http://bkrepo.example.com/generic/blueking/bcs-ops/scripts/bcs-ops.tar.gz
DEBUG:urllib3.connectionpool:Starting new HTTP connection (1): bkrepo.example.com:80
DEBUG:urllib3.connectionpool:http://bkrepo.example.com:80 "PUT /generic/blueking/bcs-ops/scripts/bcs-ops.tar.gz HTTP/1.1" 200 1386
INFO:root:Upload bcs-ops.tar.gz to http://bkrepo.example.com/generic/blueking/bcs-ops/scripts/bcs-ops.tar.gz succeeded.

# 上传标准运维流程至 sops
✦ ➜ docker run --env-file env bcs-ops-upload:test upload sops
DEBUG:root:bkrepo_url: http://bkrepo.example.com/generic/blueking/bcs-ops/scripts/bcs-ops.tar.gz
DEBUG:root:sub_pat: SCRIPT_URL_PLACEHOLDER, sub_str: http://bkrepo.example.com/generic/blueking/bcs-ops/scripts/bcs-ops.tar.gz
DEBUG:urllib3.connectionpool:Starting new HTTP connection (1): bkapi.example.com:80
DEBUG:urllib3.connectionpool:http://bkapi.example.com:80 "POST /api/c/compapi/v2/sops/import_common_template/ HTTP/1.1" 200 536
INFO:root:Upload succeeded: bcs_bk_sops_common.dat
```

#### envfile 示例

容器运行需要配置如下环境变量至`env`文件中，在容器运行时使用`--env-file`指定

```plaintext
BKAPI_HOST=bkapi.example.com // 环境的bkapi host
APP_CODE=bk_sops        // 确认 bk_sops 免登陆验证的 app_code
APP_SECRET=your_secret // 和上面对应 secret

REPO_HOST=bkrepo.example.com //环境的 bkrepo host
REPO_PROJECT=blueking        // bkrepo的项目，默认 blueking
REPO_BUCKET=bcs-ops        // bkrepo的bucket，需要在bkrepo上提前创建公开的bucket
REPO_PATH=scripts        // bkrepo 的路径
REPO_USER=xxxx        // bkrepo的用户，联系管理员获取，需要拥有上面 bucket 的权限！
REPO_PASSWD=xxxx // bkrepo的密码，联系管理员获取

LOG_LEVEL=DEBUG // 日志等级
```

## 标准运维

### bcs 公共流程模板制作说明

#### 模板编写规范

1. 流程中的文件路径，必须暴露为可修改的变量。
2. 文件模板名为【BCS】xxxx
3. 新增流程，其 id 按[已有的公共流程模板](#公共流程模板使用说明)进行序列自增（不能用重复），并在文档中说明用法。

   > 默认情况下，标准运维公共流程只能从 10000 后创建。bcs 按照约定，固定使用预留给 bcs 使用模板 id（从 10 开始）。页面上手动新增的流程必然是 10000+的，请按如下操作修改其 id 至预留范围内。

   ```sql
   -- 进入 sops 的数据库，若已知，则无需考虑
   SHOW databases LIKE '%sops%';
   USE xxxx; -- 使用上面查找到的数据库

   -- 改变 id 的操作，生产环境建议先备份数据库
   UPDATE template_commontemplate SET id = new_id WHERE id=old_id;
   ```

#### 模板导出（**必看**）

1. 必须使用 dat 格式导出（yaml 格式无法调用接口）
2. 脚本包路径处理：对于[id10 脚本包分发流程](#id10)，导出时会包含该环境的存放在`bkrepo`的脚本包下载地址，这个下载地址对于各个蓝鲸环境而是不同的，需要将这个路径替换为`SCRIPT_URL_PLACEHOLDER`。

   ```bash
   url="http://bkrepo.example.com/generic/blueking/bcs-ops/scripts/bcs-ops.tar.gz"
   SOPS_FILE=bk_sops_common_xxxx.dat python3 upload.py modify $url
   ```

#### 导入注意事项（**必看**）<a id="导入事项"></a>

1. 必须始终用<u>覆盖相同 id </u>方式导入！
2. 导入的环境中可能存在 id 不同，但 template_id 相同模板（比如一个环境同时以覆盖和新增的方式导入了两次）。执行导入动作后，若出现标准运维页面显示"系统出现异常"，可按照下面的步骤清理脏数据进行修复

   ```sql
   -- 进入 sops 的数据库，若已知，则无需考虑
   SHOW databases LIKE '%sops%';
   USE xxxx; -- 使用上面查找到的数据库

   SELECT  pipeline_template_id , count(*) AS cnt FROM template_commontemplate GROUP BY pipeline_template_id HAVING cnt > 1; -- 查看是否存在重复的 template_id
   SELECT id FROM template_commontemplate AS a JOIN (SELECT  pipeline_template_id , count(*) AS cnt FROM template_commontemplate GROUP BY pipeline_template_id HAVING cnt > 1) AS b on a.pipeline_template_id = b.pipeline_template_id; -- 查看重复id

   -- 生产环境执行删除动作前，最好先备份数据库/表，除非有十足的把握。
   DELETE FROM template_commontemplate WHERE id in (SELECT id FROM template_commontemplate AS a JOIN (SELECT  pipeline_template_id , count(*) AS cnt FROM template_commontemplate GROUP BY pipeline_template_id HAVING cnt > 1) AS b on a.pipeline_template_id = b.pipeline_template_id WHERE id > 10000); --这里 id > 10000，是因为默认情况下创建的流程 id 必然大于 10000。而覆盖导入的bcs-ops 流程id默认小于 10000。
   ```

### 公共流程模板使用说明

#### id10.【BCS】bcsops distribute <a id="id10"></a>

参数

1. `WORKSPACE`: bcs-ops 脚本包工作路径，默认 `/data/bcs-ops` (注意，这个路径与后续所有流程的`WORKSPACE`变量一致！如果需要修改，请统一修改！)
2. `HOST_IP`: 分发节点
3. `SCRIPT_URL`: 脚本包下载地址，若是通过[镜像上传](#镜像使用说明)的标准运维流程，则路径自动配置为 bkrepo 的下载路径。如果是手动上传的，则需要自己修改这个路径。(依赖这个流程的模板也要对应的更新)

功能描述

分发存储在 `bkrepo` 上的脚本包至节点机器 `HOST_IP`

#### id11. 【BCS】Setup Kubernetes on Linux <a id="id11"></a>

参数

1. `base_env`: 基础环境变量设置，见[readme.md#环境变量](../readme.md#环境变量)
2. `bcs_env`: bcs 传递环境变量，通过`;`分隔为`bcs_sops_bcs_env`。（手动执行可忽略）
3. `extra_env`: bcs 传递的额外环境变量，通过`;`分隔为`bcs_sops_extra_env`。（手动执行可忽略）
4. `ctrl_ip_list`: 控制平面节点列表，建议为奇数个，默认选取第一个 ip 作为创世节点(`ctrl_ip`)
5. `node_ip_list`: 工作平面节点列表，若仅创建集群，可以不填
6. `workspace`: 工作目录为`/data/bcs-ops`
7. `cluster_id`: bcs 传递的 cluster_id，可不填。
8. `VIP`：默认为 `192.168.1.1`，如果与环境有冲突，建议修改

```bash
# 环境变量加载顺序如下，如果需要覆盖，可以通过extra_env覆盖
${base_env}
${bcs_sops_bcs_env}
VIP=${VIP}
${bcs_sops_extra_env}
```

依赖

- [id10](#id10)
- [id12](#id12)
- [id15](#id15)

功能描述

将`ctrl_ip_list`中的第一个节点作为创世节点，并添加`node_ip_list`中的节点作为工作平面节点。\
如果是手动执行，则需要跳过最后一步：[`安装 bcs-kube-agent`](#id15)。

#### id12. 【BCS】Add Kubernetes Worker <a id="id12"></a>

参数

1. `base_env`: 基础环境变量设置，见[readme.md#环境变量](../readme.md#环境变量)
2. `bcs_env`: bcs 传递环境变量，通过`;`分隔为`bcs_sops_bcs_env`。（手动执行可忽略）
3. `extra_env`: bcs 传递的额外环境变量，通过`;`分隔为`bcs_sops_extra_env`。（手动执行可忽略）
4. `ctrl_ip_list`: 控制平面节点列表，建议为奇数个，默认选取第一个 ip 获取集群环境变量
5. `node_ip_list`: 工作平面节点列表，若仅创建集群，可以不填
6. `workspace`: 工作目录为`/data/bcs-ops`

依赖

- [id10](#id10)

功能描述

添加集群工作节点`node_ip_list`至`ctrl_ip_list`所在的集群，默认读取控制平面列表第一个节点的集群环境变量。因为[创建集群](#id11)环境变量已经配置，会自动传递给节点机器，因此这里的环境变量参数**不用配置**。

#### id13. 【BCS】Remove Kubernetes Worker

参数

1. `ctrl_ip_list`: 控制平面节点列表，默认选取第一个 ip 执行`kubectl delte nodes`操作
2. `node_ip_list`: 待移除工作平面节点列表，
3. `workspace`: 工作目录为`/data/bcs-ops`

依赖

- [id10](#id10)

功能描述

移除集群中`node_ip_list`的工作节点，并移除被移除的节点机器的容器环境和配置。

#### id14. 【BCS】Destroy Cluster

参数

1. `ctrl_ip_list`: 节点列表
2. `workspace`: 工作目录为`/data/bcs-ops`

功能描述

销毁节点列表的集群，**也可以用这个流程重置任意节点的容器环境**。

#### id15. 【BCS】安装 bcs-kube-agent <a id="id15"></a>

参数

1. cluster_id: bcs 集群 id，由 bcs 传入
2. bcs_env: bcs 环境变量，由 bcs 传入。
3. extra_env: 额外环境变量，由 bcs 传入。
4. sops_bcs_ns: bcs-kube-agent 安装的命名空间，默认为 `bcs-system`
5. sops_chart_version: bcs helm chart 版本，默认为 `1.27.0`
6. sops_chartrepo_url: bcs helm repo url，默认为 `https://hub.bktencent.com/chartrepo/blueking`

功能描述

bcs 通过安装 bcs-kube-agent 的方式纳管创建的原生集群，相关参数由 bcs 传入。

#### id16. 【BCS】K8S master replace

参数

1. master_ip 一个当前存在于集群的 master，且不是本次被替换的 master 的 ip
2. new_master_ip 本次将被替换进集群的 master 的 ip
3. unwanted_master_ip 本次将被替换出集群的 master 的 ip
4. unwanted_master_name 本次将被替换进集群的 master 的节点名
5. workspace 节点上的 bcs-ops 目录

功能描述

1. 扩容 new_master_ip 指定的 master 节点

2. 清理掉 unwanted_master_ip 指定的 master 节点上的 k8s 环境以及 unwanted_master_name 对应的 k8s 节点以及 etcd 节点

#### id17. 【BCS】etcd backup

参数

1. host_ip_list 需要进行备份的 etcd 节点 ip，多个使用,隔开
2. cacert 访问 etcd 的 ca 证书文件路径
3. cert 访问 etcd 的证书文件路径
4. key 访问 etcd 的 key 文件路径
5. backup_file 备份文件路径
6. workspace 节点上的 bcs-ops 目录

功能描述

1. 在各个 etcd 节点上，通过本机的 endpoint 获取 snapshot 到 backup_file 指定目录

#### id18. 【BCS】etcd restore

参数

1. host_ip_list 需要进行备份的 etcd 节点 ip，多个使用,隔开
2. source_host 备份文件来源机器
3. source_file 备份文件路径
4. data_dir etcd 数据目录
5. clusterinfo_file 集群信息文件路径
6. workspace 节点上的 bcs-ops 目录

功能描述

1. 将 source_file 备份文件从 source_host 传到各台 etcd 节点机器上后，根据 clusterinfo_file 中的信息将数据恢复到 data_dir 指定的目录

#### id19. etcd new (missing)

参数

1. host_ip_list 新集群的 etcd 节点 ip，多个使用,隔开
2. name etcd 集群名
3. data_dir 数据目录
4. peer_port etcd 节点 peer port
5. service_port etcd 节点 service port
6. metric_port etcd 节点 metric port
7. initial_cluster 此次恢复的 etcd 集群所有成员信息
8. cacert 访问 etcd 的 ca 证书文件路径
9. cert 访问 etcd 的证书文件路径
10. key 访问 etcd 的 key 文件路径
11. workspace 节点上的 bcs-ops 目录

功能描述

1. 根据参数基于原本 kubeadm 创建出来的 etcd.yaml 文件进行替换，并用静态 pod 的方式拉起新集群的所有节点
