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
import copy
import datetime
import json
import logging

from django.utils.translation import ugettext_lazy as _
from rest_framework import viewsets
from rest_framework.exceptions import ValidationError
from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.response import Response

from backend.bcs_web.audit_log import client as activity_client
from backend.components.bcs import k8s
from backend.container_service.clusters.base.utils import get_cluster_type
from backend.container_service.clusters.constants import ClusterType
from backend.iam.permissions.decorators import response_perms
from backend.iam.permissions.resources.namespace import NamespaceRequest, calc_iam_ns_id
from backend.iam.permissions.resources.namespace_scoped import (
    NamespaceScopedAction,
    NamespaceScopedPermCtx,
    NamespaceScopedPermission,
)
from backend.resources.namespace.constants import K8S_PLAT_NAMESPACE, K8S_SYS_NAMESPACE
from backend.templatesets.legacy_apps.configuration.constants import TemplateEditMode
from backend.templatesets.legacy_apps.configuration.models import (
    K8sService,
    Service,
    ShowVersion,
    Template,
    VersionedEntity,
)
from backend.templatesets.legacy_apps.configuration.serializers import K8sServiceCreateOrUpdateSLZ
from backend.templatesets.legacy_apps.instance.constants import (
    ANNOTATIONS_CREATE_TIME,
    ANNOTATIONS_CREATOR,
    ANNOTATIONS_UPDATE_TIME,
    ANNOTATIONS_UPDATOR,
    ANNOTATIONS_WEB_CACHE,
    K8S_SEVICE_SYS_CONFIG,
    LABLE_INSTANCE_ID,
    LABLE_TEMPLATE_ID,
    PUBLIC_ANNOTATIONS,
    PUBLIC_LABELS,
    SOURCE_TYPE_LABEL_KEY,
)
from backend.templatesets.legacy_apps.instance.drivers import get_scheduler_driver
from backend.templatesets.legacy_apps.instance.funutils import render_mako_context, update_nested_dict
from backend.templatesets.legacy_apps.instance.generator import (
    get_bcs_context,
    handel_k8s_service_db_config,
    handle_k8s_api_version,
    handle_webcache_config,
    remove_key,
)
from backend.templatesets.legacy_apps.instance.models import InstanceConfig
from backend.templatesets.legacy_apps.instance.utils_pub import get_cluster_version
from backend.uniapps import utils as app_utils
from backend.uniapps.application.base_views import BaseAPI
from backend.uniapps.application.constants import DELETE_INSTANCE, SOURCE_TYPE_MAP
from backend.uniapps.application.utils import APIResponse
from backend.uniapps.network.ext_routes import delete_svc_extended_routes, get_svc_extended_routes
from backend.uniapps.network.serializers import BatchResourceSLZ
from backend.uniapps.network.utils import get_svc_access_info
from backend.utils.errcodes import ErrorCode
from backend.utils.exceptions import ComponentError
from backend.utils.renderers import BKAPIRenderer
from backend.utils.response import PermsResponse

logger = logging.getLogger(__name__)
DEFAULT_ERROR_CODE = ErrorCode.UnknownError


class Services(viewsets.ViewSet, BaseAPI):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_services_by_cluster_id(self, request, params, project_id, cluster_id):
        """查询services"""
        if get_cluster_type(cluster_id) == ClusterType.SHARED:
            return ErrorCode.NoError, []

        access_token = request.user.token.access_token
        client = k8s.K8SClient(access_token, project_id, cluster_id, env=None)
        resp = client.get_service(params)

        if resp.get("code") != ErrorCode.NoError:
            logger.error(u"bcs_api error: %s" % resp.get("message", ""))
            return resp.get("code", DEFAULT_ERROR_CODE), resp.get("message", _("请求出现异常!"))

        return ErrorCode.NoError, resp.get("data", [])

    def get_service_info(self, request, project_id, cluster_id, namespace, name):  # noqa
        """获取单个 service 的信息"""
        if get_cluster_type(cluster_id) == ClusterType.SHARED:
            return APIResponse({"code": 400, "message": _("无法查看共享集群资源")})

        access_token = request.user.token.access_token
        params = {
            "env": "k8s",
            "namespace": namespace,
            "name": name,
        }
        client = k8s.K8SClient(access_token, project_id, cluster_id, env=None)
        resp = client.get_service(params)
        template_cate = 'k8s'
        relate_app_cate = 'deployment'

        if resp.get("code") != ErrorCode.NoError:
            raise ComponentError(resp.get("message"))

        resp_data = resp.get("data", [])
        if not resp_data:
            return APIResponse({"code": 400, "message": _("查询不到 Service[{}] 的信息").format(name)})
        s_data = resp_data[0].get('data', {})
        labels = s_data.get('metadata', {}).get('labels') or {}

        # 获取命名空间的id
        namespace_id = app_utils.get_namespace_id(
            access_token, project_id, (cluster_id, namespace), cluster_id=cluster_id
        )

        instance_id = labels.get(LABLE_INSTANCE_ID)

        # 是否关联LB
        lb_balance = labels.get('BCSBALANCE')
        if lb_balance:
            s_data['isLinkLoadBalance'] = True
            s_data['metadata']['lb_labels'] = {'BCSBALANCE': lb_balance}
        else:
            s_data['isLinkLoadBalance'] = False
        lb_name = labels.get('BCSGROUP')

        # 获取模板集信息
        template_id = labels.get(LABLE_TEMPLATE_ID)
        try:
            lasetest_ver = ShowVersion.objects.filter(template_id=template_id).order_by('-updated').first()
            show_version_name = lasetest_ver.name
            version_id = lasetest_ver.real_version_id
            version_entity = VersionedEntity.objects.get(id=version_id)
        except Exception:
            return APIResponse({"code": 400, "message": _("模板集[id:{}]没有可用的版本，无法更新service").format(template_id)})

        entity = version_entity.get_entity()

        # 获取更新人和创建人
        annotations = s_data.get('metadata', {}).get('annotations', {})
        creator = annotations.get(ANNOTATIONS_CREATOR, '')
        updator = annotations.get(ANNOTATIONS_UPDATOR, '')
        create_time = annotations.get(ANNOTATIONS_CREATE_TIME, '')
        update_time = annotations.get(ANNOTATIONS_UPDATE_TIME, '')

        # k8s 更新需要获取版本号
        resource_version = s_data.get('metadata', {}).get('resourceVersion') or ''

        web_cache = annotations.get(ANNOTATIONS_WEB_CACHE)
        if not web_cache:
            # 备注中无，则从模板中获取
            _services = entity.get('service') if entity else None
            _services_id_list = _services.split(',') if _services else []
            _s = Service.objects.filter(id__in=_services_id_list, name=name).first()
            try:
                web_cache = _s.get_config.get('webCache')
            except Exception:
                pass
        else:
            try:
                web_cache = json.loads(web_cache)
            except Exception:
                pass
        s_data['webCache'] = web_cache
        deploy_tag_list = web_cache.get('deploy_tag_list') or []

        app_weight = {}
        # 处理 k8s 中Service的关联数据
        if not deploy_tag_list:
            _servs = entity.get('K8sService') if entity else None
            _serv_id_list = _servs.split(',') if _servs else []
            _k8s_s = K8sService.objects.filter(id__in=_serv_id_list, name=name).first()
            if _k8s_s:
                deploy_tag_list = _k8s_s.get_deploy_tag_list()

        # 标签 和 备注 去除后台自动添加的
        or_annotations = s_data.get('metadata', {}).get('annotations', {})
        or_labels = s_data.get('metadata', {}).get('labels', {})
        if or_labels:
            pub_keys = PUBLIC_LABELS.keys()
            show_labels = {key: or_labels[key] for key in or_labels if key not in pub_keys}
            s_data['metadata']['labels'] = show_labels
        if or_annotations:
            pub_an_keys = PUBLIC_ANNOTATIONS.keys()
            show_annotations = {key: or_annotations[key] for key in or_annotations if key not in pub_an_keys}
            remove_key(show_annotations, ANNOTATIONS_WEB_CACHE)
            s_data['metadata']['annotations'] = show_annotations

        return APIResponse(
            {
                "data": {
                    'service': [
                        {
                            'name': name,
                            'app_id': app_weight.keys(),
                            'app_weight': app_weight,
                            'deploy_tag_list': deploy_tag_list,
                            'config': s_data,
                            'version': version_id,
                            'lb_name': lb_name,
                            'instance_id': instance_id,
                            'namespace_id': namespace_id,
                            'cluster_id': cluster_id,
                            'namespace': namespace,
                            'creator': creator,
                            'updator': updator,
                            'create_time': create_time,
                            'update_time': update_time,
                            'show_version_name': show_version_name,
                            'resource_version': resource_version,
                            'template_id': template_id,
                            'template_cate': template_cate,
                            'relate_app_cate': relate_app_cate,
                        }
                    ]
                }
            }
        )

    @response_perms(
        action_ids=[NamespaceScopedAction.DELETE, NamespaceScopedAction.UPDATE, NamespaceScopedAction.VIEW],
        permission_cls=NamespaceScopedPermission,
        resource_id_key='iam_ns_id',
    )
    def get(self, request, project_id):
        """获取项目下所有的服务"""
        params = dict(request.GET.items())
        params['env'] = 'k8s'

        # 获取命名空间的id
        namespace_dict = app_utils.get_ns_id_map(request.user.token.access_token, project_id)

        # 项目下的所有模板集id
        all_template_id_list = Template.objects.filter(
            project_id=project_id, edit_mode=TemplateEditMode.PageForm.value
        ).values_list('id', flat=True)
        all_template_id_list = [str(template_id) for template_id in all_template_id_list]
        skip_namespace_list = list(K8S_SYS_NAMESPACE)
        skip_namespace_list.extend(K8S_PLAT_NAMESPACE)

        cluster_id = params['cluster_id']
        code, cluster_services = self.get_services_by_cluster_id(request, params, project_id, cluster_id)
        if code != ErrorCode.NoError:
            return Response({'code': code, 'message': cluster_services})

        for _s in cluster_services:
            # NOTE: 兼容处理，因为key: clusterId已被前端使用；通过非bcs创建的service，不一定包含cluster_id
            _s["clusterId"] = cluster_id
            _s["cluster_id"] = cluster_id
            _config = _s.get('data', {})
            annotations = _config.get('metadata', {}).get('annotations', {})
            _s['update_time'] = annotations.get(ANNOTATIONS_UPDATE_TIME, '')
            _s['updator'] = annotations.get(ANNOTATIONS_UPDATOR, '')
            _s['status'] = 'Running'

            _s['can_update'] = True
            _s['can_update_msg'] = ''
            _s['can_delete'] = True
            _s['can_delete_msg'] = ''

            namespace_id = namespace_dict.get((cluster_id, _s['namespace'])) if namespace_dict else None
            _s['namespace_id'] = namespace_id
            _s['iam_ns_id'] = calc_iam_ns_id(cluster_id, _s['namespace'])

            labels = _config.get('metadata', {}).get('labels', {})
            template_id = labels.get(LABLE_TEMPLATE_ID)
            # 资源来源
            source_type = labels.get(SOURCE_TYPE_LABEL_KEY)
            if not source_type:
                source_type = "template" if template_id else "other"
            _s['source_type'] = SOURCE_TYPE_MAP.get(source_type)
            extended_routes = get_svc_extended_routes(project_id, _s['clusterId'])
            _s['access_info'] = get_svc_access_info(_config, _s['clusterId'], extended_routes)
            # 处理 k8s 的系统命名空间的数据
            if _s['namespace'] in skip_namespace_list:
                _s['can_update'] = _s['can_delete'] = False
                _s['can_update_msg'] = _s['can_delete_msg'] = _("不允许操作系统命名空间")
                continue

            # 非模板集创建，可以删除但是不可以更新
            _s['can_update'] = False
            _s['can_update_msg'] = _("所属模板集不存在，无法操作")
            if template_id and template_id in all_template_id_list:
                _s['can_update'] = True
                _s['can_update_msg'] = ''

        # 按时间倒序排列
        cluster_services.sort(key=lambda x: x.get('createTime', ''), reverse=True)

        return PermsResponse(
            cluster_services, NamespaceRequest(project_id=project_id, cluster_id=params['cluster_id'])
        )

    def delete_single_service(self, request, project_id, project_kind, cluster_id, namespace, namespace_id, name):
        if get_cluster_type(cluster_id) == ClusterType.SHARED:
            return {"code": 400, "message": _("无法操作共享集群资源")}

        username = request.user.username
        access_token = request.user.token.access_token

        if namespace in K8S_SYS_NAMESPACE:
            return {
                "code": 400,
                "message": _("不允许操作系统命名空间[{}]").format(','.join(K8S_SYS_NAMESPACE)),
            }
        client = k8s.K8SClient(access_token, project_id, cluster_id, env=None)
        resp = client.delete_service(namespace, name)
        s_cate = 'K8sService'

        delete_svc_extended_routes(request, project_id, cluster_id, namespace, name)

        if resp.get("code") == ErrorCode.NoError:
            # 删除成功则更新状态
            now_time = datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')
            InstanceConfig.objects.filter(namespace=namespace_id, category=s_cate, name=name,).update(
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

    def delete_services(self, request, project_id, cluster_id, namespace, name):
        username = request.user.username
        # 检查用户是否有命名空间的使用权限
        namespace_id = app_utils.get_namespace_id(
            request.user.token.access_token, project_id, (cluster_id, namespace), cluster_id=cluster_id
        )
        app_utils.can_use_namespace(request, project_id, cluster_id, namespace)

        flag, project_kind = self.get_project_kind(request, project_id)
        if not flag:
            return project_kind

        # 删除service
        resp = self.delete_single_service(request, project_id, project_kind, cluster_id, namespace, namespace_id, name)
        # 添加操作审计
        activity_client.ContextActivityLogClient(
            project_id=project_id,
            user=username,
            resource_type="instance",
            resource=name,
            resource_id=0,
            extra=json.dumps({}),
            description=_("删除Service[{}]命名空间[{}]").format(name, namespace),
        ).log_modify(activity_status="succeed" if resp.get("code") == ErrorCode.NoError else "failed")

        # 已经删除的，需要将错误信息翻译一下
        message = resp.get('message', '')
        is_delete_before = True if 'node does not exist' in message or 'not found' in message else False
        if is_delete_before:
            message = _("{}[命名空间:{}]已经被删除，请手动刷新数据").format(name, namespace)
        return Response({"code": resp.get("code"), "message": message, "data": {}})

    def batch_delete_services(self, request, project_id):
        """批量删除service"""
        username = request.user.username
        slz = BatchResourceSLZ(data=request.data)
        slz.is_valid(raise_exception=True)
        data = slz.data['data']

        # 检查用户是否有命名空间的使用权限
        namespace_list = [(ns['cluster_id'], ns.get('namespace')) for ns in data]
        namespace_list = set(namespace_list)

        # check perm
        app_utils.can_use_namespaces(request, project_id, namespace_list)

        # namespace_dict format: {(cluster_id, ns_name): ns_id}
        namespace_dict = app_utils.get_ns_id_map(request.user.token.access_token, project_id)

        project_kind = request.project.kind
        success_list = []
        failed_list = []
        for _d in data:
            cluster_id = _d.get('cluster_id')
            name = _d.get('name')
            namespace = _d.get('namespace')
            namespace_id = namespace_dict.get((cluster_id, namespace))
            # 删除service
            resp = self.delete_single_service(
                request, project_id, project_kind, cluster_id, namespace, namespace_id, name
            )
            # 处理已经删除，但是storage上报数据延迟的问题
            message = resp.get('message', '')
            is_delete_before = True if 'node does not exist' in message or 'not found' in message else False
            if resp.get("code") == ErrorCode.NoError:
                success_list.append(
                    {
                        'name': name,
                        'desc': _('{}[命名空间:{}]').format(name, namespace),
                    }
                )
            else:
                if is_delete_before:
                    message = _('已经被删除，请手动刷新数据')
                failed_list.append(
                    {
                        'name': name,
                        'desc': _('{}][命名空间:{}]:{}').format(name, namespace, message),
                    }
                )
        code = 0
        message = ''
        # 添加操作审计
        if success_list:
            name_list = [_s.get('name') for _s in success_list]
            desc_list = [_s.get('desc') for _s in success_list]
            message = _("以下service删除成功:{}").format(";".join(desc_list))
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
            message = _("以下service删除失败:{}").format(";".join(desc_list))
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

    def update_services(self, request, project_id, cluster_id, namespace, name):
        """更新 service"""
        if get_cluster_type(cluster_id) == ClusterType.SHARED:
            return Response({"code": 400, "message": _("无法操作共享集群资源")})

        access_token = request.user.token.access_token
        flag, project_kind = self.get_project_kind(request, project_id)
        if not flag:
            return project_kind

        if namespace in K8S_SYS_NAMESPACE:
            return Response(
                {"code": 400, "message": _("不允许操作系统命名空间[{}]").format(','.join(K8S_SYS_NAMESPACE)), "data": {}}
            )
        # k8s 相关数据
        slz_class = K8sServiceCreateOrUpdateSLZ
        s_sys_con = K8S_SEVICE_SYS_CONFIG
        s_cate = 'K8sService'

        request_data = request.data or {}
        request_data['version_id'] = request_data['version']
        request_data['item_id'] = 0
        request_data['project_id'] = project_id
        show_version_name = request_data.get('show_version_name', '')
        # 验证请求参数
        slz = slz_class(data=request.data)
        slz.is_valid(raise_exception=True)
        data = slz.data
        namespace_id = data['namespace_id']

        # 检查是否有命名空间的使用权限
        perm_ctx = NamespaceScopedPermCtx(
            username=request.user.username, project_id=project_id, cluster_id=cluster_id, name=namespace
        )
        NamespaceScopedPermission().can_use(perm_ctx)

        config = json.loads(data['config'])
        #  获取关联的应用列表
        version_id = data['version_id']
        version_entity = VersionedEntity.objects.get(id=version_id)

        logger.exception(f"deploy_tag_list {type(data['deploy_tag_list'])}")
        handel_k8s_service_db_config(config, data['deploy_tag_list'], version_id, is_upadte=True)
        resource_version = data['resource_version']
        config['metadata']['resourceVersion'] = resource_version
        cluster_version = get_cluster_version(access_token, project_id, cluster_id)
        config = handle_k8s_api_version(config, cluster_id, cluster_version, 'Service')
        # 前端的缓存数据储存到备注中
        config = handle_webcache_config(config)

        # 获取上下文信息
        username = request.user.username
        now_time = datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')
        context = {
            'SYS_CLUSTER_ID': cluster_id,
            'SYS_NAMESPACE': namespace,
            'SYS_VERSION_ID': version_id,
            'SYS_PROJECT_ID': project_id,
            'SYS_OPERATOR': username,
            'SYS_TEMPLATE_ID': version_entity.template_id,
            'SYS_VERSION': show_version_name,
            'LABLE_VERSION': show_version_name,
            'SYS_INSTANCE_ID': data['instance_id'],
            'SYS_CREATOR': data.get('creator', ''),
            'SYS_CREATE_TIME': data.get('create_time', ''),
            'SYS_UPDATOR': username,
            'SYS_UPDATE_TIME': now_time,
        }
        bcs_context = get_bcs_context(access_token, project_id)
        context.update(bcs_context)

        # 生成配置文件
        sys_config = copy.deepcopy(s_sys_con)
        resource_config = update_nested_dict(config, sys_config)
        resource_config = json.dumps(resource_config)
        try:
            config_profile = render_mako_context(resource_config, context)
        except Exception:
            logger.exception(u"配置文件变量替换出错\nconfig:%s\ncontext:%s" % (resource_config, context))
            raise ValidationError(_("配置文件中有未替换的变量"))

        service_name = config.get('metadata', {}).get('name')
        _config_content = {'name': service_name, 'config': json.loads(config_profile), 'context': context}

        # 更新 service
        config_objs = InstanceConfig.objects.filter(
            namespace=namespace_id,
            category=s_cate,
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
                category=s_cate,
                name=service_name,
                config=config_profile,
                instance_id=data.get('instance_id', 0),
                creator=username,
                updator=username,
                oper_type='update',
                updated=now_time,
                is_deleted=False,
            )
        _config_content['instance_config_id'] = _instance_config.id
        configuration = {namespace_id: {s_cate: [_config_content]}}

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
            description=_("更新Service[{}]命名空间[{}]").format(service_name, namespace),
        ).log_modify(activity_status="failed" if failed else "succeed")

        if failed:
            return Response(
                {
                    "code": 400,
                    "message": _("Service[{}]在命名空间[{}]更新失败，请联系集群管理员解决").format(service_name, namespace),
                    "data": {},
                }
            )
        return Response({"code": 0, "message": "OK", "data": {}})
