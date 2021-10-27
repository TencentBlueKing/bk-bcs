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

from django.utils.translation import ugettext_lazy as _

from backend.components import bcs, paas_cc
from backend.resources.namespace.utils import get_namespaces
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes


def get_project_namespaces(access_token, project_id):
    namespace_list = get_namespaces(access_token, project_id)
    if not namespace_list:
        raise error_codes.APIError(_("获取项目命名空间信息为空"))
    return namespace_list


def delete_pods(access_token, project_id, pod_names):
    # pod_names: {(cluster_id, namespace, resource_kind): [pod_name]}
    err_msg_list = []
    for key, names in pod_names.items():
        cluster_id, namespace, resource_name, _ = key
        client = bcs.k8s.K8SClient(access_token, project_id, cluster_id, None)
        for name in names:
            resp = client.delete_pod(namespace, name)
            if resp.get("code") != ErrorCode.NoError:
                err_msg_list.append(
                    _("集群:{},命名空间:{},资源名称:{}重新调度失败, {}").format(
                        cluster_id, namespace, resource_name, resp.get("message")
                    )
                )
    if err_msg_list:
        raise error_codes.APIError(";".join(err_msg_list))
