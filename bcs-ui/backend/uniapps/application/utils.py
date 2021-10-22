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
import functools
import json
import time

from django.conf import settings
from django.utils import timezone
from rest_framework.response import Response

from backend.components import paas_cc
from backend.templatesets.legacy_apps.configuration.constants import K8sResourceName
from backend.templatesets.legacy_apps.instance import constants as instance_constants
from backend.templatesets.legacy_apps.instance.models import InstanceConfig
from backend.utils.basic import getitems
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes

from . import constants

STAG_ENV = 2
PROD_ENV = 1


class APIResponse(Response):
    def __init__(self, data, *args, **kwargs):
        data.setdefault('code', 0)
        data.setdefault('message', '')
        return super(APIResponse, self).__init__(data, *args, **kwargs)


def image_handler(image):
    """处理镜像，只展示用户填写的一部分"""
    for env in constants.SPLIT_IMAGE:
        info_split = image.split("/")
        if env in info_split:
            image = "/" + "/".join(info_split[info_split.index(env) :])
            break
    return image


def get_k8s_desired_ready_instance_count(info, resource_name):
    """获取应用期望/正常的实例数量"""
    filter_keys = constants.RESOURCE_REPLICAS_KEYS[resource_name]
    # 针对不同的模板获取不同key对应的值
    ready_replicas = getitems(info, filter_keys['ready_replicas_keys'], default=0)
    desired_replicas = getitems(info, filter_keys['desired_replicas_keys'], default=0)
    return desired_replicas, ready_replicas


def cluster_env(env, ret_num_flag=True):
    """集群环境匹配"""
    all_env = settings.CLUSTER_ENV_FOR_FRONT
    front_env = all_env.get(env)
    if ret_num_flag:
        if front_env == "stag":
            return STAG_ENV
        else:
            return PROD_ENV
    else:
        return front_env


def get_project_namespaces(access_token, project_id):
    ns_resp = paas_cc.get_namespace_list(access_token, project_id, desire_all_data=True)
    if ns_resp.get('code') != ErrorCode.NoError:
        raise error_codes.APIError(ns_resp.get('message'))
    data = ns_resp.get('data') or {}
    return data.get('results') or []


def get_namespace_name_map(access_token, project_id):
    project_ns_info = get_project_namespaces(access_token, project_id)
    return {ns['name']: ns for ns in project_ns_info}


def base64_encode_params(info):
    """base64编码"""
    json_extra = bytes(json.dumps(info), 'utf-8')
    return base64.b64encode(json_extra)


def get_k8s_resource_status(resource_kind, resource, replicas, available):
    """获取资源(deployment/sts/job/ds)运行状态"""
    status = constants.ResourceStatus.Unready.value
    # 期望的数量和可用的数量都为0时，认为也是正常的
    if (available == replicas and available > 0) or (available == replicas == 0):
        status = constants.ResourceStatus.Running.value
    # 针对job添加complete状态的判断
    if resource_kind == constants.REVERSE_CATEGORY_MAP[K8sResourceName.K8sJob.value]:
        # 获取completed的replica的数量
        completed_replicas = getitems(resource, ['data', 'spec', 'completions'], default=0)
        if completed_replicas == replicas and available > 0:
            status = constants.ResourceStatus.Completed.value
    return status


def delete_instance_records(online_instances, local_instances):
    diff_insts = set(local_instances) - set(online_instances.keys())
    instance_id_list = [local_instances[key].get('id') for key in diff_insts]
    InstanceConfig.objects.filter(id__in=instance_id_list).exclude(oper_type=constants.REBUILD_INSTANCE).update(
        is_deleted=True, deleted_time=timezone.now()
    )


def get_instance_version_name(annotations, labels):
    name_key = instance_constants.ANNOTATIONS_VERSION
    return annotations.get(name_key) or labels.get(name_key)


def get_instance_version_id(annotations, labels):
    id_key = instance_constants.ANNOTATIONS_VERSION_ID
    return annotations.get(id_key) or labels.get(id_key)


def get_instance_version(annotations, labels):
    name = get_instance_version_name(annotations, labels)
    id = get_instance_version_id(annotations, labels)
    return {'version': name, 'version_id': id}


def retry_requests(func, params=None, data=None, max_retries=2):
    """查询应用信息
    因为现在通过接口以storage为数据源，因此，为防止接口失败或者接口为空的情况，增加请求次数
    """
    for i in range(1, max_retries + 1):
        try:
            resp = func(params) if params else func(**data)
            if i == max_retries:
                return resp
            # 如果为data为空时，code肯定不为0
            if not resp.get("data"):
                time.sleep(0.5)
                continue
            return resp
        except Exception:
            # 设置等待时间
            time.sleep(0.5)

    raise error_codes.APIError("query storage api error")


def exclude_records(
    cluster_id_from_params: str,
    cluster_id_from_instance: str,
    cluster_type_from_params: str,
    cluster_type_from_instance: str,
) -> bool:
    """判断是否排除记录

    :param cluster_id_from_params: 请求参数中的集群 ID，用以过滤集群下的资源
    :param cluster_id_from_instance: 实例中携带的集群 ID
    :param cluster_type_from_params: 请求参数中的集群环境，包含正式环境和测试环境
    :param cluster_type_from_instance: 实例中的集群环境类型
    :returns: 返回True/False, 其中 True标识可以排除记录
    """
    if not cluster_id_from_instance:
        return True
    if cluster_id_from_params:
        if cluster_id_from_instance != cluster_id_from_params:
            return True
    elif str(cluster_type_from_params) != str(cluster_type_from_instance):
        return True
    return False
