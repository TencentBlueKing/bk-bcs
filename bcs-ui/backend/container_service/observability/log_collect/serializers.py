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
from typing import Any, Dict

from rest_framework import serializers
from rest_framework.validators import ValidationError

from backend.utils.basic import getitems

from .constants import LogSourceType, SupportedWorkload
from .formatter import format
from .models import LogCollectMetadata


class BaseConfSLZ(serializers.Serializer):
    enable_stdout = serializers.BooleanField(default=False)
    log_paths = serializers.ListField(child=serializers.CharField(), required=False)

    def validate(self, data):
        if data['enable_stdout'] is False and not data.get('log_paths'):
            raise ValidationError('log_paths field is required when enable_stdout field is False')
        return data


class ContainerConfSLZ(BaseConfSLZ):
    name = serializers.CharField()


class WorkLoadSLZ(serializers.Serializer):
    name = serializers.CharField()
    kind = serializers.ChoiceField(choices=SupportedWorkload.get_choices())
    container_confs = serializers.ListField(child=ContainerConfSLZ(), min_length=1)


class SelectorSLZ(BaseConfSLZ):
    match_labels = serializers.JSONField(required=False, default=dict)
    match_expressions = serializers.JSONField(required=False, default=list)

    def validate(self, data):
        data = super().validate(data)

        match_labels = data.get('match_labels')
        match_expressions = data.get('match_expressions')

        if not match_labels and not match_expressions:
            raise ValidationError('match_labels or match_expressions field is required')

        if match_labels and not isinstance(match_labels, dict):
            raise ValidationError('match_labels must be a valid dict')

        if match_expressions and not isinstance(match_expressions, list):
            raise ValidationError('match_expressions must be a valid list')

        return data


class UpdateOrCreateCollectConfSLZ(serializers.Serializer):
    """
    更新或创建日志采集规则 SLZ. 其中, base, workload, selector 字段分别对应 log_source_type 为
    ALL_CONTAINERS, SELECTED_CONTAINERS, SELECTED_LABELS 时必须的容器日志采集参数
    """

    log_source_type = serializers.ChoiceField(choices=LogSourceType.get_choices())
    bk_biz_id = serializers.IntegerField()
    project_id = serializers.CharField()
    cluster_id = serializers.CharField()
    config_name = serializers.RegexField(regex=r'^[a-zA-Z0-9_]+$', min_length=5, max_length=50)
    namespace = serializers.CharField(required=False)
    add_pod_label = serializers.BooleanField(default=False)
    extra_labels = serializers.JSONField(default=dict)
    base = BaseConfSLZ(required=False)
    workload = WorkLoadSLZ(required=False)
    selector = SelectorSLZ(required=False)

    def validate(self, data):
        log_source_type = data['log_source_type']
        if log_source_type == LogSourceType.ALL_CONTAINERS:
            base = data.get('base')
            if not base:
                raise ValidationError('base field is required for log source(All Containers)')

        elif log_source_type == LogSourceType.SELECTED_CONTAINERS:
            workload = data.get('workload')
            if not workload:
                raise ValidationError('workload field is required for log source(Selected Containers)')

        else:
            selector = data.get('selector')
            if not selector:
                raise ValidationError('selector field is required for log source(Selected Labels)')

        return data


class CollectConfSLZ(serializers.ModelSerializer):
    add_pod_label = serializers.SerializerMethodField()
    extra_labels = serializers.SerializerMethodField()
    config = serializers.SerializerMethodField()
    # 标记日志平台规则是否被后台删除
    deleted = serializers.SerializerMethodField()

    class Meta:
        model = LogCollectMetadata
        exclude = ('id', 'project_id', 'is_deleted', 'deleted_time')

    def get_add_pod_label(self, obj) -> bool:
        return self._getitem(obj, 'add_pod_label')

    def get_extra_labels(self, obj) -> Dict[str, str]:
        extra_labels = self._getitem(obj, 'extra_labels')
        if extra_labels:
            return {label['key']: label['value'] for label in extra_labels}
        return {}

    def get_config(self, obj) -> Dict[str, Any]:
        container_config = self._getitem(obj, 'container_config')
        if container_config:
            return format(obj.log_source_type, container_config)
        return {}

    def get_deleted(self, obj) -> bool:
        if obj.config_id not in self.context['rule_configs']:
            return True
        return False

    def _getitem(self, obj, key) -> Any:
        return getitems(self.context['rule_configs'], [obj.config_id, key])

    def to_representation(self, instance):
        data = super().to_representation(instance)
        data.update(data.pop('config'))
        return data


class QueryLogLinksSLZ(serializers.Serializer):
    bk_biz_id = serializers.IntegerField()
    container_ids = serializers.CharField(required=False)

    def validate(self, data):
        container_ids = data.get('container_ids')
        if container_ids:
            data['container_ids'] = container_ids.split(',')

        return data
