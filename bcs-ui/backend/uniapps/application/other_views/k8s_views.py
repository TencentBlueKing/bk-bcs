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

针对k8s的应用列表部分
TODO: 状态流转需要明确
"""
import json
import logging
from datetime import datetime

from django.utils.translation import ugettext_lazy as _

from backend.celery_app.tasks.application import update_create_error_record
from backend.components.bcs.k8s import K8SClient
from backend.templatesets.legacy_apps.configuration.models import MODULE_DICT
from backend.templatesets.legacy_apps.instance.constants import InsState
from backend.templatesets.legacy_apps.instance.models import InstanceConfig, VersionInstance
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes

from .. import constants as app_constants
from .. import utils

logger = logging.getLogger(__name__)

CATEGORY_MAP = app_constants.CATEGORY_MAP
REVERSE_CATEGORY_MAP = app_constants.REVERSE_CATEGORY_MAP
FUNC_MAP = app_constants.FUNC_MAP


class K8SMuster(object):
    def get_version_instance(
        self, muster_id_list, category, cluster_type, app_name, ns_id, cluster_env_map, request_cluster_id
    ):
        """查询version instance"""
        version_inst_info = (
            VersionInstance.objects.filter(template_id__in=muster_id_list, is_deleted=False)
            .values("id", "template_id", "version_id", "instance_entity")
            .order_by("-updated")
        )
        instance_id_list = [info["id"] for info in version_inst_info]
        # 过滤数据
        instance_info = InstanceConfig.objects.filter(
            instance_id__in=instance_id_list,
            is_deleted=False,
            category__in=[CATEGORY_MAP[category]],
        ).exclude(ins_state=InsState.NO_INS.value)
        if app_name:
            instance_info = instance_info.filter(name=app_name)
        if ns_id:
            instance_info = instance_info.filter(namespace=ns_id)
        instance_info = instance_info.values("instance_id", "name", "category", "config", "id", "namespace")
        exist_inst_id_list = [info["instance_id"] for info in instance_info]
        # 数据匹配
        ret_data = {}
        for info in version_inst_info:
            entity = json.loads(info["instance_entity"])
            if info["id"] not in exist_inst_id_list:
                continue
            category_id_list = entity.get(CATEGORY_MAP[category])
            if info["template_id"] not in ret_data:
                ret_data[info["template_id"]] = {
                    "id_list": [info["id"]],
                    "inst_num": 0,
                    "app_name_list": set([]),
                    "tmpl_id_list": category_id_list,
                }
            else:
                ret_data[info["template_id"]]["id_list"].append(info["id"])
                ret_data[info["template_id"]]["tmpl_id_list"].extend(category_id_list)
        for info in instance_info:
            config = json.loads(info["config"])
            cluster_id = ((config.get("metadata") or {}).get("labels") or {}).get("io.tencent.bcs.clusterid")
            muster_id = int(((config.get("metadata") or {}).get("labels") or {}).get("io.tencent.paas.templateid"))
            # 判断是否忽略当前记录
            if utils.exclude_records(
                request_cluster_id,
                cluster_id,
                cluster_type,
                cluster_env_map.get(cluster_id, {}).get("cluster_env"),
            ):
                continue
            if info["instance_id"] in ret_data[muster_id]["id_list"]:
                ret_data[muster_id]["inst_num"] += 1
                ret_data[muster_id]["app_name_list"] = ret_data[muster_id]["app_name_list"].union(
                    set([info.get("name")])
                )

        return ret_data

    def get_category_name(self, ids, category):
        """获取category模板名称"""
        resp = MODULE_DICT[CATEGORY_MAP[category]].objects.filter(id__in=ids, is_deleted=False)
        ret_name_list = set([info.name for info in resp])
        return ret_name_list

    def muster_tmpl_handler(self, muster_id_name_map, muster_num_map):
        ret_data = []
        for muster_id, info in muster_num_map.items():
            tmpl_num = len(info.get("app_name_list") or [])
            if tmpl_num == 0:
                continue
            ret_data.append(
                {
                    "tmpl_muster_name": muster_id_name_map.get(muster_id, ""),
                    "tmpl_muster_id": muster_id,
                    "tmpl_num": tmpl_num,
                    "inst_num": (muster_num_map.get(muster_id) or {}).get("inst_num", 0),
                }
            )
        return ret_data

    def get(
        self,
        request,
        project_id,
        all_muster_list,
        muster_id_list,
        category,
        cluster_type,
        app_status,
        app_name,
        ns_id,
        cluster_env_map,
        request_cluster_id,
    ):
        if category not in CATEGORY_MAP:
            raise error_codes.CheckFailed(_("类型不正确，请确认"))
        # 获取模板ID和名称的对应关系
        muster_id_name_map = {info["id"]: info["name"] for info in all_muster_list}
        # 获取version instance，用于展示模板集下是否有实例
        muster_num_map = self.get_version_instance(
            muster_id_list, category, cluster_type, app_name, ns_id, cluster_env_map, request_cluster_id
        )
        return self.muster_tmpl_handler(muster_id_name_map, muster_num_map)


class RetriveFilterFields:
    @property
    def filter_fields(self):
        field_list = [
            'data.status',
            'resourceName',
            'namespace',
            'data.spec.parallelism',
            'data.spec.paused',
            'data.spec.replicas',
        ]
        return ','.join(field_list)


class GetMusterTemplate(RetriveFilterFields):
    """针对k8s获取模板"""

    def get_k8s_category_info(self, request, project_id, cluster_ns_inst, category):
        """获取分类的上报信息
        添加类型，只是为了mesos和k8s进行适配
        {'BCS-K8S-15007': {'K8sDeployment': {'inst_list': ['bellke-test-deploy-1'],
        'ns_list': ['abc1'], 'inst_ns_map': {'test1': 'deployment-232132132'}}}}
        """
        resource_name = CATEGORY_MAP[category]
        ret_data = {}
        for cluster_id, info in cluster_ns_inst.items():
            client = K8SClient(request.user.token.access_token, project_id, cluster_id, None)
            if not info.get(resource_name):
                continue
            curr_func = FUNC_MAP[category] % 'get'
            resp = getattr(client, curr_func)(
                {
                    'name': ','.join(info[resource_name]['inst_list']),
                    'namespace': ','.join(info[resource_name]['ns_list']),
                    'field': ','.join(app_constants.RESOURCE_STATUS_FIELD_LIST),
                }
            )
            if resp.get('code') != ErrorCode.NoError:
                raise error_codes.APIError(resp.get('message'))
            data = resp.get('data') or []
            inst_ns_map = info[resource_name]['inst_ns_map']
            exist_item = []
            for info_item in data:
                exist_item.append(info_item['resourceName'])
                replicas, available = utils.get_k8s_desired_ready_instance_count(info_item, category)
                if available != replicas or available == 0:
                    curr_key = f'{resource_name}:{info_item["resourceName"]}'
                    if curr_key in ret_data:
                        ret_data[curr_key] += 1
                    else:
                        ret_data[curr_key] = 1
            for key, val in inst_ns_map.items():
                if val in exist_item:
                    continue
                curr_key = f'{resource_name}:{val}'
                if curr_key in ret_data:
                    ret_data[curr_key] += 1
                else:
                    ret_data[curr_key] = 1
        return ret_data

    def get_category_map(self, category_ids, category):
        """获取deployment信息"""
        category_info = MODULE_DICT[CATEGORY_MAP[category]].objects.filter(id__in=category_ids).order_by("-updated")
        return {info.id: info.name for info in category_info}

    def compose_status_count_data(self, muster_tmpl_map, tmpl_create_error, inst_status):
        """组装数量"""
        ret_data = {}
        for key, val in muster_tmpl_map.items():
            ret_data[key] = {"total_num": val, "error_num": 0}
            if key in tmpl_create_error:
                ret_data[key]["error_num"] += tmpl_create_error[key]

            if key in inst_status:
                ret_data[key]["error_num"] += inst_status[key]
        return ret_data

    def compose_ret_data(self, version_tmpl_muster, version_map, tmpl_count_info, category, app_status):
        ret_data = {}
        exist_key = []
        for info in version_tmpl_muster:
            # 获取category信息
            category_detail = self.get_category_map(info["other_list"], category)
            item = {
                "tmpl_muster_id": info["tmpl_muster_id"],
                "tmpl_muster_name": info["tmpl_muster_name"],
                "version": info["version"],
                "version_id": info["version_id"],
                "last_version": info["last_version"],
                "last_version_id": info["last_version_id"],
                "total_num": 0,
                "error_num": 0,
                "allow_edit": True,
            }
            for key, val in category_detail.items():
                curr_key = "%s:%s" % (CATEGORY_MAP[category], val)
                num_info = tmpl_count_info.get(curr_key) or {}
                item_copy = item.copy()
                if version_map.get(info["version_id"]) and not num_info.get("total_num"):
                    continue
                if curr_key not in exist_key:
                    exist_key.append(curr_key)
                    item_copy["category"] = CATEGORY_MAP[category].split("K8s")[-1]
                    item_copy["tmpl_app_id"] = key
                    item_copy["tmpl_app_name"] = val
                    item_copy["total_num"] = num_info.get("total_num") or 0
                    item_copy["error_num"] = num_info.get("error_num") or 0
                    if app_status in [2, "2", None]:
                        if item_copy["error_num"]:
                            ret_data[curr_key] = item_copy
                    if app_status in [1, "1", None]:
                        if not item_copy["error_num"]:
                            ret_data[curr_key] = item_copy
                if curr_key in ret_data:
                    if ret_data[curr_key]["version_id"] < item_copy["version_id"]:
                        ret_data[curr_key].update(
                            {
                                "version": item_copy["version"],
                                "version_id": item_copy["version_id"],
                                "last_version": item_copy["last_version"],
                                "last_version_id": item_copy["last_version_id"],
                            }
                        )
                    else:
                        if version_map.get(item_copy["version_id"]):
                            ret_data[curr_key]["allow_edit"] = False

        return ret_data

    def get(
        self,
        request,
        cluster_ns_inst,
        project_id,
        kind,
        version_tmpl_muster,
        version_map,
        category,
        muster_tmpl_map,
        tmpl_create_error,
        cluster_type,
        app_status,
        app_name,
        ns_id,
        cluster_env_map,
    ):
        inst_status = self.get_k8s_category_info(request, project_id, cluster_ns_inst, category)
        # 组装状态数量
        inst_status = self.compose_status_count_data(muster_tmpl_map, tmpl_create_error, inst_status)
        # 匹配状态
        ret_data = self.compose_ret_data(version_tmpl_muster, version_map, inst_status, category, app_status)
        return ret_data


class AppInstance(RetriveFilterFields):
    def get_cluster_ns_inst(self, instance_info):
        ret_data = {}
        for key, val in instance_info.items():
            cluster_id = val["cluster_id"]
            # key： (cluster_id, namespace, resource_name)
            app_id = f'{key[1]}:{key[2]}'
            if cluster_id in ret_data:
                ret_data[cluster_id].append(app_id)
            else:
                ret_data[cluster_id] = [app_id]
        return ret_data

    def get_k8s_category_info(self, request, project_id, cluster_ns_inst, category):
        """获取分类的上报信息
        {'BCS-K8S-15007': {'K8sDeployment': {'inst_list': ['bellke-test-deploy-1'], 'ns_list': ['abc1']}}}
        """
        ret_data = {}
        for cluster_id, app_info in cluster_ns_inst.items():
            client = K8SClient(request.user.token.access_token, project_id, cluster_id, None)
            curr_func = FUNC_MAP[category] % 'get'
            resp = getattr(client, curr_func)({'field': ','.join(app_constants.RESOURCE_STATUS_FIELD_LIST)})
            if resp.get('code') != ErrorCode.NoError:
                raise error_codes.APIError.f(resp.get('message'))
            data = resp.get('data') or []
            # TODO: 关于状态匹配可以再根据实际情况进行调整
            for info in data:
                curr_app_id = f'{info["namespace"]}:{info["resourceName"]}'
                if curr_app_id not in app_info:
                    continue
                spec = (info.get('data') or {}).get('spec') or {}
                # 针对不同的模板获取不同的值
                replicas, available = utils.get_k8s_desired_ready_instance_count(info, category)
                curr_key = (cluster_id, info['namespace'], info['resourceName'])
                ret_data[curr_key] = {
                    'pod_count': f'{available}/{replicas}',
                    'build_instance': available,
                    'instance': replicas,
                    'status': utils.get_k8s_resource_status(category, info, replicas, available),
                }
                if spec.get('paused'):
                    ret_data[curr_key]['status'] = 'Paused'
        return ret_data

    def compose_data(self, instance_info, inst_status_info):
        """组装数据"""
        need_delete_id_list = []
        update_create_error_id_list = []
        for key, val in instance_info.items():
            if key in inst_status_info:
                if val.get("backend_status") in ["BackendError"]:
                    val["backend_status"] = "BackendNormal"
                    update_create_error_id_list.append(val["id"])
                val.update(inst_status_info[key])
            else:
                if val["oper_type"] == app_constants.DELETE_INSTANCE:
                    need_delete_id_list.append(val["id"])
        if update_create_error_id_list:
            update_create_error_record.delay(update_create_error_id_list)
        InstanceConfig.objects.filter(id__in=need_delete_id_list).update(is_deleted=True, deleted_time=datetime.now())

    def inst_count_handler(self, instance_list, app_status):
        ret_data = {"total_num": len(instance_list), "error_num": 0, "instance_list": instance_list}

        for val in instance_list:
            if (val["backend_status"] in app_constants.UNNORMAL_STATUS) or (
                val["status"] in app_constants.UNNORMAL_STATUS
            ):
                if app_status in [2, "2", None]:
                    ret_data["error_num"] += 1
                else:
                    instance_list.remove(val)
        return ret_data

    def get(self, request, project_id, instance_info, category, app_status):
        """针对k8s的实例"""
        cluster_ns_inst = self.get_cluster_ns_inst(instance_info)
        inst_status_info = self.get_k8s_category_info(request, project_id, cluster_ns_inst, category)
        self.compose_data(instance_info, inst_status_info)
        ret_data = self.inst_count_handler(instance_info.values(), app_status)
        return ret_data
