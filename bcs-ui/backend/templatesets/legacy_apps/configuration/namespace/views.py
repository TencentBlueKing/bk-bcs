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
from itertools import groupby

from django.conf import settings
from django.utils.translation import ugettext_lazy as _
from rest_framework import response, viewsets
from rest_framework.renderers import BrowsableAPIRenderer

from backend.bcs_web.audit_log.audit.decorators import log_audit_on_view
from backend.bcs_web.audit_log.constants import ActivityType
from backend.components import paas_cc
from backend.components.bcs.k8s import K8SClient
from backend.container_service.clusters.base.utils import append_shared_clusters, get_cluster_type, get_clusters
from backend.container_service.clusters.constants import ClusterType
from backend.container_service.projects.base.constants import LIMIT_FOR_ALL_DATA
from backend.iam.permissions.decorators import response_perms
from backend.iam.permissions.resources.namespace import (
    NamespaceAction,
    NamespaceCreatorAction,
    NamespacePermCtx,
    NamespacePermission,
    NamespaceRequest,
    calc_iam_ns_id,
)
from backend.resources import namespace as ns_resource
from backend.resources.namespace.constants import K8S_PLAT_NAMESPACE, PROJ_CODE_ANNO_KEY
from backend.resources.namespace.utils import get_namespace_by_id
from backend.templatesets.legacy_apps.configuration.constants import EnvType
from backend.templatesets.legacy_apps.configuration.utils import get_cluster_env_name
from backend.templatesets.var_mgmt.models import NameSpaceVariable
from backend.utils.basic import str2bool
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes
from backend.utils.renderers import BKAPIRenderer
from backend.utils.response import PermsResponse

from . import serializers as slz
from .auditor import NamespaceAuditor, NamespaceQuotaAuditor
from .resources import Namespace
from .tasks import sync_namespace as sync_ns_task

logger = logging.getLogger(__name__)


class NamespaceBase:
    """命名空间操作的基本方法
    其他地方也要用到，所以提取为单独的类
    """

    def create_ns_by_bcs(self, client, name, data, project_code):
        # 注解中添加上标识projectcode的信息，用于查询当前项目下，共享集群中的命名空间
        ns_config = {
            "apiVersion": "v1",
            "kind": "Namespace",
            "metadata": {"name": name, "annotations": {PROJ_CODE_ANNO_KEY: project_code}},
        }
        result = client.create_namespace(ns_config)
        # 通过错误消息判断 Namespace 是否已经存在，已经存在则直接进行下一步
        res_msg = result.get('message') or ''
        is_already_exists = res_msg.endswith("already exists")
        if result.get('code') != 0 and not is_already_exists:
            raise error_codes.ComponentError(_("创建Namespace失败，{}").format(result.get('message')))

    def delete_ns_by_bcs(self, client, name):
        result = client.delete_namespace(name)
        if result.get('code') != 0:
            raise error_codes.ComponentError.f(_("创建Namespace失败，{}").format(result.get('message')))

    def init_namespace_by_bcs(self, access_token, project_id, project_code, data):
        """k8s 的集群需要创建 Namespace"""
        client = K8SClient(access_token, project_id, data['cluster_id'], env=None)
        name = data['name']
        # 创建 ns
        self.create_ns_by_bcs(client, name, data, project_code)
        # 如果需要使用资源配额，创建配额
        if data.get("quota"):
            client = ns_resource.NamespaceQuota(access_token, project_id, data["cluster_id"])
            client.create_namespace_quota(name, data["quota"])


class NamespaceView(NamespaceBase, viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)
    iam_perm = NamespacePermission()

    def get_ns(self, request, project_id, namespace_id):
        """获取单个命名空间的信息"""
        access_token = request.user.token.access_token
        ns_info = get_namespace_by_id(access_token, project_id, namespace_id)
        return response.Response(ns_info)

    def get_clusters_without_ns(self, clusters, cluster_ids_with_ns):
        """获取不带有ns的集群"""
        clusters_without_ns = []
        for cluster_id in clusters:
            if cluster_id in cluster_ids_with_ns:
                continue
            cluster = clusters[cluster_id]
            item = {
                "environment_name": get_cluster_env_name(cluster["environment"]),
                "environment": cluster["environment"],
                "cluster_id": cluster_id,
                "name": cluster["name"],
                "results": [],
            }
            if get_cluster_type(cluster_id) == ClusterType.SHARED:
                item["is_shared"] = True
            clusters_without_ns.append(item)
        return clusters_without_ns

    def _ignore_ns_for_k8s(self, ns_list):
        """针对k8s集群，过滤掉系统和平台命名空间"""
        return [ns for ns in ns_list if ns["name"] not in K8S_PLAT_NAMESPACE]

    @response_perms(
        action_ids=[NamespaceAction.VIEW, NamespaceAction.UPDATE, NamespaceAction.DELETE],
        permission_cls=NamespacePermission,
        resource_id_key='iam_ns_id',
    )
    def list(self, request, project_id):
        """命名空间列表
        权限控制: 必须有对应集群的使用权限
        """
        access_token = request.user.token.access_token
        valid_group_by = ['env_type', 'cluster_id', 'cluster_name']

        group_by = request.GET.get('group_by')
        cluster_id = request.GET.get('cluster_id')

        # 获取全部namespace，前台分页
        result = paas_cc.get_namespace_list(access_token, project_id, limit=LIMIT_FOR_ALL_DATA)
        if result.get('code') != 0:
            raise error_codes.APIError.f(result.get('message', ''))

        results = result["data"]["results"] or []
        # 针对k8s集群过滤掉平台命名空间
        results = self._ignore_ns_for_k8s(results)

        # 补充cluster_name字段
        cluster_list = get_clusters(access_token, project_id)
        # 添加共享集群
        cluster_list = append_shared_clusters(cluster_list)
        # TODO: 后续发现cluster_id不存在时，再处理
        cluster_dict = {i["cluster_id"]: i for i in (cluster_list or [])}

        # no_vars=1 不显示变量
        no_vars = request.GET.get('no_vars')
        if no_vars == '1':
            project_var = []
        else:
            project_var = NameSpaceVariable.get_project_ns_vars(project_id)

        for i in results:
            # ns_vars = NameSpaceVariable.get_ns_vars(i['id'], project_id)
            ns_id = i['id']
            ns_vars = []
            for _var in project_var:
                _ns_values = _var['ns_values']
                _ns_value_ids = _ns_values.keys()
                ns_vars.append(
                    {
                        'id': _var['id'],
                        'key': _var['key'],
                        'name': _var['name'],
                        'value': _ns_values.get(ns_id) if ns_id in _ns_value_ids else _var['default_value'],
                    }
                )
            i['ns_vars'] = ns_vars

            if i['cluster_id'] in cluster_dict:
                i['cluster_name'] = cluster_dict[i['cluster_id']]['name']
                i['environment'] = cluster_dict[i['cluster_id']]['environment']
            else:
                i['cluster_name'] = i['cluster_id']
                i['environment'] = None

        if cluster_id:
            results = filter(lambda x: x['cluster_id'] == cluster_id, results)

        if group_by and group_by in valid_group_by:
            # 分组, 排序
            results = [
                {'name': k, 'results': sorted(list(v), key=lambda x: x['id'], reverse=True)}
                for k, v in groupby(sorted(results, key=lambda x: x[group_by]), key=lambda x: x[group_by])
            ]
            if group_by == 'env_type':
                ordering = [i.value for i in EnvType]
                results = sorted(results, key=lambda x: ordering.index(x['name']))
            else:
                results = sorted(results, key=lambda x: x['name'], reverse=True)
                # 过滤带有ns的集群id
                cluster_ids_with_ns = []
                # 按集群分组时，添加集群环境信息
                for r in results:
                    r_ns_list = r.get('results') or []
                    r_ns = r_ns_list[0] if r_ns_list else {}
                    r['environment'] = r_ns.get('environment', '')
                    r['environment_name'] = get_cluster_env_name(r['environment'])
                    r["cluster_id"] = r_ns.get("cluster_id")
                    if get_cluster_type(r["cluster_id"]) == ClusterType.SHARED:
                        r["is_shared"] = True
                    cluster_ids_with_ns.append(r_ns.get("cluster_id"))

                # 添加无命名空间集群ID
                results.extend(self.get_clusters_without_ns(cluster_dict, cluster_ids_with_ns))
        else:
            results = sorted(results, key=lambda x: x['id'], reverse=True)

        with_perms = str2bool(request.query_params.get('with_perms', True))
        if not with_perms:
            return response.Response(results)

        namespace_list = []
        for namespace in results:
            namespace['iam_ns_id'] = calc_iam_ns_id(namespace['cluster_id'], namespace['name'])
            namespace_list.append(namespace)

        return PermsResponse(namespace_list, NamespaceRequest(project_id=project_id, cluster_id=cluster_id))

    def create_flow(self, request, project_id, data):
        access_token = request.user.token.access_token
        project_code = request.project.english_name
        ns_name = data['name']
        cluster_id = data['cluster_id']

        # k8s 集群需要调用 bcs api 初始化数据
        self.init_namespace_by_bcs(access_token, project_id, project_code, data)
        has_image_secret = None

        result = paas_cc.create_namespace(
            access_token,
            project_id,
            cluster_id,
            ns_name,
            None,  # description 现在没有用到
            request.user.username,
            data['env_type'],
            has_image_secret,
        )
        if result.get('code') != 0:
            if 'Duplicate entry' in result.get('message', ''):
                message = _("创建失败，namespace名称已经在其他项目存在")
            else:
                message = result.get('message', '')
            return response.Response({'code': result['code'], 'data': None, 'message': message})
        else:
            self.iam_perm.grant_resource_creator_actions(
                NamespaceCreatorAction(
                    project_id=project_id, cluster_id=cluster_id, creator=request.user.username, name=ns_name
                ),
            )

        # 创建成功后需要保存变量信息
        result_data = result.get('data')
        if data.get('ns_vars') and result_data:
            ns_id = result_data.get('id')
            res, not_exist_vars = NameSpaceVariable.batch_save(ns_id, data['ns_vars'])
            if not_exist_vars:
                not_exist_show_msg = [f'{i["key"]}[id:{i["id"]}]' for i in not_exist_vars]
                result['message'] = _("以下变量不存在:{}").format(';'.join(not_exist_show_msg))
            result['data']['ns_vars'] = NameSpaceVariable.get_ns_vars(ns_id, project_id)
        return result

    @log_audit_on_view(NamespaceAuditor, activity_type=ActivityType.Add)
    def create(self, request, project_id, is_validate_perm=True):
        """新建命名空间
        k8s 流程：新建namespace配置文件并下发 -> 新建包含仓库账号信息的sercret配置文件并下发 -> 在paas-cc上注册
        """
        serializer = slz.CreateNamespaceSLZ(data=request.data, context={'request': request, 'project_id': project_id})
        serializer.is_valid(raise_exception=True)

        data = serializer.data

        # 判断权限
        cluster_id = data['cluster_id']
        perm_ctx = NamespacePermCtx(username=request.user.username, project_id=project_id, cluster_id=cluster_id)
        self.iam_perm.can_create(perm_ctx)

        if get_cluster_type(cluster_id) == ClusterType.SHARED:
            data["name"] = f"{settings.SHARED_CLUSTER_NS_PREFIX}{request.project.project_code}-{data['name']}"

        request.audit_ctx.update_fields(
            resource=data['name'], description=_('集群: {}, 创建命名空间: 命名空间[{}]').format(cluster_id, data["name"])
        )
        result = self.create_flow(request, project_id, data)

        return response.Response(result)

    @log_audit_on_view(NamespaceAuditor, activity_type=ActivityType.Modify)
    def update(self, request, project_id, namespace_id, is_validate_perm=True):
        """修改命名空间下变量"""
        # TODO 增加权限控制
        serializer = slz.UpdateNSVariableSLZ(
            data=request.data, context={'request': request, 'project_id': project_id, 'ns_id': namespace_id}
        )
        serializer.is_valid(raise_exception=True)
        data = serializer.validated_data

        result = {'code': 0, 'data': data, 'message': _("更新成功")}
        # 更新成功后需要保存变量信息
        if data.get('ns_vars'):
            res, not_exist_vars = NameSpaceVariable.batch_save(namespace_id, data['ns_vars'])
            if not_exist_vars:
                not_exist_show_msg = ['%s[id:%s]' % (i['key'], i['id']) for i in not_exist_vars]
                result['message'] = _("以下变量不存在:{}").format(";".join(not_exist_show_msg))
            result['data']['ns_vars'] = NameSpaceVariable.get_ns_vars(namespace_id, project_id)
        request.audit_ctx.update_fields(
            resource=namespace_id, description=_("调整命名空间({})下的变量").format(namespace_id, extra=data)
        )
        return response.Response(result)

    @log_audit_on_view(NamespaceAuditor, activity_type=ActivityType.Delete)
    def delete(self, request, project_id, namespace_id, is_validate_perm=True):
        access_token = request.user.token.access_token

        # TODO 重构
        # namespace exist
        ns_resp = paas_cc.get_namespace(access_token, project_id, namespace_id)
        ns_data = ns_resp['data']
        if ns_resp.get('code') != ErrorCode.NoError or not ns_data:
            raise error_codes.APIError(f'query namespace exist error, {ns_resp.get("message")}')
        # get namespace info
        cluster_id, ns_name = ns_data.get('cluster_id'), ns_data.get('name')

        perm_ctx = NamespacePermCtx(
            username=request.user.username, project_id=project_id, cluster_id=cluster_id, name=ns_name
        )
        self.iam_perm.can_delete(perm_ctx)

        client = Namespace(access_token, project_id, request.project.kind)
        resp = client.delete(namespace_id, cluster_id=cluster_id, ns_name=ns_name)

        request.audit_ctx.update_fields(resource=f"{cluster_id}/{ns_name}", description=_("删除命名空间"))
        return response.Response(resp)

    @log_audit_on_view(NamespaceAuditor, activity_type=ActivityType.Modify)
    def sync_namespace(self, request, project_id):
        """同步命名空间
        用来同步线上和本地存储的数据，并进行secret等的处理，保证数据一致
        """
        resp = paas_cc.get_all_clusters(request.user.token.access_token, project_id, desire_all_data=1)
        if resp.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(f'get cluster error {resp.get("message")}')
        data = resp.get('data') or {}
        results = data.get('results')
        if not results:
            raise error_codes.ResNotFoundError(f'not found cluster in project: {project_id}')

        # 共享集群的命名空间只能通过产品创建，不允许同步
        cluster_id_list = [
            info['cluster_id'] for info in results if get_cluster_type(info['cluster_id']) != ClusterType.SHARED
        ]
        # 触发后台任务进行同步数据
        sync_ns_task.delay(
            request.user.token.access_token,
            project_id,
            request.project.project_code,
            request.project.kind,
            cluster_id_list,
            request.user.username,
        )
        request.audit_ctx.update_fields(description=_("同步项目下所有集群的命名空间"))
        return response.Response({'code': 0, 'message': 'task is running'})


class NamespaceQuotaViewSet(viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def _ns_quota_client(self, access_token, project_id, cluster_id):
        return ns_resource.NamespaceQuota(access_token=access_token, project_id=project_id, cluster_id=cluster_id)

    def get_namespace_quota(self, request, project_id, cluster_id, namespace):
        """获取命名空间信息"""
        # TODO: 权限需要梳理，先不添加权限限制
        # 获取配额
        client = self._ns_quota_client(request.user.token.access_token, project_id, cluster_id)
        quota = client.get_namespace_quota(namespace)
        ns = {"project_id": project_id, "cluster_id": cluster_id, "name": namespace, "quota": quota}
        return response.Response(ns)

    @log_audit_on_view(NamespaceQuotaAuditor, activity_type=ActivityType.Modify)
    def update_namespace_quota(self, request, project_id, cluster_id, namespace):
        """更新命名空间下的资源配额"""
        serializer = slz.UpdateNamespaceQuotaSLZ(data=request.data)
        serializer.is_valid(raise_exception=True)
        data = serializer.validated_data

        client = self._ns_quota_client(request.user.token.access_token, project_id, cluster_id)
        client.update_or_create_namespace_quota(namespace, data["quota"])
        request.audit_ctx.update_fields(
            resource=f"{cluster_id}/{namespace}", description=_("更新命名空间下的资源配额"), extra=data["quota"]
        )
        return response.Response()

    @log_audit_on_view(NamespaceQuotaAuditor, activity_type=ActivityType.Delete)
    def delete_namespace_quota(self, request, project_id, cluster_id, namespace):
        """删除命名空间资源配额"""
        client = self._ns_quota_client(request.user.token.access_token, project_id, cluster_id)
        client.delete_namespace_quota(namespace)
        request.audit_ctx.update_fields(resource=f"{cluster_id}/{namespace}", description=_("删除命名空间下的资源配额"))
        return response.Response()
