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
from datetime import datetime
from typing import List, Optional

from prometheus_client import CollectorRegistry, Gauge, Summary

from backend.bcs_web.audit_log.constants import ActivityStatus, ActivityType, ResourceType
from backend.bcs_web.audit_log.models import UserActivityLog


class Exporter:
    def __init__(
        self, registry: CollectorRegistry, resource_type_list: List[ResourceType], start_at: datetime, end_at: datetime
    ):
        self.resource_type_list = resource_type_list
        self.start_at = start_at
        self.end_at = end_at
        self.registry = registry

    def add_summary_by_status(self, status_list: List[str]) -> None:
        """添加 summary metric"""
        count = UserActivityLog.objects.filter(
            resource_type__in=self.resource_type_list,
            activity_type__in=[ActivityType.Add, ActivityType.Modify],
            activity_time__range=(self.start_at, self.end_at),
            activity_status__in=status_list,
        ).count()
        # 添加指标
        s = Summary(
            f"{';'.join(self.resource_type_list)}_summary",
            f"summary from {self.start_at} to {self.end_at}",
            ["status"],
            registry=self.registry,
        )
        s.labels(status_list).observe(count)
        return

    def add_gauge_success_rate(
        self, activity_type_list: List[ResourceType], metric_name: Optional[str] = None
    ) -> None:
        """添加成功率metric"""
        query_filter = UserActivityLog.objects.filter(
            resource_type__in=self.resource_type_list,
            activity_type__in=activity_type_list,
            activity_time__range=(self.start_at, self.end_at),
        )
        total = query_filter.count()
        # NOTE: 认为成功包含 ActivityStatus.Succeed 和 ActivityStatus.Completed
        success_count = query_filter.filter(
            activity_status__in=[ActivityStatus.Succeed, ActivityStatus.Completed],
        ).count()
        # 当没有记录时，认为成功率为1(100%)
        rate = 1
        if total != 0:
            # 保留小数点后两位
            rate = round(success_count / total, 2)
        # 添加指标
        if not metric_name:
            metric_name = f"{';'.join(self.resource_type_list)}_{';'.join(activity_type_list)}_rate"
        g = Gauge(
            metric_name,
            f"gauge from {self.start_at} to {self.end_at}",
            ["status"],
            registry=self.registry,
        )
        g.labels(ActivityStatus.Succeed.value).set(rate)
        return
