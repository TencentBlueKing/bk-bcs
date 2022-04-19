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
import datetime
import logging
from dataclasses import dataclass
from typing import Dict, List, Tuple

from backend.resources.hpa.utils import HPAMetricsParser
from backend.resources.utils.format import ResourceDefaultFormatter
from backend.templatesets.legacy_apps.instance import constants as instance_constants
from backend.uniapps.application import constants as application_constants
from backend.utils import basic
from backend.utils.basic import get_with_placeholder, getitems

logger = logging.getLogger(__name__)


def get_metric_name_value(metric: Dict, field: str) -> Tuple:
    """获取metrics名称, 值"""
    metric_type = metric["type"]
    # CPU/MEM 等系统类型
    if metric_type == "Resource":
        name = metric["resource"]["name"]
        metric_value = metric["resource"][field]["averageUtilization"]
    else:
        # Pod等自定义类型
        name = metric[metric_type.lower()]["metric"]["name"]
        metric_value = metric[metric_type.lower()][field].get("averageValue")
    return name, metric_value


def get_current_metrics(resource_dict: Dict) -> Dict:
    """获取当前监控值"""
    current_metrics = {}
    for metric in resource_dict["spec"].get("metrics") or []:
        # 跳过 type 为空的场景
        if not metric.get("type"):
            continue
        name, target = get_metric_name_value(metric, field="target")
        current_metrics[name] = {"target": target, "current": None}

    for metric in resource_dict["status"].get("currentMetrics") or []:
        # 跳过 type 为空的场景
        if not metric.get("type"):
            continue
        name, current = get_metric_name_value(metric, field="current")
        current_metrics[name]["current"] = current

    return current_metrics


def get_current_metrics_display(_current_metrics: Dict) -> str:
    """当前监控值前端显示"""
    current_metrics = []

    for name, value in _current_metrics.items():
        value["name"] = name.upper()
        # None 现在为-
        if value["current"] is None:
            value["current"] = "-"
        current_metrics.append(value)
    # 按CPU, Memory显示
    current_metrics = sorted(current_metrics, key=lambda x: x["name"])
    display = ", ".join(f'{metric["name"]}({metric["current"]}/{metric["target"]})' for metric in current_metrics)

    return display


@dataclass
class Condition:
    """k8s hpa condition 结构"""

    type: str
    status: str
    lastTransitionTime: datetime.datetime
    reason: str
    message: str


def sort_by_normalize_transition_time(conditions: List[Condition]) -> List[Condition]:
    """规整 lastTransitionTime 并排序"""

    def normalize_condition(condition: Condition):
        """格式时间 lambda 函数"""
        condition["lastTransitionTime"] = basic.normalize_time(condition["lastTransitionTime"])
        return condition

    # lastTransitionTime 转换为本地时间
    conditions = map(normalize_condition, conditions)

    # 按时间倒序排序
    conditions = sorted(conditions, key=lambda x: x["lastTransitionTime"], reverse=True)

    return conditions


class HPAFormatter(ResourceDefaultFormatter):
    def __init__(self, cluster_id: str, project_code: str):
        self.cluster_id = cluster_id
        self.project_code = project_code

    def format_dict(self, resource_dict: Dict) -> Dict:
        labels = resource_dict.get("metadata", {}).get("labels") or {}
        # 获取模板集信息
        template_id = labels.get(instance_constants.LABLE_TEMPLATE_ID)
        # 资源来源
        source_type = labels.get(instance_constants.SOURCE_TYPE_LABEL_KEY)
        if not source_type:
            source_type = "template" if template_id else "other"

        annotations = resource_dict.get("metadata", {}).get("annotations") or {}
        namespace = resource_dict["metadata"]["namespace"]

        current_metrics = get_current_metrics(resource_dict)

        # k8s 注意需要调用 autoscaling/v2beta2 版本 api
        conditions = resource_dict["status"].get("conditions", [])
        conditions = sort_by_normalize_transition_time(conditions)

        data = {
            "cluster_id": self.cluster_id,
            "name": resource_dict["metadata"]["name"],
            "namespace": namespace,
            "max_replicas": resource_dict["spec"]["maxReplicas"],
            "min_replicas": resource_dict["spec"]["minReplicas"],
            "current_replicas": resource_dict["status"]["currentReplicas"],
            "current_metrics_display": get_current_metrics_display(current_metrics),
            "current_metrics": current_metrics,
            "conditions": conditions,
            "source_type": application_constants.SOURCE_TYPE_MAP.get(source_type),
            "creator": annotations.get(instance_constants.ANNOTATIONS_CREATOR, ""),
            "create_time": annotations.get(instance_constants.ANNOTATIONS_CREATE_TIME, ""),
            "ref_name": getitems(resource_dict, "spec.scaleTargetRef.name", ""),
            "ref_kind": getitems(resource_dict, "spec.scaleTargetRef.kind", ""),
        }

        data["update_time"] = annotations.get(instance_constants.ANNOTATIONS_UPDATE_TIME, data["create_time"])
        data["updator"] = annotations.get(instance_constants.ANNOTATIONS_UPDATOR, data["creator"])
        return data


class HPAFormatter4Dashboard(ResourceDefaultFormatter):
    """ HPA 格式化（资源视图用）"""

    def format_dict(self, resource_dict: Dict) -> Dict:
        res = self.format_common_dict(resource_dict)
        ref = resource_dict['spec']['scaleTargetRef']
        res.update(
            {
                'reference': f"{ref['kind']}/{ref['name']}",
                'targets': HPAMetricsParser(resource_dict).parse(),
                'min_pods': get_with_placeholder(resource_dict, 'spec.minReplicas', '<unset>'),
                'max_pods': getitems(resource_dict, 'spec.maxReplicas'),
                'replicas': get_with_placeholder(resource_dict, 'status.currentReplicas'),
            }
        )
        return res
