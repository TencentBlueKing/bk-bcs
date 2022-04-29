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
from typing import Dict, List

from attr import dataclass

from backend.resources.constants import HPA_METRIC_MAX_DISPLAY_NUM, MetricSourceType
from backend.utils.basic import getitems


@dataclass
class HPAMetricsParser:
    """
    HPA 指标解析器，解析逻辑参考
    kubernetes/kubernetes formatHPAMetrics
    https://github.com/kubernetes/kubernetes/blob/master/pkg/printers/internalversion/printers.go#L2027
    """

    hpa: Dict

    def __attrs_post_init__(self):
        self.specs = getitems(self.hpa, 'spec.metrics') or []
        self.statuses = getitems(self.hpa, 'status.currentMetrics') or []
        self.metrics = []

    def parse(self) -> str:
        """获取 HPA Metrics 信息"""
        if not self.specs:
            return '<none>'

        for idx, spec in enumerate(self.specs):
            # 根据不同的来源类型，选择不用的解析方法
            parse_func = {
                MetricSourceType.External: self._parse_external_metric,
                MetricSourceType.Pods: self._parse_pods_metric,
                MetricSourceType.Object: self._parse_object_metric,
                MetricSourceType.Resource: self._parse_resource_metric,
                MetricSourceType.ContainerResource: self._parse_container_resource_metric,
            }[spec['type']]
            self.metrics.append(parse_func(idx, spec))

        count = len(self.metrics)
        # 如果长度过长，则进行截断处理
        if count > HPA_METRIC_MAX_DISPLAY_NUM:
            self.metrics = self.metrics[:HPA_METRIC_MAX_DISPLAY_NUM]
            return f"{', '.join(self.metrics)} + {count - HPA_METRIC_MAX_DISPLAY_NUM} more..."

        return ', '.join(self.metrics)

    def _parse_external_metric(self, idx: int, spec: Dict) -> str:
        """解析来源自 External 的指标信息"""
        current = '<unknown>'
        if getitems(spec, 'external.target.averageValue') is not None:
            if len(self.statuses) > idx:
                ext_cur_avg_val = getitems(self.statuses[idx], 'external.current.averageValue')
                current = ext_cur_avg_val if ext_cur_avg_val else current
            return f"{current}/{getitems(spec, 'external.target.averageValue')} (avg)"
        else:
            if len(self.statuses) > idx:
                ext_cur_val = getitems(self.statuses[idx], 'external.current.value')
                current = ext_cur_val if ext_cur_val else current
            return f"{current}/{getitems(spec, 'external.target.value')}"

    def _parse_pods_metric(self, idx: int, spec: Dict) -> str:
        """解析来源自 Pods 的指标信息"""
        current = '<unknown>'
        if len(self.statuses) > idx and self.statuses[idx].get('pods') is not None:
            current = getitems(self.statuses[idx], 'pods.current.averageValue')
        return f"{current}/{getitems(spec, 'pods.target.averageValue')}"

    def _parse_object_metric(self, idx: int, spec: Dict) -> str:
        """解析来源自 Object 的指标信息"""
        current = '<unknown>'
        if getitems(spec, 'object.target.averageValue') is not None:
            if len(self.statuses) > idx:
                obj_cur_avg_val = getitems(self.statuses[idx], 'object.current.averageValue')
                current = obj_cur_avg_val if obj_cur_avg_val else current
            return f"{current}/{getitems(spec, 'object.target.averageValue')} (avg)"
        else:
            if len(self.statuses) > idx:
                obj_cur_val = getitems(self.statuses[idx], 'object.current.value')
                current = obj_cur_val if obj_cur_val else current
            return f"{current}/{getitems(spec, 'object.target.value')}"

    def _parse_resource_metric(self, idx: int, spec: Dict) -> str:
        """解析来源自 Resource 的指标信息"""
        current = '<unknown>'
        if getitems(spec, 'resource.target.averageValue') is not None:
            if len(self.statuses) > idx:
                res_cur_avg_val = getitems(self.statuses[idx], 'resource.current.averageValue')
                current = res_cur_avg_val if res_cur_avg_val else current
            return f"{current}/{getitems(spec, 'resource.target.averageValue')}"
        else:
            if len(self.statuses) > idx:
                res_cur_avg_utilization = getitems(self.statuses[idx], 'resource.current.averageUtilization')
                current = f'{res_cur_avg_utilization}%' if res_cur_avg_utilization else current

            target = '<auto>'
            res_tar_avg_utilization = getitems(spec, 'resource.target.averageUtilization')
            target = f'{res_tar_avg_utilization}%' if res_tar_avg_utilization else target
            return f'{current}/{target}'

    def _parse_container_resource_metric(self, idx: int, spec: Dict) -> str:
        """解析来源自 ContainerResource 的指标信息"""
        current = '<unknown>'
        if getitems(spec, 'containerResource.target.averageValue') is not None:
            if len(self.statuses) > idx:
                res_cur_avg_val = getitems(self.statuses[idx], 'containerResource.current.averageValue')
                current = res_cur_avg_val if res_cur_avg_val else current
            return f"{current}/{getitems(spec, 'containerResource.target.averageValue')}"
        else:
            if len(self.statuses) > idx:
                res_cur_avg_utilization = getitems(self.statuses[idx], 'containerResource.current.averageUtilization')
                current = f'{res_cur_avg_utilization}%' if res_cur_avg_utilization else current

            target = '<auto>'
            res_tar_avg_utilization = getitems(spec, 'containerResource.target.averageUtilization')
            target = f'{res_tar_avg_utilization}%' if res_tar_avg_utilization else target
            return f'{current}/{target}'
