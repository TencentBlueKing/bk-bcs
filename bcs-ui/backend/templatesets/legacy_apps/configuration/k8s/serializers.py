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
from django.conf import settings
from django.utils.translation import ugettext_lazy as _
from rest_framework import serializers
from rest_framework.exceptions import ValidationError

from ..constants import K8sResourceName
from ..validator import get_name_from_config, validate_name_duplicate, validate_res_config, validate_variable_inconfig
from .validator import get_config_schema, validate_k8s_res_name, validate_pod_selector


class BCSResourceSLZ(serializers.Serializer):
    config = serializers.JSONField(required=True)
    version_id = serializers.IntegerField(required=False)
    resource_id = serializers.IntegerField(required=False)
    project_id = serializers.CharField(required=False)

    def to_internal_value(self, data):
        data = super().to_internal_value(data)
        if 'name' not in data:
            data['name'] = get_name_from_config(data['config'])
        return data

    def _validate_name_duplicate(self, data):
        validate_name_duplicate(data)

    def _validate_config(self, data):
        pass

    def validate(self, data):
        self._validate_config(data)
        self._validate_name_duplicate(data)
        return data


class K8sPodUnitSLZ(BCSResourceSLZ):
    desc = serializers.CharField(max_length=256, required=False, allow_blank=True)

    def _validate_config(self, data):
        config = data['config']
        resource_name = data['resource_name']
        short_name = resource_name[3:]
        try:
            name = data['name']
            validate_k8s_res_name(name)
        except ValidationError as e:
            raise ValidationError(f'{short_name} {e}')

        try:
            validate_pod_selector(config)
        except ValidationError as e:
            raise ValidationError(_("{}[{}]中选择器{}").format(short_name, name, e))

        # 校验配置信息中的变量名是否规范
        validate_variable_inconfig(config)

        if settings.IS_TEMPLATE_VALIDATE:
            validate_res_config(config, short_name, get_config_schema(resource_name))


class K8sDeploymentSLZ(K8sPodUnitSLZ):
    resource_name = serializers.CharField(default=K8sResourceName.K8sDeployment.value)


class K8sDaemonsetSLZ(K8sPodUnitSLZ):
    resource_name = serializers.CharField(default=K8sResourceName.K8sDaemonSet.value)


class K8sJobSLZ(K8sPodUnitSLZ):
    resource_name = serializers.CharField(default=K8sResourceName.K8sJob.value)


class K8sStatefulSetSLZ(K8sPodUnitSLZ):
    resource_name = serializers.CharField(default=K8sResourceName.K8sStatefulSet.value)
    service_tag = serializers.CharField(default='', allow_blank=True, allow_null=True)


class K8sConfigMapSLZ(BCSResourceSLZ):
    resource_name = serializers.CharField(default=K8sResourceName.K8sConfigMap.value)
    namespace_id = serializers.CharField(required=False, allow_blank=True, allow_null=True)
    instance_id = serializers.CharField(required=False, allow_blank=True, allow_null=True)

    def _validate_config(self, data):
        config = data['config']
        resource_name = data['resource_name']
        short_name = resource_name[3:]

        try:
            name = data['name']
            validate_k8s_res_name(name)
        except ValidationError as e:
            raise ValidationError(f'{short_name} {e}')

        # 校验配置信息中的变量名是否规范
        validate_variable_inconfig(config)

        if settings.IS_TEMPLATE_VALIDATE:
            validate_res_config(config, short_name, get_config_schema(resource_name))


class K8sSecretSLZ(K8sConfigMapSLZ):
    resource_name = serializers.CharField(default=K8sResourceName.K8sSecret.value)


class K8sIngressSLZ(BCSResourceSLZ):
    resource_name = serializers.CharField(default=K8sResourceName.K8sIngress.value)
    namespace_id = serializers.CharField(required=False, allow_blank=True, allow_null=True)

    def _validate_config(self, data):
        short_name = data['resource_name'][3:]
        try:
            name = data['name']
            validate_k8s_res_name(name)
        except ValidationError as e:
            raise ValidationError(f'{short_name} {e}')
        if settings.IS_TEMPLATE_VALIDATE:
            validate_res_config(data['config'], short_name, get_config_schema(data['resource_name']))


class K8sServiceSLZ(BCSResourceSLZ):
    resource_name = serializers.CharField(default=K8sResourceName.K8sService.value)
    deploy_tag_list = serializers.JSONField(required=False)
    resource_version = serializers.CharField(required=False, allow_blank=True, allow_null=True)
    namespace_id = serializers.CharField(required=False, allow_blank=True, allow_null=True)
    instance_id = serializers.CharField(required=False, allow_blank=True, allow_null=True)
    creator = serializers.CharField(required=False, allow_blank=True, allow_null=True)
    create_time = serializers.CharField(required=False, allow_blank=True, allow_null=True)

    def _validate_config(self, data):
        if settings.IS_TEMPLATE_VALIDATE:
            resource_name = data['resource_name']
            short_name = resource_name[3:]
            validate_res_config(data['config'], short_name, get_config_schema(resource_name))

    def validate_deploy_tag_list(self, deploy_tag_list):
        if not deploy_tag_list:
            deploy_tag_list = []
        if not isinstance(deploy_tag_list, list):
            raise ValidationError(_("Service模板: 关联应用参数格式错误"))
        return deploy_tag_list

    def validate(self, data):
        # 目前仅支持配置了selector的Service的创建, 因此需要校验deploy_tag_list字段
        if not data.get('version_id'):
            raise ValidationError(_("请先创建 Deployment/StatefulSet/Daemonset，再创建 Service"))

        if not data.get('deploy_tag_list'):
            raise ValidationError(_("Service模板中{}: 请选择关联的 Deployment/StatefulSet/Daemonset").format(data.get('name')))

        config = data['config']
        if not data.get('namespace_id') and not data.get('instance_id'):
            # 校验配置信息中的变量名是否规范
            validate_variable_inconfig(config)
            self._validate_name_duplicate(data)

        return data


class K8sHPASLZ(BCSResourceSLZ):
    resource_name = serializers.CharField(default=K8sResourceName.K8sHPA.value)

    def _validate_config(self, data):
        config = data['config']
        resource_name = data['resource_name']
        short_name = resource_name[3:]
        try:
            name = data['name']
            validate_k8s_res_name(name)
        except ValidationError as e:
            raise ValidationError(f'{short_name} {e}')

        if settings.IS_TEMPLATE_VALIDATE:
            validate_res_config(config, short_name, get_config_schema(resource_name))
