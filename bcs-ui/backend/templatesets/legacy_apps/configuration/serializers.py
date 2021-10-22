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
import json
import logging
import re

from django.conf import settings
from django.utils import timezone
from django.utils.translation import ugettext_lazy as _
from jsonschema import SchemaError
from jsonschema import ValidationError as JsonValidationError
from jsonschema import validate as json_validate
from rest_framework import serializers
from rest_framework.exceptions import ValidationError

from backend.templatesets.legacy_apps.instance.constants import InsState
from backend.templatesets.legacy_apps.instance.models import InstanceConfig, VersionInstance

from . import models
from .constants_bak import CONFIGMAP_SCHEM, SECRET_SCHEM, SERVICE_SCHEM
from .constants_k8s import K8S_CONFIGMAP_SCHEM, K8S_SERVICE_SCHEM
from .utils import check_var_by_config, to_bcs_res_name

logger = logging.getLogger(__name__)

RE_NAME = re.compile(r"^[a-z]{1}[a-z0-9-]{0,254}$")
RE_SHOW_NAME = re.compile(r"^[a-zA-Z0-9-_.]{1,45}$")


# TODO refactor 后面去除check_resource_name
def check_resource_name(resource_name, data, name):
    # 校验配置信息中的变量名是否规范
    check_var_by_config(data["config"])

    if ("item_id" in data or "resource_id" in data) and "project_id" in data:
        if "item_id" in data:
            result = models.Template.check_resource_name(
                data["project_id"], resource_name, data["item_id"], name, data["version_id"]
            )
        else:
            result = models.Template.check_resource_name(
                data["project_id"], resource_name, data["resource_id"], name, data["version_id"]
            )
        if not result:
            raise ValidationError(_("{}名称:{}已经在项目模板中被占用,请重新填写").format(resource_name, name))

    return data


def is_tempalte_instance(template_id):
    """模板是否被实例化过"""
    ins_id_list = VersionInstance.objects.filter(
        template_id=template_id, is_bcs_success=True, is_deleted=False
    ).values_list("id", flat=True)
    is_exists = (
        InstanceConfig.objects.filter(instance_id__in=ins_id_list, is_deleted=False, is_bcs_success=True)
        .exclude(ins_state=InsState.NO_INS.value)
        .exists()
    )
    return is_exists


def filter_instance_resource_by_version(data, version_id):
    """只返回实例化过的版本内的资源"""
    exist_instance_ids = VersionInstance.objects.filter(
        version_id=version_id, is_bcs_success=True, is_deleted=False
    ).values_list("id", flat=True)
    _exist_config = InstanceConfig.objects.filter(
        instance_id__in=exist_instance_ids, is_deleted=False, is_bcs_success=True
    )

    new_data = {}
    for _cate in data:
        new_res_list = []
        res_list = data[_cate]
        for _res in res_list:
            _name = _res["name"]
            _is_exist = _exist_config.filter(category=_cate, name=_name).exists()
            if _is_exist:
                new_res_list.append(_res)
        if len(new_res_list):
            new_data[_cate] = new_res_list
    return new_data


def get_template_info(tpl, kind):
    """获取模板的基本信息（模板&模板列表）,不再使用 TemplateSLZ"""
    show_version = models.ShowVersion.objects.filter(template_id=tpl.id).order_by("-updated").first()
    if show_version:
        latest_show_version = show_version.name
        latest_show_version_id = show_version.id
        latest_version = show_version.real_version_id
        latest_version_id = show_version.real_version_id
        latest_version_comment = show_version.comment
    else:
        latest_version = ""
        latest_version_id = 0
        latest_version_comment = ""
        draft = tpl.get_draft
        latest_show_version_id = -1
        if draft:
            latest_show_version = _("草稿")
        else:
            latest_show_version = ""

    data = {
        "id": tpl.id,
        "name": tpl.name,
        "category": tpl.category,
        "desc": tpl.desc,
        "creator": tpl.creator,
        "updator": tpl.updator,
        "created": timezone.localtime(tpl.created).strftime("%Y-%m-%d %H:%M:%S"),
        "updated": timezone.localtime(tpl.updated).strftime("%Y-%m-%d %H:%M:%S"),
        "category_name": tpl.get_category_display(),
        "logo": tpl.log_url,
        "containers": tpl.get_containers(kind, show_version),
        "is_locked": tpl.is_locked,
        "locker": tpl.locker,
        "edit_mode": tpl.edit_mode,
        "latest_version": latest_version,
        "latest_version_id": latest_version_id,
        "latest_show_version": latest_show_version,
        "latest_show_version_id": latest_show_version_id,
        "latest_version_comment": latest_version_comment,
    }
    return data


class ResourceSLZ(serializers.ModelSerializer):
    """模板集版本中的资源信息"""

    data = serializers.SerializerMethodField()
    lb_services = serializers.JSONField(source="get_lb_services", read_only=True)

    class Meta:
        model = models.VersionedEntity
        fields = (
            "data",
            "lb_services",
        )

    def get_data(self, obj):
        data = obj.get_version_instance_resources()
        category = self.context.get("category")
        name = self.context.get("tmpl_app_name")
        # 应用-实例化：只显示应用相关的资源
        if (category in ["deployment", "application"]) or (category in models.POD_RES_LIST):
            # 只查询单个应用实例化所需要的资源
            data = models.get_app_resource(data, category, name, is_related_res=True, version_id=obj.id)
        # 传给前端的资源类型统一
        new_data = {models.CATE_SHOW_NAME.get(x, x): data[x] for x in data}
        return new_data


class ResourceRequstSLZ(serializers.Serializer):
    """资源的请求参数"""

    category = serializers.CharField(required=False, allow_blank=True)
    tmpl_app_name = serializers.CharField(required=False, allow_blank=True)
    is_filter_version = serializers.CharField(required=False, allow_blank=True)

    def validate_category(self, category):
        project_kind = self.context.get("project_kind")
        return to_bcs_res_name(project_kind, category)


class TemplateCreateSLZ(serializers.Serializer):
    """创建／更新模板集基本信息"""

    name = serializers.CharField(max_length=30, required=True)
    desc = serializers.CharField(max_length=50, required=False, allow_blank=True)
    project_id = serializers.CharField(max_length=64)

    def validate(self, data):
        is_exist = models.Template.default_objects.filter(name=data["name"], project_id=data["project_id"]).exists()
        if is_exist:
            detail = {"field": [_("模板集名称[{}]已经存在").format(data["name"])]}
            raise ValidationError(detail=detail)
        return data


# ############################## k8s 相关资源
K8S_RENAME = re.compile(r"^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$")
K8S_NAME_ERROR_MSG = _("名称格式错误，以小写字母或数字开头，只能包含：小写字母、数字、中划线(-)、点(.)")


def check_k8s_deployment_id(data, deploy_tag_list):
    """检查 k8s Service 中关联的 Deployment 是否合法"""
    if not data["version_id"]:
        raise ValidationError(_("请先创建 Deplpyment ，再创建 Service"))

    if not deploy_tag_list:
        raise ValidationError(_("请选择关联的 Deplpyment"))

    return data


class K8sServiceCreateOrUpdateSLZ(serializers.Serializer):
    # k8s 中 service 可不关联应用
    deploy_tag_list = serializers.JSONField(required=False)
    config = serializers.JSONField(required=True)
    version_id = serializers.IntegerField(required=False)
    item_id = serializers.IntegerField(required=False)
    project_id = serializers.CharField(required=False)

    resource_version = serializers.CharField(required=False, allow_blank=True, allow_null=True)
    namespace_id = serializers.CharField(required=False, allow_blank=True, allow_null=True)
    instance_id = serializers.CharField(required=False, allow_blank=True, allow_null=True)
    creator = serializers.CharField(required=False, allow_blank=True, allow_null=True)
    create_time = serializers.CharField(required=False, allow_blank=True, allow_null=True)

    def validate_config(self, config):
        # k8s Service 名称可以支持变量
        # name = config.get('metadata', {}).get('name') or ""
        # if not K8S_RENAME.match(name):
        #     raise ValidationError(
        #         u"Service %s" % K8S_NAME_ERROR_MSG)

        if settings.IS_TEMPLATE_VALIDATE:
            try:
                json_validate(config, K8S_SERVICE_SCHEM)
            except JsonValidationError as e:
                raise ValidationError(_("Service 配置信息格式错误{}").format(e.message))
            except SchemaError as e:
                raise ValidationError(_("Service 配置信息格式错误{}").format(e))

        return json.dumps(config)

    def validate_deploy_tag_list(self, deploy_tag_list):
        if not deploy_tag_list:
            deploy_tag_list = []
        return json.dumps(deploy_tag_list)

    def validate(self, data):
        config = json.loads(data["config"])
        name = config.get("metadata", {}).get("name") or ""

        if not data.get("namespace_id") and not data.get("instance_id"):
            check_resource_name("K8sService", data, name)

        deploy_tag_list = data.get("deploy_tag_list")
        try:
            deploy_tag_list = json.loads(deploy_tag_list)
        except Exception:
            deploy_tag_list = []
        # k8s 中 service 可不关联应用
        data["deploy_tag_list"] = deploy_tag_list
        if not deploy_tag_list:
            return data
        elif not isinstance(deploy_tag_list, list):
            raise ValidationError(_("关联Deployment参数格式错误"))

        return check_k8s_deployment_id(data, deploy_tag_list)


class K8sConfigMapCreateOrUpdateSLZ(serializers.Serializer):
    config = serializers.JSONField(required=True)
    version_id = serializers.IntegerField(required=False)
    item_id = serializers.IntegerField(required=False)
    project_id = serializers.CharField(required=False)

    namespace_id = serializers.CharField(required=False, allow_blank=True, allow_null=True)
    instance_id = serializers.CharField(required=False, allow_blank=True, allow_null=True)

    def validate_config(self, config):
        name = config.get("metadata", {}).get("name") or ""
        if not K8S_RENAME.match(name):
            raise ValidationError(u"ConfigMap %s" % K8S_NAME_ERROR_MSG)

        if settings.IS_TEMPLATE_VALIDATE:
            try:
                json_validate(config, K8S_CONFIGMAP_SCHEM)
            except JsonValidationError as e:
                raise ValidationError(_("ConfigMap 配置信息格式错误{}").format(e.message))
            except SchemaError as e:
                raise ValidationError(_("ConfigMap 配置信息格式错误{}").format(e))

        return json.dumps(config)

    def validate(self, data):
        config = json.loads(data["config"])
        name = config.get("metadata", {}).get("name") or ""
        return check_resource_name("K8sConfigMap", data, name)


class K8sSecretCreateOrUpdateSLZ(serializers.Serializer):
    config = serializers.JSONField(required=True)
    version_id = serializers.IntegerField(required=False)
    item_id = serializers.IntegerField(required=False)
    project_id = serializers.CharField(required=False)

    namespace_id = serializers.CharField(required=False, allow_blank=True, allow_null=True)
    instance_id = serializers.CharField(required=False, allow_blank=True, allow_null=True)

    def validate(self, data):
        config = data["config"]
        # 转换为字符串，方便后续的变量匹配等操作
        data["config"] = json.dumps(config)
        name = config.get("metadata", {}).get("name") or ""
        return check_resource_name("K8sSecret", data, name)
