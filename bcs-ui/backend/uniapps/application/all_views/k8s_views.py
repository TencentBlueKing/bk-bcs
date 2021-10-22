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
import json
import logging
from datetime import datetime

from django.utils.translation import ugettext_lazy as _

from backend.celery_app.tasks.application import update_create_error_record
from backend.components import paas_cc
from backend.components.bcs.k8s import K8SClient
from backend.kube_core.hpa.utils import get_deployment_hpa
from backend.templatesets.legacy_apps.configuration.models import Template
from backend.templatesets.legacy_apps.instance.constants import InsState
from backend.templatesets.legacy_apps.instance.models import (
    InstanceConfig,
    InstanceEvent,
    MetricConfig,
    VersionInstance,
)
from backend.uniapps.application import constants as app_constants
from backend.uniapps.application import utils
from backend.uniapps.application.constants import (
    CATEGORY_MAP,
    DELETE_INSTANCE,
    FUNC_MAP,
    REVERSE_CATEGORY_MAP,
    SOURCE_TYPE_MAP,
    UNNORMAL_STATUS,
)
from backend.uniapps.application.utils import get_instance_version, get_instance_version_name, retry_requests
from backend.utils.basic import getitems
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes

logger = logging.getLogger(__name__)


class GetNamespace(object):
    def get_namespaces(self, request, project_id):
        resp = paas_cc.get_namespace_list(request.user.token.access_token, project_id, desire_all_data=True)
        if resp.get("code") != ErrorCode.NoError:
            raise error_codes.APIError.f(resp.get("message"))
        return resp.get("data", {}).get("results") or []

    def get_ns_app_info(self, ns_id_list, category, inst_name, project_id, request):
        """获取命名空间下的应用列表"""
        all_info = (
            InstanceConfig.objects.filter(namespace__in=ns_id_list, is_deleted=False, category=category)
            .exclude(ins_state=InsState.NO_INS.value)
            .order_by("-updated")
        )
        if inst_name:
            all_info = all_info.filter(name=inst_name)
        ret_data = {}
        ns_inst = {}
        create_error = {}
        ns_name_inst = []
        all_namespaces = self.get_namespaces(request, project_id)
        ns_map = {info["id"]: info for info in all_namespaces}
        for info in all_info:
            if info.namespace in ret_data:
                ret_data[info.namespace] += 1
            else:
                ret_data[info.namespace] = 1
            key = info.namespace
            ns_name = ns_map.get(int(key), {}).get("name")
            if int(key) in ns_inst:
                ns_inst[int(key)].append(info.name)
            else:
                ns_inst[int(key)] = [info.name]
            if not info.is_bcs_success:
                if int(key) in create_error:
                    create_error[int(key)] += 1
                else:
                    create_error[int(key)] = 1
            else:
                instance_config = json.loads(info.config)
                cluster_id = instance_config['metadata']['labels'].get('io.tencent.bcs.clusterid')
                ns_name_inst.append((cluster_id, ns_name, info.name))
        return ret_data, ns_inst, create_error, ns_name_inst

    def get_ns_inst_error_count(
        self, func, request, project_id, category, kind, ns_name_list, inst_name, cluster_id_list, ns_name_inst
    ):
        """获取实例错误数据"""
        error_data = {}
        all_data = {}
        category = REVERSE_CATEGORY_MAP[category]
        for cluster_id in set(cluster_id_list):
            # 要展示client创建的应用，因此不能指定名称和命名空间
            flag, resp = func(
                request,
                project_id,
                cluster_id,
                instance_name=inst_name,
                category=category,
                project_kind=kind,
                namespace=",".join(set(ns_name_list)),
                field=[
                    "data.status",
                    "resourceName",
                    "namespace",
                    "data.spec.parallelism",
                    "data.spec.paused",
                    'data.spec.replicas',
                ],
            )
            if not flag:
                logger.error("请求storage接口出现异常, 详情: %s" % resp)
                continue
            resp_data = resp.get("data") or []
            diff_inst = set(ns_name_inst) - set(
                [(cluster_id, info["namespace"], info["resourceName"]) for info in resp_data]
            )
            for info in resp_data:
                ns_name = info["namespace"]
                spec = (info.get("data") or {}).get("spec") or {}
                # 针对不同的模板获取不同的值
                replicas, available = utils.get_k8s_desired_ready_instance_count(info, category)
                pause_status = spec.get("paused")
                if (cluster_id, ns_name) in all_data:
                    all_data[(cluster_id, ns_name)] += 1
                else:
                    all_data[(cluster_id, ns_name)] = 1
                if (available != replicas or available <= 0) and (not pause_status):
                    if (cluster_id, ns_name) in error_data:
                        error_data[(cluster_id, ns_name)] += 1
                    else:
                        error_data[(cluster_id, ns_name)] = 1
            for info in diff_inst:
                if (cluster_id, info[0]) in all_data:
                    all_data[(cluster_id, info[0])] += 1
                if info[0] in error_data:
                    error_data[(cluster_id, info[0])] += 1
                else:
                    error_data[(cluster_id, info[0])] = 1
        return error_data, all_data

    def get(
        self, request, ns_id_list, category, ns_map, project_id, kind, func, inst_name, ns_name_list, cluster_id_list
    ):
        category = CATEGORY_MAP[category]
        ns_app_info, ns_inst, create_error, ns_name_inst = self.get_ns_app_info(
            ns_id_list, category, inst_name, project_id, request
        )
        ns_inst_error_count, all_ns_inst_count = self.get_ns_inst_error_count(
            func, request, project_id, category, kind, ns_name_list, inst_name, cluster_id_list, ns_name_inst
        )
        return ns_app_info, ns_inst_error_count, create_error, all_ns_inst_count


class GetInstances(object):
    def get_muster_info(self, tmpl_id):
        tmpl_info = Template.objects.filter(id=tmpl_id).first()
        if not tmpl_info:
            return None
        return tmpl_info.name

    def get_insts(self, ns_id, category, inst_name):
        """获取实例"""
        category = CATEGORY_MAP[category]
        all_inst_list = (
            InstanceConfig.objects.filter(namespace=ns_id, category=category, is_deleted=False)
            .exclude(ins_state=InsState.NO_INS.value)
            .order_by("-updated")
        )
        if inst_name:
            all_inst_list = all_inst_list.filter(name=inst_name)
        return all_inst_list

    def compose_inst_info(self, all_inst_list, cluster_env_map):
        """组装实例信息"""
        ret_data = {}
        for info in all_inst_list:
            conf = json.loads(info.config)
            metadata = conf.get("metadata", {})
            key_name = (metadata.get("namespace"), metadata.get("name"))
            labels = metadata.get("labels")
            backend_status = "BackendNormal"
            oper_type_flag = ""
            if not info.is_bcs_success:
                if info.oper_type == "create":
                    backend_status = "BackendError"
                else:
                    oper_type_flag = info.oper_type
            cluster_id = labels.get("io.tencent.bcs.clusterid")
            muster_id = labels.get("io.tencent.paas.templateid")
            cluster_name_env_map = cluster_env_map.get(cluster_id) or {}
            ret_data[key_name] = {
                "from_platform": True,
                "id": info.id,
                "name": metadata.get("name"),
                "namespace": metadata.get("namespace"),
                "namespace_id": info.namespace,
                "create_at": info.created,
                "update_at": info.updated,
                "backend_status": backend_status,
                "backend_status_message": _("请求失败，已通知管理员!"),
                "status": "Unready",
                "status_message": _("请点击查看详情"),
                "creator": info.creator,
                "category": REVERSE_CATEGORY_MAP[info.category],
                "oper_type": info.oper_type,
                "oper_type_flag": oper_type_flag,
                "cluster_id": cluster_id,
                "pod_count": "0/0",
                "build_instance": 0,
                "instance": 0,
                "muster_id": muster_id,
                "muster_name": self.get_muster_info(muster_id),
                "cluster_name": cluster_name_env_map.get("cluster_name"),
                "cluster_env": cluster_name_env_map.get("cluster_env"),
                "environment": cluster_name_env_map.get("cluster_env_str"),
            }
            annotations = metadata.get('annotations') or {}
            ret_data[key_name].update(get_instance_version(annotations, labels))
        return ret_data

    def get_cluster_ns_inst(self, instance_info):
        """获取集群、名称、命名空间"""
        ret_data = {}
        for key, val in instance_info.items():
            cluster_id = val["cluster_id"]
            if cluster_id in ret_data:
                ret_data[cluster_id]["inst_list"].append(key[1])
                ret_data[cluster_id]["ns_list"].append(key[0])
                ret_data[cluster_id]["ns_inst_map"][key[1]] = key[0]
            else:
                ret_data[cluster_id] = {"inst_list": [key[1]], "ns_list": [key[0]], "ns_inst_map": {key[1]: key[0]}}
        return ret_data

    def get_k8s_category_info(self, request, project_id, resource_name, inst_name, cluster_id, ns_name):
        """获取分类的上报信息
        {'BCS-K8S-15007': {'K8sDeployment': {'inst_list': ['bellke-test-deploy-1'], 'ns_list': ['abc1']}}}
        """
        ret_data = {}
        client = K8SClient(request.user.token.access_token, project_id, cluster_id, None)
        curr_func = FUNC_MAP[resource_name] % 'get'
        resp = retry_requests(
            getattr(client, curr_func),
            params={
                "name": inst_name,
                "namespace": ns_name,
                "field": ','.join(app_constants.RESOURCE_STATUS_FIELD_LIST),
            },
        )
        if resp.get('code') != ErrorCode.NoError:
            raise error_codes.APIError.f(resp.get('message'))
        data = resp.get('data') or []

        # 添加HPA绑定信息
        data = get_deployment_hpa(request, project_id, cluster_id, ns_name, data)

        for info in data:
            spec = getitems(info, ['data', 'spec'], default={})
            # 针对不同的模板获取不同的值
            replicas, available = utils.get_k8s_desired_ready_instance_count(info, resource_name)
            curr_key = (info['namespace'], info['resourceName'])
            labels = getitems(info, ['data', 'metadata', 'labels'], default={})
            source_type = labels.get('io.tencent.paas.source_type') or 'other'
            annotations = getitems(info, ['data', 'metadata', 'annotations'], default={})
            ret_data[curr_key] = {
                'backend_status': 'BackendNormal',
                'backend_status_message': _('请求失败，已通知管理员!'),
                'category': resource_name,
                'pod_count': f'{available}/{replicas}',
                'build_instance': available,
                'instance': replicas,
                'status': utils.get_k8s_resource_status(resource_name, info, replicas, available),
                'name': info['resourceName'],
                'namespace': info['namespace'],
                'create_at': info['createTime'],
                'update_at': info['updateTime'],
                'source_type': SOURCE_TYPE_MAP.get(source_type),
                'version': get_instance_version_name(annotations, labels),  # 标识应用的线上版本
                'hpa': info['hpa'],  # 是否绑定了HPA
            }
            if spec.get('paused'):
                ret_data[curr_key]['status'] = 'Paused'
        return ret_data

    def compose_data(self, instance_info, inst_status_info, cluster_env_map, cluster_id, ns_id, ns_name_id):
        """组装数据"""
        # need_delete_id_list = []
        update_create_error_id_list = []
        for key, val in inst_status_info.items():
            if key in instance_info:
                if instance_info[key].get("backend_status") in ["BackendError"]:
                    instance_info[key]["backend_status"] = "BackendNormal"
                    update_create_error_id_list.append(instance_info[key]["id"])
                val.pop("version", None)
                # 针对模板集创建的应用更新时间以数据库为准，否则，从api中直接获取
                val.pop('update_at', None)
                instance_info[key].update(val)
            else:
                val['namespace_id'] = ns_name_id.get(val.get('namespace'))
                val["id"] = 0
                val["from_platform"] = False
                val["oper_type"] = "create"
                instance_info[key] = val
            cluster_name_env_map = cluster_env_map.get(cluster_id) or {}
            instance_info[key].update(
                {
                    "namespace_id": ns_id,
                    "cluster_id": cluster_id,
                    "cluster_name": cluster_name_env_map.get("cluster_name"),
                    "cluster_env": cluster_name_env_map.get("cluster_env"),
                    "environment": cluster_name_env_map.get("cluster_env_str"),
                }
            )
        if update_create_error_id_list:
            update_create_error_record.delay(update_create_error_id_list)

        utils.delete_instance_records(inst_status_info, instance_info)

    def inst_count_handler(self, instance_list, app_status):
        ret_data = {
            "error_num": 0,
        }
        instance_list = list(instance_list)
        inst_list_copy = copy.deepcopy(instance_list)
        for val in inst_list_copy:
            if (val["backend_status"] in UNNORMAL_STATUS) or (val["status"] in UNNORMAL_STATUS):
                if app_status in [2, "2", None]:
                    ret_data["error_num"] += 1
                else:
                    instance_list.remove(val)
            else:
                if app_status not in [1, "1", None]:
                    instance_list.remove(val)
        ret_data["total_num"] = len(instance_list)
        ret_data["instance_list"] = instance_list
        return ret_data

    def get(
        self,
        request,
        project_id,
        ns_id,
        category,
        inst_name,
        app_status,
        cluster_env_map,
        cluster_id,
        ns_name,
        ns_name_id,
    ):
        """获取命名空间下的实例"""
        all_inst_list = self.get_insts(ns_id, category, inst_name)
        # 组装查询版本数据
        instance_info = self.compose_inst_info(all_inst_list, cluster_env_map)
        # 进行匹配
        inst_status_info = self.get_k8s_category_info(request, project_id, category, inst_name, cluster_id, ns_name)
        self.compose_data(instance_info, inst_status_info, cluster_env_map, cluster_id, ns_id, ns_name_id)
        ret_data = self.inst_count_handler(instance_info.values(), app_status)
        return ret_data
