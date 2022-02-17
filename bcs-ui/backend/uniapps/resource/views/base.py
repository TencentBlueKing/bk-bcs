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
import base64
import copy
import datetime
import json
import logging

from django.utils.translation import ugettext_lazy as _
from rest_framework.exceptions import ValidationError
from rest_framework.response import Response

from backend.accounts import bcs_perm
from backend.bcs_web.audit_log import client as activity_client
from backend.components.bcs import k8s
from backend.container_service.clusters.base.utils import get_cluster_type
from backend.container_service.clusters.constants import ClusterType
from backend.iam.permissions.resources.namespace import calc_iam_ns_id
from backend.iam.permissions.resources.namespace_scoped import NamespaceScopedPermCtx, NamespaceScopedPermission
from backend.resources.namespace.constants import K8S_PLAT_NAMESPACE, K8S_SYS_NAMESPACE
from backend.templatesets.legacy_apps.instance.constants import (
    ANNOTATIONS_CREATOR,
    ANNOTATIONS_UPDATE_TIME,
    ANNOTATIONS_UPDATOR,
    K8S_IMAGE_SECRET_PRFIX,
    LABLE_INSTANCE_ID,
    LABLE_TEMPLATE_ID,
    SOURCE_TYPE_LABEL_KEY,
)
from backend.templatesets.legacy_apps.instance.drivers import get_scheduler_driver
from backend.templatesets.legacy_apps.instance.funutils import render_mako_context, update_nested_dict
from backend.templatesets.legacy_apps.instance.generator import GENERATOR_DICT
from backend.templatesets.legacy_apps.instance.models import InstanceConfig
from backend.uniapps import utils as app_utils
from backend.uniapps.application.constants import DELETE_INSTANCE, SOURCE_TYPE_MAP
from backend.uniapps.network.serializers import BatchResourceSLZ
from backend.uniapps.resource.constants import CREATE_TIME_REGEX
from backend.utils.basic import getitems
from backend.utils.errcodes import ErrorCode

logger = logging.getLogger(__name__)


class ResourceOperate:

    category = None
    cate = None
    # 更新相关参数
    sys_config = None
    slz = None
    desc = "cluster: {cluster_id}, namespace: {namespace}, delete {resource_name}: {name}"

    def delete_single_resource(self, request, project_id, cluster_id, namespace, namespace_id, name):
        if get_cluster_type(cluster_id) == ClusterType.SHARED:
            return Response({"code": 400, "message": _("无法操作共享集群资源")})

        username = request.user.username
        access_token = request.user.token.access_token

        if namespace in K8S_SYS_NAMESPACE:
            return {
                "code": 400,
                "message": _("不允许操作系统命名空间[{}]").format(','.join(K8S_SYS_NAMESPACE)),
            }
        client = k8s.K8SClient(access_token, project_id, cluster_id, env=None)
        curr_func = getattr(client, "delete_%s" % self.category)
        resp = curr_func(namespace, name)

        if resp.get("code") == ErrorCode.NoError:
            # 删除成功则更新状态
            now_time = datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')
            InstanceConfig.objects.filter(namespace=namespace_id, category=self.cate, name=name).update(
                creator=username,
                updator=username,
                oper_type=DELETE_INSTANCE,
                updated=now_time,
                deleted_time=now_time,
                is_deleted=True,
                is_bcs_success=True,
            )
        return {
            "code": resp.get("code"),
            "message": resp.get("message"),
        }

    def delete_resource(self, request, project_id, cluster_id, namespace, name):
        username = request.user.username

        # 检查用户是否有命名空间的使用权限
        namespace_id = app_utils.get_namespace_id(
            request.user.token.access_token, project_id, (cluster_id, namespace), cluster_id=cluster_id
        )
        app_utils.can_use_namespace(request, project_id, cluster_id, namespace)

        resp = self.delete_single_resource(request, project_id, cluster_id, namespace, namespace_id, name)
        # 添加操作审计
        activity_client.ContextActivityLogClient(
            project_id=project_id,
            user=username,
            resource_type="instance",
            resource=name,
            resource_id=0,
            extra=json.dumps({}),
            description=self.desc.format(
                cluster_id=cluster_id, namespace=namespace, resource_name=self.category, name=name
            ),
        ).log_modify(activity_status="succeed" if resp.get("code") == ErrorCode.NoError else "failed")

        # 已经删除的，需要将错误信息翻译一下
        message = resp.get('message', '')
        is_delete_before = True if 'node does not exist' in message or 'not found' in message else False
        if is_delete_before:
            message = _("{}[命名空间:{}]已经被删除，请手动刷新数据").format(name, namespace)
        return Response({"code": resp.get("code"), "message": message, "data": {}})

    def batch_delete_resource(self, request, project_id):
        """批量删除资源"""
        username = request.user.username

        slz = BatchResourceSLZ(data=request.data)
        slz.is_valid(raise_exception=True)
        data = slz.data['data']

        namespace_list = [(ns['cluster_id'], ns.get('namespace')) for ns in data]
        namespace_list = set(namespace_list)

        # 检查用户是否有命名空间的使用权限
        app_utils.can_use_namespaces(request, project_id, namespace_list)

        # namespace_dict format: {(cluster_id, ns_name): ns_id}
        namespace_dict = app_utils.get_ns_id_map(request.user.token.access_token, project_id)

        success_list = []
        failed_list = []
        for _d in data:
            cluster_id = _d.get('cluster_id')
            name = _d.get('name')
            namespace = _d.get('namespace')
            namespace_id = namespace_dict.get((cluster_id, namespace))
            # 删除service
            resp = self.delete_single_resource(request, project_id, cluster_id, namespace, namespace_id, name)
            # 处理已经删除，但是storage上报数据延迟的问题
            message = resp.get('message', '')
            is_delete_before = True if 'node does not exist' in message or 'not found' in message else False
            if resp.get("code") == ErrorCode.NoError:
                success_list.append(
                    {
                        'name': name,
                        'desc': self.desc.format(
                            cluster_id=cluster_id, namespace=namespace, resource_name=self.category, name=name
                        ),
                    }
                )
            else:
                if is_delete_before:
                    message = _('已经被删除，请手动刷新数据')
                desc = self.desc.format(
                    cluster_id=cluster_id, namespace=namespace, resource_name=self.category, name=name
                )
                failed_list.append(
                    {
                        'name': name,
                        'desc': f'{desc}, message: {message}',
                    }
                )
        code = 0
        message = ''
        # 添加操作审计
        if success_list:
            name_list = [_s.get('name') for _s in success_list]
            desc_list = [_s.get('desc') for _s in success_list]
            message = _("以下{}删除成功:{}").format(self.category, ";".join(desc_list))
            activity_client.ContextActivityLogClient(
                project_id=project_id,
                user=username,
                resource_type="instance",
                resource=';'.join(name_list),
                resource_id=0,
                extra=json.dumps({}),
                description=";".join(desc_list),
            ).log_modify(activity_status="succeed")

        if failed_list:
            name_list = [_s.get('name') for _s in failed_list]
            desc_list = [_s.get('desc') for _s in failed_list]

            code = 4004
            message = _("以下{}删除失败:{}").format(self.category, ";".join(desc_list))
            activity_client.ContextActivityLogClient(
                project_id=project_id,
                user=username,
                resource_type="instance",
                resource=';'.join(name_list),
                resource_id=0,
                extra=json.dumps({}),
                description=message,
            ).log_modify(activity_status="failed")

        return Response({"code": code, "message": message, "data": {}})

    def update_resource(self, request, project_id, cluster_id, namespace, name):
        """更新"""
        if get_cluster_type(cluster_id) == ClusterType.SHARED:
            return Response({"code": 400, "message": _("无法操作共享集群资源")})

        access_token = request.user.token.access_token

        if namespace in K8S_SYS_NAMESPACE:
            return Response(
                {"code": 400, "message": _("不允许操作系统命名空间[{}]").format(','.join(K8S_SYS_NAMESPACE)), "data": {}}
            )

        request_data = request.data or {}
        request_data['project_id'] = project_id
        # 验证请求参数
        slz = self.slz(data=request.data)
        slz.is_valid(raise_exception=True)
        data = slz.data

        try:
            config = json.loads(data["config"])
        except Exception:
            config = data["config"]
        namespace_id = data['namespace_id']
        username = request.user.username

        # 检查是否有命名空间的使用权限
        perm_ctx = NamespaceScopedPermCtx(
            username=username, project_id=project_id, cluster_id=cluster_id, name=namespace
        )
        NamespaceScopedPermission().can_use(perm_ctx)

        # 对配置文件做处理
        gparams = {"access_token": access_token, "project_id": project_id, "username": username}
        generator = GENERATOR_DICT.get(self.cate)(0, namespace_id, **gparams)
        config = generator.handle_db_config(db_config=config)
        # 获取上下文信息
        context = generator.context
        now_time = context.get('SYS_UPDATE_TIME')
        instance_id = data.get('instance_id', 0)
        context.update(
            {
                'SYS_CREATOR': data.get('creator', ''),
                'SYS_CREATE_TIME': data.get('create_time', ''),
                'SYS_INSTANCE_ID': instance_id,
            }
        )

        # 生成配置文件
        sys_config = copy.deepcopy(self.sys_config)
        resource_config = update_nested_dict(config, sys_config)
        resource_config = json.dumps(resource_config)
        try:
            config_profile = render_mako_context(resource_config, context)
        except Exception:
            logger.exception(u"配置文件变量替换出错\nconfig:%s\ncontext:%s" % (resource_config, context))
            raise ValidationError(_("配置文件中有未替换的变量"))

        config_profile = generator.format_config_profile(config_profile)

        service_name = config.get('metadata', {}).get('name')
        _config_content = {'name': service_name, 'config': json.loads(config_profile), 'context': context}

        # 更新db中的数据
        config_objs = InstanceConfig.objects.filter(
            namespace=namespace_id,
            category=self.cate,
            name=service_name,
        )
        if config_objs.exists():
            config_objs.update(
                creator=username,
                updator=username,
                oper_type='update',
                updated=now_time,
                is_deleted=False,
            )
            _instance_config = config_objs.first()
        else:
            _instance_config = InstanceConfig.objects.create(
                namespace=namespace_id,
                category=self.cate,
                name=service_name,
                config=config_profile,
                instance_id=instance_id,
                creator=username,
                updator=username,
                oper_type='update',
                updated=now_time,
                is_deleted=False,
            )
        _config_content['instance_config_id'] = _instance_config.id
        configuration = {namespace_id: {self.cate: [_config_content]}}

        driver = get_scheduler_driver(access_token, project_id, configuration, request.project.kind)
        result = driver.instantiation(is_update=True)

        failed = []
        if isinstance(result, dict):
            failed = result.get('failed') or []
        # 添加操作审计
        activity_client.ContextActivityLogClient(
            project_id=project_id,
            user=username,
            resource_type="instance",
            resource=service_name,
            resource_id=_instance_config.id,
            extra=json.dumps(configuration),
            description=_("更新{}[{}]命名空间[{}]").format(self.category, service_name, namespace),
        ).log_modify(activity_status="failed" if failed else "succeed")

        if failed:
            return Response(
                {
                    "code": 400,
                    "message": _("{}[{}]在命名空间[{}]更新失败，请联系集群管理员解决").format(self.category, service_name, namespace),
                    "data": {},
                }
            )
        return Response({"code": 0, "message": "OK", "data": {}})

    def handle_data(
        self,
        data,
        s_cate,
        cluster_id,
        is_decode,
        namespace_dict=None,
    ):
        for _s in data:
            _config = _s.get('data', {})
            annotations = _config.get('metadata', {}).get('annotations', {})
            _s['creator'] = annotations.get(ANNOTATIONS_CREATOR, '')
            _s['create_time'] = _s.get('createTime', '')
            _s['update_time'] = annotations.get(ANNOTATIONS_UPDATE_TIME, _s['create_time'])
            _s['updator'] = annotations.get(ANNOTATIONS_UPDATOR, _s['creator'])
            _s['status'] = 'Running'

            _s['can_update'] = True
            _s['can_update_msg'] = ''
            _s['can_delete'] = True
            _s['can_delete_msg'] = ''

            _s['iam_ns_id'] = calc_iam_ns_id(cluster_id, _s['namespace'])
            _s['namespace_id'] = namespace_dict.get((cluster_id, _s['namespace'])) if namespace_dict else None
            _s['cluster_id'] = cluster_id
            _s['name'] = _s['resourceName']

            labels = _config.get('metadata', {}).get('labels', {})
            # 获取模板集信息
            template_id = labels.get(LABLE_TEMPLATE_ID)
            instance_id = labels.get(LABLE_INSTANCE_ID)
            # 资源来源
            source_type = labels.get(SOURCE_TYPE_LABEL_KEY)
            if not source_type:
                source_type = "template" if template_id else "other"
            _s['source_type'] = SOURCE_TYPE_MAP.get(source_type)

            # 处理 k8s 的系统命名空间的数据
            if _s['namespace'] in K8S_SYS_NAMESPACE:
                _s['can_update'] = _s['can_delete'] = False
                _s['can_update_msg'] = _s['can_delete_msg'] = _("不允许操作系统命名空间")
                continue

            # 处理平台集群和命名空间下的数据
            if _s['namespace'] in K8S_PLAT_NAMESPACE:
                _s['can_update'] = _s['can_delete'] = False
                _s['can_update_msg'] = _s['can_delete_msg'] = _("不允许操作平台命名空间")
                continue

            # 处理创建命名空间时生成的default-token-xxx
            if s_cate == 'K8sSecret' and _s['name'].startswith('default-token'):
                is_namespace_default_token = True
            else:
                is_namespace_default_token = False

            # 处理系统默认生成的Secret
            if s_cate == 'K8sSecret' and _s['name'] == '%s%s' % (K8S_IMAGE_SECRET_PRFIX, _s['namespace']):
                is_k8s_image_secret = True
            else:
                is_k8s_image_secret = False

            if is_k8s_image_secret or is_namespace_default_token:
                _s['can_update'] = _s['can_delete'] = False
                _s['can_update_msg'] = _s['can_delete_msg'] = _("不允许操作系统数据")
                continue

            if template_id:
                try:
                    instance_id = int(instance_id)
                except Exception:
                    instance_id = 0
            else:
                # 非模板集创建，可以删除但是不可以更新
                _s['can_update'] = False
                _s['can_update_msg'] = _("不是由模板实例化生成，无法更新")

            _s['instance_id'] = instance_id

            # k8s Secret base64 解码内容
            if is_decode and _s['can_update'] and s_cate == 'K8sSecret':
                _d = _config.get('data')
                for _key in _d:
                    if _d[_key]:
                        try:
                            _d[_key] = base64.b64decode(_d[_key]).decode("utf-8")
                        except Exception:
                            pass

        # k8s configmap / secret 获取的是 data 中的数据
        ret_data = []
        for info in data:
            if 'createTime' in info:
                info["createTime"] = ' '.join(CREATE_TIME_REGEX.findall(info["createTime"])[:2])
            info_data = getitems(info, ['data', 'data'], {})
            if info_data:
                info['data']['data'] = dict(sorted(info_data.items(), key=lambda x: x[0]))
            ret_data.append(info)
