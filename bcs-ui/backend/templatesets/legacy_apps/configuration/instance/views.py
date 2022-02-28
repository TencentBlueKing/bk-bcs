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

DONE：
- 所有 Model 相关的查询都需要查询是否属于：project_id （没有存储外键关系，需要手动处理相应逻辑）
- service 实例化：添加关联 application 的name 到 selector
- 实例化全部的配置文件（不需要前台传参数）
- 生成配置文件时 去掉前端添加的配置项目(Applicaiotn 实例化配置文件中去掉 imageName ／ imageVersion ／ 调度约束类型)
- 环境变量 & 挂载卷 选择 configmap 和 sercret，实例化时处理前端参数
- 负载均衡的实例化,service 中 关联lb 时，1)前端参数校验，2）service 中添加 BCSGROUP label
- 实例化/更新应用: 根据资源的名称查询依赖项 （configmap 和 sercret）
- service 与 Application 关联时，selector 是或的关系

TODO:
- 实例化失败，再次实例化时，应该更新而不是创建数据
"""
import logging

from django.utils.translation import ugettext_lazy as _
from rest_framework import viewsets
from rest_framework.exceptions import ValidationError
from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.response import Response

from backend.bcs_web.audit_log.audit.context import AuditContext
from backend.bcs_web.audit_log.constants import ActivityStatus, ActivityType
from backend.components import paas_cc
from backend.iam.permissions.resources.namespace_scoped import NamespaceScopedPermCtx, NamespaceScopedPermission
from backend.iam.permissions.resources.templateset import TemplatesetPermCtx, TemplatesetPermission
from backend.templatesets.legacy_apps.instance.constants import InsState
from backend.templatesets.legacy_apps.instance.models import InstanceConfig, VersionInstance
from backend.templatesets.legacy_apps.instance.serializers import (
    InstanceNamespaceSLZ,
    PreviewInstanceSLZ,
    VersionInstanceCreateOrUpdateSLZ,
)
from backend.templatesets.legacy_apps.instance.utils import (
    handle_all_config,
    preview_config_json,
    validate_instance_entity,
    validate_ns_by_tempalte_id,
    validate_template_id,
    validate_version_id,
)
from backend.uniapps.application.base_views import error_codes
from backend.utils.renderers import BKAPIRenderer

from ..auditor import TemplatesetAuditor
from ..constants import K8sResourceName
from ..models import CATE_SHOW_NAME, MODULE_DICT
from ..tasks import check_instance_status

logger = logging.getLogger(__name__)


def get_ins_by_template_id(template_id):
    # 获取模板集下所有实例化过的 verison
    show_version_name_list = VersionInstance.objects.filter(template_id=template_id, is_deleted=False).values(
        'show_version_name', 'id'
    )

    # 将 instance_id 按 show_version_name 分组
    show_name_ins_dict = {}
    for _show in show_version_name_list:
        show_version_name = _show['show_version_name']
        ins_id = _show['id']
        if show_version_name in show_name_ins_dict:
            ins_id_list = show_name_ins_dict[show_version_name]
            ins_id_list.append(ins_id)
            show_name_ins_dict[show_version_name] = ins_id_list
        else:
            show_name_ins_dict[show_version_name] = [ins_id]

    # 根据 show_version_name 查询所有被实例化的资源
    return show_name_ins_dict


def get_res_by_show_name(template_id, show_version_name):
    show_name_ins_dict = get_ins_by_template_id(template_id)
    ins_id_list = show_name_ins_dict.get(show_version_name)

    _ins_config_list = InstanceConfig.objects.filter(
        instance_id__in=ins_id_list, is_deleted=False, is_bcs_success=True
    )

    data = {}
    for _config in _ins_config_list:
        _cate = _config.category
        _name = _config.name
        if _cate in data:
            res_data_list = data[_cate]

            res_name_list = [_r['name'] for _r in res_data_list]
            if _name not in res_name_list:
                res_data_list.append({'name': _name, 'id': _name})
        else:
            data[_cate] = [{'name': _name, 'id': _name}]
    return data


class TemplateInstView(viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_exist_showver_name(self, request, project_id, template_id):
        """查询模板集下所有已经实例化过的可见版本名称"""
        validate_template_id(project_id, template_id)
        # 因为没有回滚操作，需要去掉versioninstance中的is_bcs_success为True
        valid_instance = VersionInstance.objects.filter(template_id=template_id, is_deleted=False)
        show_version_names = valid_instance.values('show_version_name', 'show_version_id')

        data = []
        show_version_name_list = []
        for _show in show_version_names:
            _show_name = _show['show_version_name']
            _show_version_id = _show['show_version_id']
            if _show_name not in show_version_name_list:
                show_version_name_list.append(_show_name)
                # 判断版本下是否还有未删除的资源
                _ins_id_list = valid_instance.filter(show_version_name=_show_name).values_list('id', flat=True)
                is_exists = (
                    InstanceConfig.objects.filter(instance_id__in=_ins_id_list, is_deleted=False, is_bcs_success=True)
                    .exclude(ins_state=InsState.NO_INS.value)
                    .exists()
                )
                if is_exists:
                    data.append(
                        {
                            'id': _show_name,
                            'version': _show_name,
                            'template_id': template_id,
                            'show_version_id': _show_version_id,
                        }
                    )
        # 根据 show_version_id（创建时间） 降序排列
        data = sorted(data, key=lambda x: x['show_version_id'], reverse=True)
        return Response({"code": 0, "message": "OK", "data": {"results": data}})

    def get_resource_by_show_name(self, request, project_id, template_id):
        validate_template_id(project_id, template_id)
        show_version_name = request.GET.get('show_version_name')
        data = get_res_by_show_name(template_id, show_version_name)
        # 传给前端的资源类型统一
        new_data = {CATE_SHOW_NAME.get(x, x): data[x] for x in data}
        return Response({"code": 0, "message": "OK", "data": {"data": new_data}})

    def get_exist_version(self, request, project_id, template_id):
        validate_template_id(project_id, template_id)
        show_name_ins_dict = get_ins_by_template_id(template_id)

        # 查询每个版本实例化的数量
        exist_version = {}
        for show_name in show_name_ins_dict:
            ins_id_list = show_name_ins_dict.get(show_name)
            _ins_count = (
                InstanceConfig.objects.filter(instance_id__in=ins_id_list, is_deleted=False, is_bcs_success=True)
                .exclude(ins_state=InsState.NO_INS.value)
                .count()
            )
            if _ins_count > 0:
                exist_version[show_name] = _ins_count
        return Response({"code": 0, "message": "OK", "data": {"exist_version": exist_version}})


class VersionInstanceView(viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def preview_config(self, request, project_id):
        project_kind = request.project.kind
        serializer = PreviewInstanceSLZ(data=request.data, context={'project_kind': project_kind})
        serializer.is_valid(raise_exception=True)
        slz_data = serializer.data

        # 验证 version_id 是否属于该项目
        version_id = slz_data['version_id']
        show_version_id = slz_data['show_version_id']

        template, version_entity = validate_version_id(
            project_id, version_id, is_return_all=True, show_version_id=show_version_id
        )

        # 验证用户是否有模板集的实例化权限
        perm_ctx = TemplatesetPermCtx(username=request.user.username, project_id=project_id, template_id=template.id)
        TemplatesetPermission().can_instantiate(perm_ctx)

        template_id = version_entity.template_id
        version = version_entity.version
        tem_instance_entity = version_entity.get_version_instance_resource_ids

        # 验证前端传过了的预览资源是否在该版本的资源
        req_instance_entity = slz_data.get('instance_entity') or {}
        instance_entity = validate_instance_entity(req_instance_entity, tem_instance_entity)

        access_token = self.request.user.token.access_token
        username = self.request.user.username

        namespace_id = slz_data['namespace']
        lb_info = slz_data.get('lb_info', {})

        resp = paas_cc.get_namespace(access_token, project_id, namespace_id)
        if resp.get('code') != 0:
            return Response(
                {
                    'code': 400,
                    'message': f"查询命名空间(namespace_id:{project_id}-{namespace_id})出错:{resp.get('message')}",
                }
            )

        namespace_info = resp['data']
        perm_ctx = NamespaceScopedPermCtx(
            username=username,
            project_id=project_id,
            cluster_id=namespace_info['cluster_id'],
            name=namespace_info['name'],
        )
        NamespaceScopedPermission().can_use(perm_ctx)

        # 查询当前命名空间的变量信息
        variable_dict = slz_data.get('variable_info', {}).get(namespace_id) or {}
        params = {
            "instance_id": "instanceID",
            "version_id": version_id,
            "show_version_id": show_version_id,
            "template_id": template_id,
            "version": version,
            "project_id": project_id,
            "access_token": access_token,
            "username": username,
            "lb_info": lb_info,
            "variable_dict": variable_dict,
            "is_preview": True,
        }
        data = preview_config_json(namespace_id, instance_entity, **params)
        return Response({"code": 0, "message": "OK", "data": data})

    def get_tmpl_name(self, instance_entity):
        all_tmpl_name_dict = {}
        for category, tmpl_id_list in instance_entity.items():
            # 需要单独处理 Metric, 不需要轮询Metric 的状态
            # 非标准日志采集相关的 configmap\K8sConfigMap 不需要轮询状态
            # secret 相关的也不需要轮询状态
            if category in ['metric', 'configmap', 'K8sConfigMap', 'secret', 'K8sSecret']:
                continue

            if category not in MODULE_DICT:
                raise error_codes.CheckFailed(_("类型只能为{}").format(";".join(MODULE_DICT.keys())))
            tmpl_name = MODULE_DICT[category].objects.filter(id__in=tmpl_id_list).values_list("name", flat=True)
            if category in all_tmpl_name_dict:
                all_tmpl_name_dict[category].extend(list(tmpl_name))
            else:
                all_tmpl_name_dict = {category: list(tmpl_name)}
        return all_tmpl_name_dict

    def post(self, request, project_id):
        """实例化模板"""
        self.project_id = project_id
        version_id = request.data.get('version_id')
        show_version_id = request.data.get('show_version_id')

        template, version_entity = validate_version_id(
            project_id, version_id, is_return_all=True, show_version_id=show_version_id
        )
        # 验证用户是否有模板集实例化权限
        perm_ctx = TemplatesetPermCtx(username=request.user.username, project_id=project_id, template_id=template.id)
        TemplatesetPermission().can_instantiate(perm_ctx)

        self.template_id = version_entity.template_id
        tem_instance_entity = version_entity.get_version_instance_resource_ids

        project_kind = request.project.kind
        self.slz = VersionInstanceCreateOrUpdateSLZ(data=request.data, context={'project_kind': project_kind})
        self.slz.is_valid(raise_exception=True)
        slz_data = self.slz.data

        # 验证前端传过了的预览资源是否在该版本的资源
        req_instance_entity = slz_data.get('instance_entity') or {}
        self.instance_entity = validate_instance_entity(req_instance_entity, tem_instance_entity)

        namespaces = slz_data['namespaces']
        ns_list = namespaces.split(',') if namespaces else []

        access_token = self.request.user.token.access_token
        username = self.request.user.username

        # 判断 template 下 前台传过来的 namespace 是否已经实例化过
        res, ns_name_list, namespace_dict = validate_ns_by_tempalte_id(
            self.template_id, ns_list, access_token, project_id, req_instance_entity
        )
        if not res:
            return Response(
                {
                    "code": 400,
                    "message": _("以下命名空间已经实例化过，不能再实例化\n{}").format("\n".join(ns_name_list)),
                    "data": ns_name_list,
                }
            )

        slz_data['ns_list'] = ns_list
        slz_data['instance_entity'] = self.instance_entity
        slz_data['template_id'] = self.template_id
        slz_data['project_id'] = project_id
        slz_data['version_id'] = version_id
        slz_data['show_version_id'] = show_version_id

        result = handle_all_config(slz_data, access_token, username, project_kind=request.project.kind)
        instance_entity = slz_data.get("instance_entity")
        all_tmpl_name_dict = self.get_tmpl_name(instance_entity)

        # 添加操作记录
        temp_name = version_entity.get_template_name()
        for i in result['success']:
            TemplatesetAuditor(
                audit_ctx=AuditContext(
                    project_id=project_id,
                    user=username,
                    resource=temp_name,
                    resource_id=self.template_id,
                    extra=self.instance_entity,
                    description=_("实例化模板集[{}]命名空间[{}]").format(temp_name, i['ns_name']),
                    activity_type=ActivityType.Add,
                    activity_status=ActivityStatus.Succeed,
                )
            ).log_raw()

        failed_ns_name_list = []
        failed_msg = []
        is_show_failed_msg = False
        # 针对createError的触发后台任务轮训
        if result.get('failed'):
            check_instance_status.delay(
                request.user.token.access_token,
                project_id,
                request.project.get("kind"),
                all_tmpl_name_dict,
                result['failed'],
            )
        for i in result['failed']:
            if i['res_type']:
                description = _("实例化模板集[{}]命名空间[{}]，在实例化{}时失败，错误消息：{}").format(
                    temp_name, i['ns_name'], i['res_type'], i['err_msg']
                )
                failed_ns_name_list.append(_("{}(实例化{}时)").format(i['ns_name'], i['res_type']))
            else:
                description = _("实例化模板集[{}]命名空间[{}]失败，错误消息：{}").format(temp_name, i['ns_name'], i['err_msg'])
                failed_ns_name_list.append(i['ns_name'])
                if i.get('show_err_msg'):
                    failed_msg.append(i['err_msg'])
                    is_show_failed_msg = True

            TemplatesetAuditor(
                audit_ctx=AuditContext(
                    project_id=project_id,
                    user=username,
                    resource=temp_name,
                    resource_id=self.template_id,
                    extra=self.instance_entity,
                    description=description,
                    activity_type=ActivityType.Add,
                    activity_status=ActivityStatus.Failed,
                )
            ).log_raw()

            if is_show_failed_msg:
                msg = '\n'.join(failed_msg)
            else:
                msg = _("以下命名空间实例化失败，\n{}，请联系集群管理员解决").format("\n".join(failed_ns_name_list))
            if failed_ns_name_list:
                return Response({"code": 400, "message": msg, "data": failed_ns_name_list})

        return Response(
            {
                "code": 0,
                "message": "OK",
                "data": {
                    "version_id": version_id,
                    "template_id": self.template_id,
                },
            }
        )


class InstanceNameSpaceView(viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def post(self, request, project_id, version_id):
        version_entity = validate_version_id(project_id, version_id, is_version_entity_retrun=True)
        template_id = version_entity.template_id

        project_kind = request.project.kind
        self.slz = InstanceNamespaceSLZ(data=request.data, context={'project_kind': project_kind})
        self.slz.is_valid(raise_exception=True)
        slz_data = self.slz.data

        if 'instance_entity' not in slz_data:
            raise ValidationError(_("请选择要实例化的模板"))
        instance_entity = slz_data['instance_entity']

        # 根据template_id 查询已经被实例化过的 ns
        exist_instance_id = VersionInstance.objects.filter(template_id=template_id, is_deleted=False).values_list(
            'id', flat=True
        )
        filter_ns = InstanceConfig.objects.filter(
            instance_id__in=exist_instance_id, is_deleted=False, is_bcs_success=True
        ).exclude(ins_state=InsState.NO_INS.value)

        exist_ns = []
        # 查询每类资源已经实例化的ns，求合集，这些已经实例化过的ns不能再被实例化
        for cate in instance_entity:
            cate_data = instance_entity[cate]
            cate_name_list = [i.get('name') for i in cate_data if i.get('name')]
            cate_ns = filter_ns.filter(category=cate, name__in=cate_name_list)
            exist_ns.extend(list(cate_ns))

        ns_resources = {}
        for inst_config in exist_ns:
            ns_id = int(inst_config.namespace)

            # HPA只通过模板集管理，可以重试实例化(apply操作)
            if inst_config.category == K8sResourceName.K8sHPA.value:
                continue

            if ns_id not in ns_resources:
                ns_resources[ns_id] = [
                    inst_config.category,
                ]
            else:
                ns_resources[ns_id].append(inst_config.category)

        for ns_id, resources in ns_resources.items():
            ns_resources[ns_id] = list(set(resources))

        return Response({"code": 0, "message": "OK", "data": {"ns_resources": ns_resources}})
