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
from rest_framework import serializers


class LabelsItemSLZ(serializers.Serializer):
    cluster_id = serializers.CharField()
    inner_ip = serializers.CharField()
    labels = serializers.JSONField(default=[])


class NodeLabelsSLZ(serializers.Serializer):
    node_labels = serializers.ListField(child=LabelsItemSLZ())


class FilterNodeLabelsSLZ(NodeLabelsSLZ):
    pass


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
