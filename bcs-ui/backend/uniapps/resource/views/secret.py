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
import logging

from django.utils.translation import ugettext_lazy as _
from rest_framework import viewsets

from backend.components.bcs import k8s
from backend.templatesets.legacy_apps.configuration.serializers import K8sSecretCreateOrUpdateSLZ
from backend.templatesets.legacy_apps.instance.constants import K8S_SECRET_SYS_CONFIG
from backend.uniapps import utils as app_utils
from backend.uniapps.application.base_views import BaseAPI
from backend.uniapps.application.utils import APIResponse
from backend.uniapps.resource.constants import DEFAULT_SEARCH_FIELDS
from backend.utils.errcodes import ErrorCode

from .base import ResourceOperate

logger = logging.getLogger(__name__)


class Secrets(viewsets.ViewSet, BaseAPI, ResourceOperate):

    category = 'secret'
    cate = 'K8sSecret'
    sys_config = K8S_SECRET_SYS_CONFIG

    def get_secrets_by_cluster_id(self, request, params, project_id, cluster_id):
        """查询secrets"""
        search_fields = copy.deepcopy(DEFAULT_SEARCH_FIELDS)

        search_fields.append("data.data")
        params.update({"field": ",".join(search_fields)})
        client = k8s.K8SClient(request.user.token.access_token, project_id, cluster_id, env=None)
        resp = client.get_secret(params)

        if resp.get("code") != ErrorCode.NoError:
            logger.error(u"bcs_api error: %s" % resp.get("message", ""))
            return resp.get("code", ErrorCode.UnknownError), resp.get("message", _("请求出现异常!"))
        data = resp.get("data") or []
        return 0, data

    def get(self, request, project_id):
        """ 获取项目下所有的secrets """
        cluster_dicts = self.get_project_cluster_info(request, project_id)
        cluster_data = cluster_dicts.get('results', {}) or {}

        data = []
        params = dict(request.GET.items())
        is_decode = request.GET.get('decode')
        is_decode = True if is_decode == '1' else False
        # get project namespace info
        namespace_dict = app_utils.get_ns_id_map(request.user.token.access_token, project_id)

        for cluster_info in cluster_data:
            cluster_id = cluster_info.get('cluster_id')
            # 当参数中集群ID存在时，判断集群ID匹配成功后，继续后续逻辑
            if params.get('cluster_id') and params['cluster_id'] != cluster_id:
                continue
            cluster_env = cluster_info.get('environment')
            code, cluster_secrets = self.get_secrets_by_cluster_id(request, params, project_id, cluster_id)
            # 单个集群错误时，不抛出异常信息
            if code != ErrorCode.NoError:
                continue
            self.handle_data(
                request,
                cluster_secrets,
                self.cate,
                project_id,
                cluster_id,
                is_decode,
                cluster_env,
                cluster_info.get('name', ''),
                namespace_dict=namespace_dict,
            )
            data += cluster_secrets

        # 按时间倒序排列
        data.sort(key=lambda x: x.get('createTime', ''), reverse=True)
        return APIResponse({"code": ErrorCode.NoError, "data": {"data": data, "length": len(data)}, "message": "ok"})

    def delete_secret(self, request, project_id, cluster_id, namespace, name):
        return self.delete_resource(request, project_id, cluster_id, namespace, name)

    def batch_delete_secrets(self, request, project_id):
        return self.batch_delete_resource(request, project_id)

    def update_secret(self, request, project_id, cluster_id, namespace, name):
        self.slz = K8sSecretCreateOrUpdateSLZ
        return self.update_resource(request, project_id, cluster_id, namespace, name)
