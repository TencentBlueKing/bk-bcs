# -*- coding: utf-8 -*-
"""
Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community
Edition) available.
Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://opensource.org/licenses/MIT

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.
"""
from typing import Dict, List, Tuple

import attr

from backend.components import sops
from backend.container_service.infras.hosts.constants import APPLY_HOST_TEMPLATE_ID, SOPS_BIZ_ID


@attr.dataclass
class HostData:
    region: str
    vpc_name: str
    cvm_type: str
    disk_type: str
    disk_size: int
    replicas: int
    zone_id: str

    @classmethod
    def from_dict(cls, init_data: Dict) -> "HostData":
        fields = [f.name for f in attr.fields(cls)]
        return cls(**{k: v for k, v in init_data.items() if k in fields})


def create_and_start_host_application(cc_app_id: str, username: str, host_data: HostData) -> Tuple[int, str]:
    """创建并启动申请主机任务流程"""
    client = sops.SopsClient()
    # 组装创建任务参数
    task_name = f"[{cc_app_id}]apply host resource"
    data = sops.CreateTaskParams(
        name=task_name,
        constants={
            "${appID}": cc_app_id,
            "${user}": username,
            "${qcloudRegionId}": host_data.region,
            "${cvm_type}": host_data.cvm_type,
            "${diskSize}": host_data.disk_size,
            "${replicas}": host_data.replicas,
            "${vpc_name}": host_data.vpc_name,
            "${zone_id}": host_data.zone_id,
            "${disk_type}": host_data.disk_type,
        },
    )
    # 创建任务
    resp_data = client.create_task(bk_biz_id=SOPS_BIZ_ID, template_id=APPLY_HOST_TEMPLATE_ID, data=data)
    task_id = resp_data["task_id"]
    task_url = resp_data["task_url"]

    # 启动任务
    client.start_task(bk_biz_id=SOPS_BIZ_ID, task_id=task_id)

    return task_id, task_url


def get_task_state_and_steps(task_id: str) -> Dict:
    """获取任务总状态及步骤状态"""
    client = sops.SopsClient()
    resp_data = client.get_task_status(bk_biz_id=SOPS_BIZ_ID, task_id=task_id)

    # NOTE: 现阶段不处理SUSPENDED(暂停)状态，当任务处于RUNNING状态, 认为任务处于执行中
    steps = {}
    for step_id, detail in (resp_data.get("children") or {}).items():
        name = detail.get("name") or ""
        # NOTE: 过滤掉sops中的开始和结束节点(两个空标识节点)
        if "EmptyEndEvent" in name or "EmptyStartEvent" in name:
            continue
        steps[name] = {"state": detail["state"], "step_id": step_id}

    # 返回任务状态, 步骤名称及状态
    return {"state": resp_data["state"], "steps": steps}


def get_applied_ip_list(task_id: str, step_id: str) -> List[str]:
    """获取申领的机器列表"""
    client = sops.SopsClient()
    resp_data = client.get_task_node_data(bk_biz_id=SOPS_BIZ_ID, task_id=task_id, node_id=step_id)

    # 获取返回的IP列表
    # outputs 结构: [{key: xxx, value: str}, {key: log_outputs, value: {ip_list: "127.0.0.1,127.0.0.2"}}]
    outputs = resp_data["outputs"]

    ips = ""
    for i in outputs:
        if i["key"] != "log_outputs":
            continue
        # NOTE: 接口返回中是以英文逗号分隔的字符串
        ips = i["value"]["ip_list"]

    return ips.split(",")
