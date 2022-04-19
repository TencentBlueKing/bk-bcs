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
from django.utils.translation import ugettext_lazy as _
from rest_framework import serializers
from rest_framework.exceptions import ValidationError

from backend.container_service.clusters.constants import ClusterManagerNodeStatus as node_status


class QueryNodeListSLZ(serializers.Serializer):
    node_name_list = serializers.ListField(child=serializers.CharField())


class TaintSLZ(serializers.Serializer):
    key = serializers.CharField()
    value = serializers.CharField(default="", allow_blank=True)
    effect = serializers.CharField()


class NodeTaintSLZ(serializers.Serializer):
    node_name = serializers.CharField()
    taints = serializers.ListField(child=TaintSLZ())


class NodeTaintListSLZ(serializers.Serializer):
    node_taint_list = serializers.ListField(child=NodeTaintSLZ())


class NodeLabelSLZ(serializers.Serializer):
    node_name = serializers.CharField()
    labels = serializers.JSONField(default={})


class NodeLabelListSLZ(serializers.Serializer):
    node_label_list = serializers.ListField(child=NodeLabelSLZ())


class ClusterNodesSLZ(serializers.Serializer):
    host_ips = serializers.ListField(child=serializers.CharField())

    def validate_host_ips(self, host_ips):
        # 限制操作的节点的数量为10个，目的是减少等待时间
        if len(host_ips) > 10:
            raise ValidationError(_("节点数量不能超过10个"))
        return host_ips


class NodeStatusSLZ(serializers.Serializer):
    node_name_list = serializers.ListField(child=serializers.CharField())
    status = serializers.ChoiceField(choices=[node_status.REMOVABLE.value, node_status.RUNNING.value])
