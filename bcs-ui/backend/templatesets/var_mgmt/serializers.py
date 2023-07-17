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
import logging
import re

from django.utils.functional import cached_property
from django.utils.translation import ugettext_lazy as _
from rest_framework import serializers
from rest_framework.exceptions import ValidationError

from backend.container_service.clusters.base.utils import get_clusters
from backend.container_service.clusters.constants import ClusterType
from backend.resources.namespace.utils import get_namespaces
from backend.templatesets.legacy_apps.configuration.constants import VARIABLE_PATTERN

from ..legacy_apps.instance.serializers import InstanceNamespaceSLZ
from .constants import VariableCategory, VariableScope
from .models import Variable

logger = logging.getLogger(__name__)

RE_KEY = re.compile(r'^%s{0,63}$' % VARIABLE_PATTERN)
SYS_KEYS = [
    'SYS_BCS_ZK',
    'SYS_CC_ZK',
    'SYS_BCSGROUP',
    'SYS_TEMPLATE_ID',
    'SYS_VERSION_ID',
    'SYS_VERSION',
    'SYS_INSTANCE_ID',
    'SYS_CREATOR',
    'SYS_UPDATOR',
    'SYS_OPERATOR',
    'SYS_CREATE_TIME',
    'SYS_UPDATE_TIME',
]

# 命名空间变量标识
NAMESPACE_SCOPE = "namespace"


class SearchVariableSLZ(serializers.Serializer):
    scope = serializers.CharField(default='')
    search_key = serializers.CharField(default='')
    limit = serializers.IntegerField(default=10)
    offset = serializers.IntegerField(default=0)
    cluster_type = serializers.ChoiceField(choices=ClusterType.get_choices(), required=False)

    def validate(self, data):
        # 如果是共享集群仅能过滤命名空间下的变量
        if data.get("cluster_type") == ClusterType.SHARED:
            data["scope"] = NAMESPACE_SCOPE

        return data


class ListVariableSLZ(serializers.ModelSerializer):
    default = serializers.DictField(source='get_default_data')
    name = serializers.SerializerMethodField()
    category_name = serializers.SerializerMethodField()
    scope_name = serializers.SerializerMethodField()

    class Meta:
        model = Variable
        fields = (
            'id',
            'name',
            'key',
            'default',
            'default_value',
            'desc',
            'category',
            'category_name',
            'scope',
            'scope_name',
            'creator',
            'created',
            'updated',
            'updator',
        )

    def get_name(self, obj):
        return _(obj.name)

    def get_category_name(self, obj):
        return _(obj.get_category_display())

    def get_scope_name(self, obj):
        return _(obj.get_scope_display())


class VariableSLZ(serializers.ModelSerializer):
    scope = serializers.ChoiceField(choices=VariableScope.get_choices(), required=True)
    name = serializers.CharField(max_length=256, required=True)
    key = serializers.RegexField(
        RE_KEY, max_length=64, required=True, error_messages={'invalid': _("KEY 只能包含字母、数字、中划线和下划线，且以字母开头，最大长度为64个字符")}
    )
    default = serializers.JSONField(required=False)
    desc = serializers.CharField(max_length=256, required=False, allow_blank=True)
    project_id = serializers.CharField(max_length=64, required=True)

    class Meta:
        model = Variable
        fields = ('id', 'name', 'key', 'default', 'desc', 'category', 'scope', 'project_id')

    # TODO add validate_project_id

    def validate_default(self, default):
        if not isinstance(default, dict):
            raise ValidationError(_("default字段非字典类型"))
        if 'value' not in default:
            raise ValidationError(_("default字段没有以value作为键值"))
        return default

    def validate_key(self, key):
        if key in SYS_KEYS:
            raise ValidationError('KEY[{}]{}'.format(key, _("为系统变量名，不允许添加")))
        return key

    def to_representation(self, instance):
        instance.default = instance.get_default_data()
        return super().to_representation(instance)


class CreateVariableSLZ(VariableSLZ):
    def create(self, validated_data):
        exists = Variable.objects.filter(key=validated_data['key'], project_id=validated_data['project_id']).exists()
        if exists:
            detail = {'field': ['{}KEY{}{}'.format(_("变量"), validated_data['key'], _("已经存在"))]}
            raise ValidationError(detail=detail)

        variable = Variable.objects.create(**validated_data)
        return variable


class UpdateVariableSLZ(VariableSLZ):
    def update(self, instance, validated_data):
        if instance.category == VariableCategory.SYSTEM.value:
            raise ValidationError(_("系统内置变量不允许操作"))

        if validated_data.get('key') != instance.key:
            raise ValidationError(_('变量 Key 不允许编辑'))

        if validated_data.get('scope') != instance.scope:
            raise ValidationError(_('变量作用域不允许编辑'))

        instance.name = validated_data.get('name')
        instance.default = validated_data.get('default')
        instance.desc = validated_data.get('desc')
        instance.updator = validated_data.get('updator')
        instance.save()
        return instance


class SearchVariableWithNamespaceSLZ(InstanceNamespaceSLZ):
    namespaces = serializers.CharField(required=True)

    def validate(self, data):
        pass

    def to_internal_value(self, data):
        data = super().to_internal_value(data)
        data['namespaces'] = data['namespaces'].split(',')
        return data


class VariableDeleteSLZ(serializers.Serializer):
    id_list = serializers.JSONField(required=True)


class ClusterVariableSLZ(serializers.Serializer):
    cluster_vars = serializers.JSONField(required=True)


class NsVariableSLZ(serializers.Serializer):
    ns_vars = serializers.JSONField(required=True)


class VariableItemSLZ(serializers.Serializer):
    name = serializers.CharField(max_length=256, required=True)
    key = serializers.RegexField(
        RE_KEY, max_length=64, required=True, error_messages={'invalid': _("KEY 只能包含字母、数字、中划线和下划线，且以字母开头，最大长度为64个字符")}
    )
    value = serializers.CharField(required=True)
    desc = serializers.CharField(default='')
    scope = serializers.ChoiceField(choices=VariableScope.get_choices(), required=True)
    vars = serializers.ListField(child=serializers.JSONField(), required=False)


class ImportVariableSLZ(serializers.Serializer):
    variables = serializers.ListField(child=VariableItemSLZ(), min_length=1)

    @cached_property
    def clusters(self):
        data = get_clusters(self.context['access_token'], self.context['project_id'])
        if data:
            return [c["cluster_id"] for c in data]
        return []

    @cached_property
    def namespaces(self):
        data = get_namespaces(self.context['access_token'], self.context['project_id'])
        return {f"{n['cluster_id']}/{n['name']}": n['id'] for n in data}

    def _validate_cluster_var(self, var):
        for c_var in var['vars']:
            cluster_id = c_var.get('cluster_id')
            if cluster_id not in self.clusters:
                raise ValidationError(_("集群变量中, 集群ID({})不存在").format(cluster_id))
            if 'value' not in c_var:
                raise ValidationError(_("集群变量中, 集群ID({})的value未设置").format(cluster_id))

    def _validate_ns_var(self, var):
        for n_var in var['vars']:
            namespace = f"{n_var.get('cluster_id')}/{n_var.get('namespace')}"
            ns_id = self.namespaces.get(namespace)
            if not ns_id:
                raise ValidationError(_("命名空间变量中, 命名空间({})不存在").format(namespace))

            if 'value' not in n_var:
                raise ValidationError(_("命名空间变量中, 命名空间({})的value未设置").format(namespace))

            n_var['ns_id'] = ns_id

    def validate(self, data):
        for var in data['variables']:
            if var['scope'] == VariableScope.CLUSTER.value:
                self._validate_cluster_var(var)
            if var['scope'] == VariableScope.NAMESPACE.value:
                self._validate_ns_var(var)
        return data
