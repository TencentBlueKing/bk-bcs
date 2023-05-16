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
import time

import arrow
from django.utils.translation import ugettext_lazy as _
from rest_framework import serializers
from rest_framework.exceptions import ValidationError

from backend.components import data as data_api
from backend.container_service.clusters import constants
from backend.container_service.clusters.models import NodeLabel

# metrics 默认时间 1小时
METRICS_DEFAULT_TIMEDELTA = 3600


class NodeLabelSLZ(serializers.ModelSerializer):
    project_id = serializers.CharField(max_length=32, required=True)
    cluster_id = serializers.CharField(max_length=32, required=True)
    node_id = serializers.IntegerField(required=True)
    labels = serializers.JSONField(required=False)

    class Meta:
        model = NodeLabel
        fields = ('id', 'project_id', 'cluster_id', 'node_id', 'labels', 'creator', 'updator')


class NodeLabelUpdateSLZ(serializers.ModelSerializer):
    node_id = serializers.IntegerField(required=True)
    labels = serializers.JSONField(required=False)

    class Meta:
        model = NodeLabel
        fields = ('id', 'project_id', 'cluster_id', 'node_id', 'labels', 'creator', 'updator')


class SearchResourceBaseSLZ(serializers.Serializer):
    res_id = serializers.CharField(required=True)


class SummaryMetricsSLZ(SearchResourceBaseSLZ):
    ip_resource = serializers.CharField(required=False)


class MetricsSLZBase(serializers.Serializer):
    metric = serializers.ChoiceField(choices=list(data_api.NodeMetricFields.keys()))
    start_at = serializers.DateTimeField(required=False)
    end_at = serializers.DateTimeField(required=False)

    def validate(self, data):
        now = int(time.time() * 1000)
        # handle the start_at
        if 'start_at' in data:
            data['start_at'] = arrow.get(data['start_at']).timestamp * 1000
        else:
            # default one hour
            data['start_at'] = now - METRICS_DEFAULT_TIMEDELTA * 1000
        # handle the end_at
        if 'end_at' in data:
            data['end_at'] = arrow.get(data['end_at']).timestamp * 1000
        else:
            data['end_at'] = now
        # start_at must be less than end_at
        if data['end_at'] <= data['start_at']:
            raise ValidationError(_('param[start_at] must be less than [end_at]'))
        return data


class MetricsSLZ(MetricsSLZBase):
    res_id = serializers.CharField(required=True)


class MetricsMultiSLZ(MetricsSLZBase):
    res_id_list = serializers.ListField(required=True)


class FetchCCHostSLZ(serializers.Serializer):
    """获取 CMDB 业务下可用主机列表"""

    limit = serializers.IntegerField(label=_('查询行数'), default=constants.DEFAULT_NODE_LIMIT)
    offset = serializers.IntegerField(label=_('偏移量'), default=0)
    ip_list = serializers.ListField(label=_('待过滤 IP 列表'), default=list)
    set_id = serializers.IntegerField(label=_('集群 ID'), default=None)
    module_id = serializers.IntegerField(label=_('模块 ID'), default=None)
    fuzzy = serializers.BooleanField(label=_('是否模糊匹配 IP'), default=False)
    desire_all_data = serializers.BooleanField(label=_('请求全量数据'), default=False)
